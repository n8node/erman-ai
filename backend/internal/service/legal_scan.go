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
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

const legalScanRunTimeout = 180 * time.Second

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

	normalizeLegalScanCrawlConfig(&input.Crawl, up.Plan)
	if !input.Crawl.SameHostOnly {
		input.Crawl.SameHostOnly = true
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

	ctx, cancel := context.WithTimeout(context.Background(), legalScanRunTimeoutFor(input.Crawl.MaxPages))
	defer cancel()

	if err := s.runs.UpdateStatus(ctx, runID, model.RunStatusProcessing); err != nil {
		s.logger.Error("legal scan status update failed", "run_id", runID, "error", err)
		return
	}

	onProgress := func(meta model.LegalScanCrawlMeta, layer1 model.LegalScanLayer1) {
		partial := model.LegalScanOutput{Layer1: layer1}
		partialJSON, _ := json.Marshal(partial)
		if err := s.runs.UpdateRunProcessingOutput(ctx, runID, partialJSON); err != nil {
			s.logger.Error("legal scan partial output failed", "run_id", runID, "error", err)
		}
	}

	crawlResult, err := crawlLegalScanSite(ctx, input.URL, input.SiteFeatures, legalScanCrawlOptions{
		MaxPages:     input.Crawl.MaxPages,
		MaxDepth:     input.Crawl.MaxDepth,
		SameHostOnly: input.Crawl.SameHostOnly,
		OnProgress:   onProgress,
	})
	if err != nil {
		s.failRun(ctx, runID, err.Error())
		return
	}

	layer1 := runLegalScanLayer1FromPages(
		crawlResult.Pages,
		input.URL,
		crawlResult.HTTPS,
		input.SiteFeatures,
		&crawlResult.Meta,
	)

	layer1 = s.enrichLayer1WithLLM(ctx, run, crawlResult.Pages, layer1, input)

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

	reportPart := buildDeterministicLegalScanReport(layer1, matched, input.Industry)

	final := model.LegalScanOutput{
		Layer1:       layer1,
		Summary:      reportPart.Summary,
		Risks:        reportPart.Risks,
		IndustryNote: reportPart.IndustryNote,
		Disclaimer:   reportPart.Disclaimer,
	}

	outJSON, err := json.Marshal(final)
	if err != nil {
		s.failRun(ctx, runID, "failed to save output")
		return
	}

	if err := s.runs.UpdateRunDone(ctx, runID, outJSON, 0, ""); err != nil {
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

func (s *LegalScanService) GetRunForUser(ctx context.Context, runID, userID string) (*model.ToolRun, error) {
	return s.runs.GetByIDForUser(ctx, runID, userID)
}

func ParseLegalScanRun(run *model.ToolRun) (model.LegalScanInput, model.LegalScanOutput, error) {
	var in model.LegalScanInput
	var out model.LegalScanOutput
	if err := json.Unmarshal(run.Input, &in); err != nil {
		return in, out, err
	}
	if err := json.Unmarshal(run.Output, &out); err != nil {
		return in, out, err
	}
	return in, out, nil
}

func normalizeLegalScanInput(in *model.LegalScanInput) {
	in.URL = strings.TrimSpace(in.URL)
	if in.URL != "" && !strings.HasPrefix(in.URL, "http://") && !strings.HasPrefix(in.URL, "https://") {
		in.URL = "https://" + in.URL
	}
	in.Industry = strings.TrimSpace(in.Industry)
	in.CompanySize = strings.TrimSpace(in.CompanySize)
	if in.Crawl.MaxPages <= 0 {
		in.Crawl.MaxPages = legalScanDefaultMaxPages
	}
	if in.Crawl.MaxDepth < 0 {
		in.Crawl.MaxDepth = legalScanDefaultMaxDepth
	}
	if !in.Crawl.SameHostOnly {
		in.Crawl.SameHostOnly = true
	}
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
	if in.Crawl.MaxPages < 1 || in.Crawl.MaxPages > legalScanCrawlMaxPagesCap {
		return fmt.Errorf("%w: max_pages out of range", ErrInvalidInput)
	}
	if in.Crawl.MaxDepth < 0 || in.Crawl.MaxDepth > 5 {
		return fmt.Errorf("%w: max_depth out of range", ErrInvalidInput)
	}
	return nil
}

var ErrLegalRiskNotFound = errors.New("legal risk not found")
