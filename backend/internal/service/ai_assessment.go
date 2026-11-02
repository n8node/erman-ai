package service

import (
	"context"
	"encoding/json"
	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
	"github.com/erman-ai/erman-ai/internal/repository"
	"math"
	"strings"
	"time"
)

const AIAssessmentToolSlug = "ai-assessment"

type AIAssessmentService struct {
	cfg     *config.Config
	runs    *repository.ToolRunRepository
	plans   *repository.PlanRepository
	billing *BillingService
	llm     *LLMService
	llmCfg  *StrategyLLMSettingsService
	usage   *repository.UsageLogRepository
}

func NewAIAssessmentService(cfg *config.Config, runs *repository.ToolRunRepository, plans *repository.PlanRepository, billing *BillingService, llm *LLMService, llmCfg *StrategyLLMSettingsService, usage *repository.UsageLogRepository) *AIAssessmentService {
	return &AIAssessmentService{cfg: cfg, runs: runs, plans: plans, billing: billing, llm: llm, llmCfg: llmCfg, usage: usage}
}

func (s *AIAssessmentService) StartRun(ctx context.Context, userID string, in model.AIAssessmentInput, locale string) (*model.ToolRun, error) {
	if strings.TrimSpace(in.Description) == "" {
		return nil, ErrInvalidInput
	}
	if in.Locale == "" {
		in.Locale = locale
	}
	if err := s.billing.CheckToolLimit(ctx, userID, AIAssessmentToolSlug); err != nil {
		return nil, err
	}
	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(in)
	run, err := s.runs.CreatePending(ctx, userID, AIAssessmentToolSlug, up.PlanSlug, raw)
	if err != nil {
		return nil, err
	}
	if err = s.runs.IncrementUsageCounter(ctx, userID, AIAssessmentToolSlug); err != nil {
		return nil, err
	}
	go s.processRun(run.ID)
	return run, nil
}

func (s *AIAssessmentService) processRun(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	run, err := s.runs.GetByID(ctx, id)
	if err != nil {
		return
	}
	_ = s.runs.UpdateStatus(ctx, id, model.RunStatusProcessing)
	var in model.AIAssessmentInput
	if json.Unmarshal(run.Input, &in) != nil {
		_ = s.runs.UpdateRunError(ctx, id, "invalid input")
		return
	}
	stored, err := s.llmCfg.GetStored(ctx)
	if err != nil {
		_ = s.runs.UpdateRunError(ctx, id, "llm settings unavailable")
		return
	}
	settings := stored.Config.StrategyLLMSettings
	provider := settings.Provider
	creds := s.llm.CredentialsFromStored(stored.Config)
	key := s.llm.ResolveKey(provider, creds)
	if key == "" || (provider == model.LLMProviderYandex && creds.YandexFolderID == "") {
		_ = s.runs.UpdateRunError(ctx, id, "active LLM provider is not configured")
		return
	}
	payload, _ := json.Marshal(in)
	req := LLMCompletionRequest{Provider: provider, Model: settings.ActiveModel(), SystemPrompt: prompts.DefaultAIAssessmentSystemPrompt, UserPrompt: string(payload), Temperature: settings.Temperature, MaxTokens: settings.MaxTokens, APIKey: key, BaseURL: s.llm.BaseURL(provider), FolderID: creds.YandexFolderID}
	res, err := s.llm.Complete(ctx, req)
	if err != nil {
		_ = s.runs.UpdateRunError(ctx, id, err.Error())
		return
	}
	var out model.AIAssessmentOutput
	if err = json.Unmarshal([]byte(cleanAssessmentJSON(res.Content)), &out); err != nil {
		_ = s.runs.UpdateRunError(ctx, id, "invalid AI JSON response")
		return
	}
	out = normalizeAssessment(out, in)
	raw, _ := json.Marshal(out)
	if err = s.runs.UpdateRunDone(ctx, id, raw, int64(res.TotalTokens), res.Model); err != nil {
		return
	}
	usd, rub := s.llmCfg.UsageCosts(stored.Config, provider, res.Model, res.PromptTokens, res.CompletionTokens)
	_ = s.usage.Create(ctx, run.UserID, id, string(provider), res.Model, res.PromptTokens, res.CompletionTokens, usd, rub)
}

func cleanAssessmentJSON(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "```json")
	v = strings.TrimPrefix(v, "```")
	v = strings.TrimSuffix(v, "```")
	return strings.TrimSpace(v)
}
func normalizeAssessment(out model.AIAssessmentOutput, in model.AIAssessmentInput) model.AIAssessmentOutput {
	auto := in.AutomationPct
	if auto <= 0 {
		auto = 60
	}
	if auto > 100 {
		auto = 100
	}
	out.Financial = CalculateAssessmentFinancial(in, auto)
	out.Scenarios = []model.AssessmentScenario{}
	for _, p := range []struct {
		name string
		pct  float64
	}{{"Консервативный", 40}, {"Базовый", auto}, {"Оптимистичный", 85}} {
		f := CalculateAssessmentFinancial(in, p.pct)
		out.Scenarios = append(out.Scenarios, model.AssessmentScenario{Name: p.name, AutomationPct: p.pct, NetBenefitMonthly: f.MonthlyBenefit, PaybackMonths: f.PaybackMonths, ROIPct: f.ROIPct})
	}
	score := 0.0
	weight := 0.0
	critical := false
	for _, c := range out.Criteria {
		score += float64(c.Score) * c.Weight
		weight += c.Weight
		if c.Name != "" && c.Score <= 1 && strings.Contains(strings.ToLower(c.Name), "ошиб") {
			critical = true
		}
	}
	if weight > 0 {
		score = score / weight * 20
	}
	out.ReadinessScore = math.Round(score*10) / 10
	out.Autonomy, out.AutonomyText = assessmentAutonomy(out.ReadinessScore, critical)
	if out.RiskLevel == "" {
		out.RiskLevel = "medium"
	}
	if out.ProcessName == "" {
		out.ProcessName = "Анализируемый процесс"
	}
	return out
}
