package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

const strategyRunTimeout = 180 * time.Second

var strategyPrepPhases = []struct {
	ID    string
	Delay time.Duration
}{
	{"profile", 1200 * time.Millisecond},
	{"processes", 1400 * time.Millisecond},
	{"data", 1200 * time.Millisecond},
	{"priorities", 1400 * time.Millisecond},
	{"roadmap", 1200 * time.Millisecond},
}

type StrategyService struct {
	cfg      *config.Config
	runs     *repository.ToolRunRepository
	plans    *repository.PlanRepository
	billing  *BillingService
	llm      *LLMService
	llmCfg   *StrategyLLMSettingsService
	usageLog *repository.UsageLogRepository
	streams  *StrategyStreamHub
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
		llm: llm, llmCfg: llmCfg, usageLog: usageLog,
		streams: NewStrategyStreamHub(), logger: logger,
	}
}

func (s *StrategyService) StreamHub() *StrategyStreamHub {
	return s.streams
}

func (s *StrategyService) StartRun(ctx context.Context, userID string, input model.StrategyInput, locale string) (*model.ToolRun, error) {
	normalizeStrategyInput(&input)
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
	if err := s.attachCalculatorContexts(ctx, userID, &input); err != nil {
		return nil, err
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

func normalizeStrategyInput(input *model.StrategyInput) {
	ids := append([]string(nil), input.CalculatorRunIDs...)
	if input.CalculatorRunID != nil && strings.TrimSpace(*input.CalculatorRunID) != "" {
		ids = append(ids, strings.TrimSpace(*input.CalculatorRunID))
	}
	input.CalculatorRunIDs = uniqueStringsLimit(ids, model.MaxStrategyCalculatorRuns)
	input.CalculatorRunID = nil
}

func uniqueStringsLimit(items []string, max int) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
		if len(out) >= max {
			break
		}
	}
	return out
}

func (s *StrategyService) attachCalculatorContexts(ctx context.Context, userID string, input *model.StrategyInput) error {
	if len(input.CalculatorRunIDs) == 0 {
		return nil
	}
	contexts := make([]model.CalculatorRunContext, 0, len(input.CalculatorRunIDs))
	for _, runID := range input.CalculatorRunIDs {
		ctxItem, err := s.loadCalculatorContext(ctx, userID, runID)
		if err != nil {
			return err
		}
		contexts = append(contexts, ctxItem)
	}
	input.CalculatorContexts = contexts
	return nil
}

func (s *StrategyService) loadCalculatorContext(ctx context.Context, userID, runID string) (model.CalculatorRunContext, error) {
	run, err := s.runs.GetByIDForUser(ctx, runID, userID)
	if err != nil {
		return model.CalculatorRunContext{}, repository.ErrNotFound
	}
	if run.ToolSlug != "calculator" || run.Status != model.RunStatusDone {
		return model.CalculatorRunContext{}, ErrInvalidInput
	}
	var calcIn model.CalculatorInput
	var calcOut model.CalculatorOutput
	if err := json.Unmarshal(run.Input, &calcIn); err != nil {
		return model.CalculatorRunContext{}, ErrInvalidInput
	}
	if err := json.Unmarshal(run.Output, &calcOut); err != nil {
		return model.CalculatorRunContext{}, ErrInvalidInput
	}
	return model.CalculatorRunContext{
		RunID:              runID,
		ProcessName:        calcIn.ProcessName,
		NetBenefitMonthly:  calcOut.NetBenefitMonthly,
		PaybackMonths:      calcOut.PaybackMonths,
		ROIHorizonPct:      calcOut.ROIHorizonPct,
		NPV:                calcOut.NPV,
		Capex:              calcIn.Capex,
		MonthlySupport:     calcIn.MonthlySolutionCost,
		Recommendation:     string(calcOut.Recommendation),
		RecommendationText: calcOut.RecommendationText,
	}, nil
}

func (s *StrategyService) processRun(runID string) {
	ctx, cancel := context.WithTimeout(context.Background(), strategyRunTimeout)
	defer cancel()
	defer func() {
		time.AfterFunc(5*time.Minute, func() { s.streams.Cleanup(runID) })
	}()

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

	for _, phase := range strategyPrepPhases {
		select {
		case <-ctx.Done():
			s.failRun(ctx, runID, ctx.Err().Error())
			return
		case <-time.After(phase.Delay):
		}
		s.publishPhase(runID, phase.ID, "active")
		s.publishPhase(runID, phase.ID, "done")
	}

	s.publishPhase(runID, "generating", "active")

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
	llmReq := LLMCompletionRequest{
		Provider:     provider,
		Model:        settings.ActiveModel(),
		SystemPrompt: s.llmCfg.ResolvedSystemPrompt(settings),
		UserPrompt:   string(userPayload),
		Temperature:  settings.Temperature,
		MaxTokens:    settings.MaxTokens,
		APIKey:       apiKey,
		BaseURL:      s.llm.BaseURL(provider),
	}

	result, err := s.streamWithRetry(ctx, runID, llmReq)
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

	s.streams.Publish(runID, StrategyStreamEvent{
		Type: StrategyEventDone,
		Data: strategyEventData(map[string]string{"status": "done"}),
	})
}

func (s *StrategyService) streamWithRetry(ctx context.Context, runID string, req LLMCompletionRequest) (*LLMCompletionResult, error) {
	onDelta := func(delta string) {
		s.streams.Publish(runID, StrategyStreamEvent{
			Type: StrategyEventChunk,
			Data: strategyEventData(map[string]string{"delta": delta}),
		})
	}
	result, err := s.llm.StreamComplete(ctx, req, onDelta)
	if err == nil {
		return result, nil
	}
	if !isRetryableLLMError(err) {
		return nil, err
	}
	return s.llm.StreamComplete(ctx, req, onDelta)
}

func (s *StrategyService) publishPhase(runID, id, status string) {
	s.streams.Publish(runID, StrategyStreamEvent{
		Type: StrategyEventPhase,
		Data: strategyEventData(map[string]string{"id": id, "status": status}),
	})
}

func (s *StrategyService) failRun(ctx context.Context, runID, msg string) {
	if len(msg) > 500 {
		msg = msg[:500]
	}
	_ = s.runs.UpdateRunError(ctx, runID, msg)
	s.streams.Publish(runID, StrategyStreamEvent{
		Type: StrategyEventRunError,
		Data: strategyEventData(map[string]string{"message": msg}),
	})
}

func (s *StrategyService) CanStreamRun(ctx context.Context, runID, userID string) error {
	run, err := s.runs.GetByIDForUser(ctx, runID, userID)
	if err != nil {
		return repository.ErrNotFound
	}
	if run.ToolSlug != "strategy" {
		return ErrInvalidInput
	}
	return nil
}

func validateStrategyInput(in model.StrategyInput) error {
	if strings.TrimSpace(in.CompanyName) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.BusinessDescription) == "" {
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
	validData := map[string]bool{"": true, "none": true, "basic": true, "team": true}
	if !validData[in.DataMaturity] {
		return ErrInvalidInput
	}
	validChange := map[string]bool{"": true, "low": true, "medium": true, "high": true}
	if !validChange[in.ChangeReadiness] {
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
	if len(in.CalculatorRunIDs) > model.MaxStrategyCalculatorRuns {
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
