package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/i18n"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
	"github.com/erman-ai/erman-ai/internal/repository"
)

const legalScanRunTimeout = 120 * time.Second

type LegalScanService struct {
	cfg      *config.Config
	runs     *repository.ToolRunRepository
	plans    *repository.PlanRepository
	risks    *repository.LegalRiskRepository
	billing  *BillingService
	llm      *LLMService
	llmCfg   *LegalScanLLMSettingsService
	strategy *StrategyLLMSettingsService
	usageLog *repository.UsageLogRepository
	logger   *slog.Logger
}

func NewLegalScanService(
	cfg *config.Config,
	runs *repository.ToolRunRepository,
	plans *repository.PlanRepository,
	risks *repository.LegalRiskRepository,
	billing *BillingService,
	llm *LLMService,
	llmCfg *LegalScanLLMSettingsService,
	strategy *StrategyLLMSettingsService,
	usageLog *repository.UsageLogRepository,
	logger *slog.Logger,
) *LegalScanService {
	return &LegalScanService{
		cfg: cfg, runs: runs, plans: plans, risks: risks, billing: billing,
		llm: llm, llmCfg: llmCfg, strategy: strategy, usageLog: usageLog, logger: logger,
	}
}

func (s *LegalScanService) StartRun(ctx context.Context, userID string, input model.LegalScanInput, locale string) (*model.ToolRun, error) {
	normalizeLegalScanInput(&input)
	if err := validateLegalScanInput(input); err != nil {
		return nil, err
	}
	if err := s.billing.CheckToolLimit(ctx, userID, "legal-scan"); err != nil {
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

	run, err := s.runs.CreatePending(ctx, userID, "legal-scan", up.PlanSlug, inJSON)
	if err != nil {
		return nil, err
	}

	if err := s.runs.IncrementUsageCounter(ctx, userID, "legal-scan"); err != nil {
		return nil, err
	}

	go s.processRun(run.ID)

	return run, nil
}

func (s *LegalScanService) processRun(runID string) {
	run, err := s.runs.GetByID(context.Background(), runID)
	if err != nil {
		return
	}
	var input model.LegalScanInput
	if err := json.Unmarshal(run.Input, &input); err != nil {
		_ = s.runs.UpdateRunError(context.Background(), runID, "invalid input data")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), legalScanRunTimeout)
	defer cancel()

	if err := s.runs.UpdateStatus(ctx, runID, model.RunStatusProcessing); err != nil {
		s.logger.Error("legal scan status update failed", "run_id", runID, "error", err)
		return
	}

	page, err := fetchLegalScanSite(ctx, input.URL)
	if err != nil {
		s.failRun(ctx, runID, err.Error())
		return
	}

	layer1 := runLegalScanLayer1(page, input.SiteFeatures)

	partial := model.LegalScanOutput{Layer1: layer1}
	partialJSON, _ := json.Marshal(partial)
	if err := s.runs.UpdateRunProcessingOutput(ctx, runID, partialJSON); err != nil {
		s.logger.Error("legal scan partial output failed", "run_id", runID, "error", err)
	}

	allRisks, err := s.risks.ListActive(ctx)
	if err != nil {
		s.failRun(ctx, runID, "risk table unavailable")
		return
	}
	matched := matchLegalRisks(allRisks, layer1.Findings, input.SiteFeatures)

	llmStored, err := s.llmCfg.GetStored(ctx)
	if err != nil {
		s.failRun(ctx, runID, "llm settings unavailable")
		return
	}
	strategyStored, err := s.strategy.GetStored(ctx)
	if err != nil {
		s.failRun(ctx, runID, "llm keys unavailable")
		return
	}

	settings := llmStored.Config.LegalScanLLMSettings
	provider := settings.Provider
	apiKey := s.llm.ResolveKey(provider, strategyStored.Config.OpenRouterAPIKey, strategyStored.Config.DeepSeekAPIKey)
	if apiKey == "" {
		s.failRun(ctx, runID, "llm api key not configured")
		return
	}

	findingsJSON, _ := json.Marshal(layer1.Findings)
	riskPayload := buildRiskTablePayload(matched)
	riskJSON, _ := json.Marshal(riskPayload)

	locale := i18n.NormalizeLocale(input.Locale)
	systemPrompt := settings.SystemPrompt
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = prompts.LegalScanSystemPrompt(locale)
	}

	maxTokens := settings.MaxTokens
	if maxTokens > 8192 {
		maxTokens = 8192
	}

	req := LLMCompletionRequest{
		Provider:     provider,
		Model:        settings.ActiveModel(),
		SystemPrompt: systemPrompt,
		UserPrompt: prompts.LegalScanUserPrompt(
			input.Industry,
			input.SiteFeatures.TrafficFromAds,
			string(findingsJSON),
			string(riskJSON),
		),
		Temperature: settings.Temperature,
		MaxTokens:   maxTokens,
		APIKey:      apiKey,
		BaseURL:     s.llm.BaseURL(provider),
	}

	var llmPart model.LegalScanOutput
	if len(matched) == 0 {
		llmPart = model.LegalScanOutput{
			Summary: model.LegalScanSummary{},
			Risks:   []model.LegalScanRiskItem{},
			Disclaimer: legalScanDisclaimerRU,
		}
	} else {
		result, err := s.llm.Complete(ctx, req)
		if err != nil && isRetryableLLMError(err) {
			result, err = s.llm.Complete(ctx, req)
		}
		if err != nil {
			s.failRun(ctx, runID, err.Error())
			return
		}

		raw, err := parseLegalScanLLMOutput(result.Content)
		if err != nil {
			s.logger.Warn("legal scan llm parse failed, using fallback", "run_id", runID, "error", err)
			raw = &model.LegalScanLLMOutput{}
		}
		validated := validateLegalScanOutput(raw, matched)
		llmPart = validated

		_ = s.usageLog.Create(ctx, run.UserID, runID, result.Model, result.PromptTokens, result.CompletionTokens, 0)
	}

	final := model.LegalScanOutput{
		Layer1:       layer1,
		Summary:      llmPart.Summary,
		Risks:        llmPart.Risks,
		IndustryNote: llmPart.IndustryNote,
		Disclaimer:   llmPart.Disclaimer,
	}
	if final.Summary.RisksCount == 0 {
		final.Summary.RisksCount = len(final.Risks)
	}

	outJSON, err := json.Marshal(final)
	if err != nil {
		s.failRun(ctx, runID, "failed to save output")
		return
	}

	modelUsed := settings.ActiveModel()
	if err := s.runs.UpdateRunDone(ctx, runID, outJSON, 0, modelUsed); err != nil {
		s.logger.Error("legal scan save failed", "run_id", runID, "error", err)
		s.failRun(context.Background(), runID, "failed to save output")
	}
}

func (s *LegalScanService) failRun(ctx context.Context, runID, msg string) {
	if len(msg) > 500 {
		msg = msg[:500]
	}
	_ = s.runs.UpdateRunError(ctx, runID, msg)
}

func normalizeLegalScanInput(in *model.LegalScanInput) {
	in.URL = strings.TrimSpace(in.URL)
	if in.URL != "" && !strings.HasPrefix(in.URL, "http://") && !strings.HasPrefix(in.URL, "https://") {
		in.URL = "https://" + in.URL
	}
	in.Industry = strings.TrimSpace(in.Industry)
	in.CompanySize = strings.TrimSpace(in.CompanySize)
}

func validateLegalScanInput(in model.LegalScanInput) error {
	if in.URL == "" {
		return fmt.Errorf("%w: url required", ErrInvalidInput)
	}
	u, err := url.Parse(in.URL)
	if err != nil || u.Host == "" {
		return fmt.Errorf("%w: invalid url", ErrInvalidInput)
	}
	if in.Industry == "" {
		return fmt.Errorf("%w: industry required", ErrInvalidInput)
	}
	return nil
}

var ErrLegalRiskNotFound = errors.New("legal risk not found")
