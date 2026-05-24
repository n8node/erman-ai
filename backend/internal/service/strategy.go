package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

type StrategyService struct {
	cfg      *config.Config
	runs     *repository.ToolRunRepository
	plans    *repository.PlanRepository
	billing  *BillingService
	llm      *LLMService
	llmCfg   *StrategyLLMSettingsService
	usageLog *repository.UsageLogRepository
	logger   *slog.Logger
}

func NewStrategyService(
	cfg *config.Config,
	runs *repository.ToolRunRepository,
	plans *repository.PlanRepository,
	billing *BillingService,
	llm *LLMService,
	llmCfg *StrategyLLMSettingsService,
	usageLog *repository.UsageLogRepository,
	logger *slog.Logger,
) *StrategyService {
	return &StrategyService{
		cfg: cfg, runs: runs, plans: plans, billing: billing,
		llm: llm, llmCfg: llmCfg, usageLog: usageLog, logger: logger,
	}
}

func (s *StrategyService) StartRun(ctx context.Context, userID string, input model.StrategyInput, locale string) (*model.ToolRun, error) {
	if err := validateStrategyInput(input); err != nil {
		return nil, err
	}
	if err := s.billing.CheckToolLimit(ctx, userID, "strategy"); err != nil {
		return nil, err
	}

	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, err
	}

	input.Locale = locale
	if input.CalculatorRunID != nil && *input.CalculatorRunID != "" {
		if err := s.attachCalculatorContext(ctx, userID, &input, *input.CalculatorRunID); err != nil {
			return nil, err
		}
	}

	inJSON, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	run, err := s.runs.CreatePending(ctx, userID, "strategy", up.PlanSlug, inJSON)
	if err != nil {
		return nil, err
	}

	if err := s.runs.IncrementUsageCounter(ctx, userID, "strategy"); err != nil {
		return nil, err
	}

	go s.processRun(run.ID)

	return run, nil
}

func (s *StrategyService) attachCalculatorContext(ctx context.Context, userID string, input *model.StrategyInput, runID string) error {
	run, err := s.runs.GetByIDForUser(ctx, runID, userID)
	if err != nil {
		return repository.ErrNotFound
	}
	if run.ToolSlug != "calculator" || run.Status != model.RunStatusDone {
		return ErrInvalidInput
	}
	var calcIn model.CalculatorInput
	var calcOut model.CalculatorOutput
	if err := json.Unmarshal(run.Input, &calcIn); err != nil {
		return ErrInvalidInput
	}
	if err := json.Unmarshal(run.Output, &calcOut); err != nil {
		return ErrInvalidInput
	}
	if input.PainPoints == "" {
		input.PainPoints = fmt.Sprintf(
			"Calculator context — process: %s; net benefit: %.0f RUB/mo; payback: %.1f months; recommendation: %s",
			calcIn.ProcessName, calcOut.NetBenefitMonthly, calcOut.PaybackMonths, calcOut.Recommendation,
		)
	}
	return nil
}

func (s *StrategyService) processRun(runID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := s.runs.UpdateStatus(ctx, runID, model.RunStatusProcessing); err != nil {
		s.logger.Error("strategy run status update failed", "run_id", runID, "error", err)
		return
	}

	run, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return
	}

	var input model.StrategyInput
	if err := json.Unmarshal(run.Input, &input); err != nil {
		s.failRun(ctx, runID, "invalid input data")
		return
	}

	stored, err := s.llmCfg.GetStored(ctx)
	if err != nil {
		s.failRun(ctx, runID, "llm settings unavailable")
		return
	}
	settings := stored.Config.StrategyLLMSettings
	provider := settings.Provider
	apiKey := s.llm.ResolveKey(provider, stored.Config.OpenRouterAPIKey, stored.Config.DeepSeekAPIKey)
	if apiKey == "" {
		s.failRun(ctx, runID, "llm api key not configured")
		return
	}

	userPayload, _ := json.Marshal(input)
	result, err := s.completeWithRetry(ctx, LLMCompletionRequest{
		Provider:     provider,
		Model:        settings.ActiveModel(),
		SystemPrompt: s.llmCfg.ResolvedSystemPrompt(settings),
		UserPrompt:   string(userPayload),
		Temperature:  settings.Temperature,
		MaxTokens:    settings.MaxTokens,
		APIKey:       apiKey,
		BaseURL:      s.llm.BaseURL(provider),
	})
	if err != nil {
		s.failRun(ctx, runID, err.Error())
		return
	}

	output, err := parseStrategyOutput(result.Content)
	if err != nil {
		s.failRun(ctx, runID, "invalid strategy json: "+err.Error())
		return
	}

	outJSON, err := json.Marshal(output)
	if err != nil {
		s.failRun(ctx, runID, "failed to save output")
		return
	}

	if err := s.runs.UpdateRunDone(ctx, runID, outJSON, int64(result.TotalTokens), result.Model); err != nil {
		s.logger.Error("strategy run save failed", "run_id", runID, "error", err)
		return
	}

	_ = s.usageLog.Create(ctx, run.UserID, runID, result.Model, result.PromptTokens, result.CompletionTokens, 0)
}

func (s *StrategyService) failRun(ctx context.Context, runID, msg string) {
	if len(msg) > 500 {
		msg = msg[:500]
	}
	_ = s.runs.UpdateRunError(ctx, runID, msg)
}

func (s *StrategyService) completeWithRetry(ctx context.Context, req LLMCompletionRequest) (*LLMCompletionResult, error) {
	result, err := s.llm.Complete(ctx, req)
	if err == nil {
		return result, nil
	}
	if !isRetryableLLMError(err) {
		return nil, err
	}
	return s.llm.Complete(ctx, req)
}

func isRetryableLLMError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof")
}

func validateStrategyInput(in model.StrategyInput) error {
	if strings.TrimSpace(in.CompanyName) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.Industry) == "" {
		return ErrInvalidInput
	}
	validSize := map[string]bool{"1-10": true, "11-50": true, "51-200": true, "201-1000": true, "1000+": true}
	if !validSize[in.CompanySize] {
		return ErrInvalidInput
	}
	validLevel := map[string]bool{"none": true, "exploring": true, "piloting": true, "scaling": true}
	if !validLevel[in.CurrentAILevel] {
		return ErrInvalidInput
	}
	if len(in.MainGoals) == 0 {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.PainPoints) == "" {
		return ErrInvalidInput
	}
	validTimeline := map[string]bool{"3months": true, "6months": true, "1year": true, "2years": true}
	if !validTimeline[in.Timeline] {
		return ErrInvalidInput
	}
	return nil
}

func parseStrategyOutput(content string) (*model.StrategyOutput, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var out model.StrategyOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.ExecutiveSummary) == "" {
		return nil, errors.New("executive_summary required")
	}
	if len(out.RecommendedSolutions) == 0 {
		return nil, errors.New("recommended_solutions required")
	}
	return &out, nil
}

func ParseStrategyRun(run *model.ToolRun) (model.StrategyInput, model.StrategyOutput, error) {
	var in model.StrategyInput
	var out model.StrategyOutput
	if err := json.Unmarshal(run.Input, &in); err != nil {
		return in, out, err
	}
	if len(run.Output) > 0 {
		if err := json.Unmarshal(run.Output, &out); err != nil {
			return in, out, err
		}
	}
	return in, out, nil
}
