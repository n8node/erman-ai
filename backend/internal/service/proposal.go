package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/i18n"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
	"github.com/erman-ai/erman-ai/internal/repository"
)

const proposalRunTimeout = 120 * time.Second

type ProposalService struct {
	cfg      *config.Config
	runs     *repository.ToolRunRepository
	plans    *repository.PlanRepository
	billing  *BillingService
	llm      *LLMService
	llmCfg   *StrategyLLMSettingsService
	usageLog *repository.UsageLogRepository
	logger   *slog.Logger
}

func NewProposalService(
	cfg *config.Config,
	runs *repository.ToolRunRepository,
	plans *repository.PlanRepository,
	billing *BillingService,
	llm *LLMService,
	llmCfg *StrategyLLMSettingsService,
	usageLog *repository.UsageLogRepository,
	logger *slog.Logger,
) *ProposalService {
	return &ProposalService{
		cfg: cfg, runs: runs, plans: plans, billing: billing,
		llm: llm, llmCfg: llmCfg, usageLog: usageLog, logger: logger,
	}
}

func (s *ProposalService) StartRun(ctx context.Context, userID string, input model.ProposalInput, locale string) (*model.ToolRun, error) {
	normalizeProposalInput(&input)
	if err := validateProposalInput(input); err != nil {
		return nil, err
	}
	if err := s.billing.CheckToolLimit(ctx, userID, "proposal"); err != nil {
		return nil, err
	}

	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, err
	}

	input.Locale = locale
	if err := s.attachCalculatorContext(ctx, userID, &input); err != nil {
		return nil, err
	}

	inJSON, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	run, err := s.runs.CreatePending(ctx, userID, "proposal", up.PlanSlug, inJSON)
	if err != nil {
		return nil, err
	}

	if err := s.runs.IncrementUsageCounter(ctx, userID, "proposal"); err != nil {
		return nil, err
	}

	go s.processRun(run.ID)

	return run, nil
}

func normalizeProposalInput(input *model.ProposalInput) {
	input.ProposalScenario = model.NormalizeProposalScenario(input.ProposalScenario)
	input.PriorContactSummary = strings.TrimSpace(input.PriorContactSummary)
	input.ProblemSource = strings.TrimSpace(input.ProblemSource)
	input.ClientCompany = strings.TrimSpace(input.ClientCompany)
	input.ClientContact = strings.TrimSpace(input.ClientContact)
	input.ClientIndustry = strings.TrimSpace(input.ClientIndustry)
	input.ClientProblem = strings.TrimSpace(input.ClientProblem)
	input.SolutionName = strings.TrimSpace(input.SolutionName)
	input.SolutionDescription = strings.TrimSpace(input.SolutionDescription)
	input.PaymentSchedule = strings.TrimSpace(input.PaymentSchedule)
	input.SenderCompany = strings.TrimSpace(input.SenderCompany)
	input.SenderContact = strings.TrimSpace(input.SenderContact)
	input.SenderPhone = strings.TrimSpace(input.SenderPhone)
	input.SenderEmail = strings.TrimSpace(input.SenderEmail)

	clean := make([]string, 0, len(input.Deliverables))
	for _, d := range input.Deliverables {
		d = strings.TrimSpace(d)
		if d != "" {
			clean = append(clean, d)
		}
	}
	input.Deliverables = clean

	if input.CalculatorRunID != nil {
		id := strings.TrimSpace(*input.CalculatorRunID)
		if id == "" {
			input.CalculatorRunID = nil
		} else {
			input.CalculatorRunID = &id
		}
	}
}

func (s *ProposalService) attachCalculatorContext(ctx context.Context, userID string, input *model.ProposalInput) error {
	if input.CalculatorRunID == nil {
		return nil
	}
	run, err := s.runs.GetByIDForUser(ctx, *input.CalculatorRunID, userID)
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
	ctxItem := model.CalculatorRunContext{
		RunID:              *input.CalculatorRunID,
		ProcessName:        calcIn.ProcessName,
		NetBenefitMonthly:  calcOut.NetBenefitMonthly,
		PaybackMonths:      calcOut.PaybackMonths,
		ROIHorizonPct:      calcOut.ROIHorizonPct,
		NPV:                calcOut.NPV,
		Capex:              calcIn.Capex,
		MonthlySupport:     calcIn.MonthlySolutionCost,
		Recommendation:     string(calcOut.Recommendation),
		RecommendationText: calcOut.RecommendationText,
	}
	input.CalculatorContext = &ctxItem
	return nil
}

func (s *ProposalService) processRun(runID string) {
	run, err := s.runs.GetByID(context.Background(), runID)
	if err != nil {
		return
	}
	var input model.ProposalInput
	if err := json.Unmarshal(run.Input, &input); err != nil {
		_ = s.runs.UpdateRunError(context.Background(), runID, "invalid input data")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), proposalRunTimeout)
	defer cancel()

	if err := s.runs.UpdateStatus(ctx, runID, model.RunStatusProcessing); err != nil {
		s.logger.Error("proposal run status update failed", "run_id", runID, "error", err)
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

	userPayload, err := json.Marshal(input)
	if err != nil {
		s.failRun(ctx, runID, "failed to build prompt")
		return
	}

	locale := i18n.NormalizeLocale(input.Locale)

	maxTokens := settings.MaxTokens
	if maxTokens > 16000 {
		maxTokens = 16000
	}

	basePrompt := settings.ProposalSystemPrompt
	if basePrompt == "" {
		basePrompt = prompts.DefaultProposalSystemPrompt
	}

	req := LLMCompletionRequest{
		Provider:     provider,
		Model:        settings.ActiveModel(),
		SystemPrompt: prompts.BuildProposalSystemPrompt(basePrompt, locale, input.ProposalScenario, input.IncludePricing),
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

	output, err := parseProposalOutput(result.Content)
	if err != nil {
		s.failRun(ctx, runID, "invalid proposal output from model")
		return
	}

	outJSON, err := json.Marshal(output)
	if err != nil {
		s.failRun(ctx, runID, "failed to save output")
		return
	}

	if err := s.runs.UpdateRunDone(ctx, runID, outJSON, int64(result.TotalTokens), result.Model); err != nil {
		s.logger.Error("proposal run save failed", "run_id", runID, "error", err)
		s.failRun(context.Background(), runID, "failed to save proposal output")
		return
	}

	costUSD, costRUB := s.llmCfg.UsageCosts(stored.Config, provider, result.Model, result.PromptTokens, result.CompletionTokens)
	_ = s.usageLog.Create(ctx, run.UserID, runID, string(provider), result.Model, result.PromptTokens, result.CompletionTokens, costUSD, costRUB)
}

func (s *ProposalService) failRun(ctx context.Context, runID, msg string) {
	if len(msg) > 500 {
		msg = msg[:500]
	}
	_ = s.runs.UpdateRunError(ctx, runID, msg)
}

func (s *ProposalService) GetRunForUser(ctx context.Context, runID, userID string) (*model.ToolRun, error) {
	run, err := s.runs.GetByIDForUser(ctx, runID, userID)
	if err != nil {
		return nil, err
	}
	if run.ToolSlug != "proposal" {
		return nil, ErrInvalidInput
	}
	return run, nil
}
