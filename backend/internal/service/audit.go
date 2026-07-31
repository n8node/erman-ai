package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/i18n"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
	"github.com/erman-ai/erman-ai/internal/repository"
)

const auditRunTimeout = 180 * time.Second

type AuditService struct {
	cfg      *config.Config
	runs     *repository.ToolRunRepository
	plans    *repository.PlanRepository
	billing  *BillingService
	llm      *LLMService
	llmCfg   *StrategyLLMSettingsService
	usageLog *repository.UsageLogRepository
	logger   *slog.Logger
}

func NewAuditService(
	cfg *config.Config,
	runs *repository.ToolRunRepository,
	plans *repository.PlanRepository,
	billing *BillingService,
	llm *LLMService,
	llmCfg *StrategyLLMSettingsService,
	usageLog *repository.UsageLogRepository,
	logger *slog.Logger,
) *AuditService {
	return &AuditService{
		cfg: cfg, runs: runs, plans: plans, billing: billing,
		llm: llm, llmCfg: llmCfg, usageLog: usageLog, logger: logger,
	}
}

type auditLLMPayload struct {
	model.AuditInput
	ProcessScoresPrecomputed []model.AuditProcessScore `json:"process_scores_precomputed"`
	TotalMonthlyCostRub      float64                   `json:"total_monthly_cost_rub"`
	TotalMonthlySavingsRub   float64                   `json:"total_monthly_savings_est_rub"`
}

func (s *AuditService) StartRun(ctx context.Context, userID string, input model.AuditInput, locale string) (*model.ToolRun, error) {
	normalizeAuditInput(&input)
	if err := validateAuditInput(input); err != nil {
		return nil, err
	}
	if err := s.billing.CheckToolLimit(ctx, userID, "audit"); err != nil {
		return nil, err
	}

	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, err
	}

	input.Locale = locale
	inJSON, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	run, err := s.runs.CreatePending(ctx, userID, "audit", up.PlanSlug, inJSON)
	if err != nil {
		return nil, err
	}

	if err := s.runs.IncrementUsageCounter(ctx, userID, "audit"); err != nil {
		return nil, err
	}

	go s.processRun(run.ID)

	return run, nil
}

func (s *AuditService) processRun(runID string) {
	run, err := s.runs.GetByID(context.Background(), runID)
	if err != nil {
		return
	}
	var input model.AuditInput
	if err := json.Unmarshal(run.Input, &input); err != nil {
		_ = s.runs.UpdateRunError(context.Background(), runID, "invalid input data")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), auditRunTimeout)
	defer cancel()

	if err := s.runs.UpdateStatus(ctx, runID, model.RunStatusProcessing); err != nil {
		s.logger.Error("audit run status update failed", "run_id", runID, "error", err)
		return
	}

	stored, err := s.llmCfg.GetStored(ctx)
	if err != nil {
		s.failRun(ctx, runID, "llm settings unavailable")
		return
	}
	settings := stored.Config.StrategyLLMSettings
	provider := settings.Provider
	creds := s.llm.CredentialsFromStored(stored.Config)
	apiKey := s.llm.ResolveKey(provider, creds)
	if apiKey == "" {
		s.failRun(ctx, runID, "llm api key not configured")
		return
	}
	if provider == model.LLMProviderYandex && creds.YandexFolderID == "" {
		s.failRun(ctx, runID, "yandex folder id not configured")
		return
	}

	scores := computeAuditProcessScores(input.Processes)
	totalCost, totalSavings := auditTotals(scores)

	payload := auditLLMPayload{
		AuditInput:               input,
		ProcessScoresPrecomputed: scores,
		TotalMonthlyCostRub:      totalCost,
		TotalMonthlySavingsRub:   totalSavings,
	}
	userPayload, err := json.Marshal(payload)
	if err != nil {
		s.failRun(ctx, runID, "failed to build prompt")
		return
	}

	locale := i18n.NormalizeLocale(input.Locale)
	maxTokens := settings.MaxTokens
	if maxTokens > 16000 {
		maxTokens = 16000
	}

	req := LLMCompletionRequest{
		Provider:     provider,
		Model:        settings.ActiveModel(),
		SystemPrompt: prompts.AuditSystemPrompt(locale),
		UserPrompt:   string(userPayload),
		Temperature:  settings.Temperature,
		MaxTokens:    maxTokens,
		APIKey:       apiKey,
		BaseURL:      s.llm.BaseURL(provider),
		FolderID:     creds.YandexFolderID,
		Proxy:        settings.ProxyForProvider(provider),
	}

	result, err := s.llm.Complete(ctx, req)
	if err != nil && isRetryableLLMError(err) {
		result, err = s.llm.Complete(ctx, req)
	}
	if err != nil {
		s.failRun(ctx, runID, err.Error())
		return
	}

	output, err := parseAuditOutput(result.Content)
	if err != nil {
		s.failRun(ctx, runID, "invalid audit output from model")
		return
	}

	if len(output.ProcessScores) == 0 {
		output.ProcessScores = scores
	}
	if output.TotalMonthlyCost == 0 {
		output.TotalMonthlyCost = totalCost
	}
	if output.TotalMonthlySavings == 0 {
		output.TotalMonthlySavings = totalSavings
	}

	outJSON, err := json.Marshal(output)
	if err != nil {
		s.failRun(ctx, runID, "failed to save output")
		return
	}

	if err := s.runs.UpdateRunDone(ctx, runID, outJSON, int64(result.TotalTokens), result.Model); err != nil {
		s.logger.Error("audit run save failed", "run_id", runID, "error", err)
		s.failRun(context.Background(), runID, "failed to save audit output")
		return
	}

	costUSD, costRUB := s.llmCfg.UsageCosts(stored.Config, provider, result.Model, result.PromptTokens, result.CompletionTokens)
	_ = s.usageLog.Create(ctx, run.UserID, runID, string(provider), result.Model, result.PromptTokens, result.CompletionTokens, costUSD, costRUB)
}

func (s *AuditService) failRun(ctx context.Context, runID, msg string) {
	if len(msg) > 500 {
		msg = msg[:500]
	}
	_ = s.runs.UpdateRunError(ctx, runID, msg)
}

func (s *AuditService) GetRunForUser(ctx context.Context, runID, userID string) (*model.ToolRun, error) {
	run, err := s.runs.GetByIDForUser(ctx, runID, userID)
	if err != nil {
		return nil, err
	}
	if run.ToolSlug != "audit" {
		return nil, ErrInvalidInput
	}
	return run, nil
}
