package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var (
	ErrFeatureNotAvailable = errors.New("feature not available on current plan")
	ErrShareLimitExceeded  = errors.New("share report limit exceeded")
	ErrPartnerOnly         = errors.New("partner accounts only")
)

type BillingService struct {
	plans   *repository.PlanRepository
	runs    *repository.ToolRunRepository
}

func NewBillingService(plans *repository.PlanRepository, runs *repository.ToolRunRepository) *BillingService {
	return &BillingService{plans: plans, runs: runs}
}

func (s *BillingService) GetUserPlan(ctx context.Context, userID string) (*repository.UserPlan, error) {
	return s.plans.GetUserPlan(ctx, userID)
}

func (s *BillingService) HasFeature(plan *model.Plan, key string) bool {
	if plan.Features == nil {
		return false
	}
	v, ok := plan.Features[key]
	if !ok {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	default:
		return false
	}
}

func (s *BillingService) ShareReportLimit(plan *model.Plan) int {
	if plan.Features == nil {
		return 0
	}
	v, ok := plan.Features["share_report_limit"]
	if !ok {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	default:
		return 0
	}
}

func (s *BillingService) CanShareReport(ctx context.Context, userID string) error {
	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return err
	}
	if !s.HasFeature(&up.Plan, "share_report") {
		return ErrFeatureNotAvailable
	}
	limit := s.ShareReportLimit(&up.Plan)
	if limit == -1 {
		return nil
	}
	count, err := s.runs.GetUsageCount(ctx, userID, "share_report")
	if err != nil {
		return err
	}
	if count >= limit {
		return ErrShareLimitExceeded
	}
	return nil
}

func (s *BillingService) CanExportPDF(ctx context.Context, userID string) error {
	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return err
	}
	if !s.HasFeature(&up.Plan, "export_pdf") {
		return ErrFeatureNotAvailable
	}
	return nil
}

type CalculatorService struct {
	cfg   *config.Config
	runs  *repository.ToolRunRepository
	plans *repository.PlanRepository
}

func NewCalculatorService(cfg *config.Config, runs *repository.ToolRunRepository, plans *repository.PlanRepository) *CalculatorService {
	return &CalculatorService{cfg: cfg, runs: runs, plans: plans}
}

func (s *CalculatorService) Calculate(input model.CalculatorInput, locale string) model.CalculatorOutput {
	return CalculateExcel(input, locale)
}

func (s *CalculatorService) Run(ctx context.Context, userID string, input model.CalculatorInput, locale string) (*model.ToolRun, model.CalculatorOutput, error) {
	if err := validateCalculatorInput(input); err != nil {
		return nil, model.CalculatorOutput{}, err
	}

	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, model.CalculatorOutput{}, err
	}

	// Sync Hm from process steps when in process mode
	if input.HmMode == "process" {
		input.HoursSavedMonth = ResolveHoursSaved(input)
	}

	output := CalculateExcel(input, locale)
	inJSON, _ := json.Marshal(input)
	outJSON, _ := json.Marshal(output)

	run, err := s.runs.Create(ctx, userID, "calculator", up.PlanSlug, inJSON, outJSON)
	if err != nil {
		return nil, model.CalculatorOutput{}, err
	}
	return run, output, nil
}

func (s *CalculatorService) GetRunForUser(ctx context.Context, runID, userID string) (*model.ToolRun, error) {
	return s.runs.GetByIDForUser(ctx, runID, userID)
}

func validateCalculatorInput(in model.CalculatorInput) error {
	if in.ProcessName == "" {
		return ErrInvalidInput
	}
	if in.HourlyCostLoaded <= 0 {
		return ErrInvalidInput
	}
	if in.Capex < 0 || in.MonthlySolutionCost < 0 || in.OtherBenefitMonthly < 0 {
		return ErrInvalidInput
	}
	Hm := ResolveHoursSaved(in)
	if Hm <= 0 {
		return ErrInvalidInput
	}
	if in.Utilization < 0 || in.Utilization > 1 {
		return ErrInvalidInput
	}
	if in.AutomationPct < 0 || in.AutomationPct > 100 {
		return ErrInvalidInput
	}
	if in.ErrorRateBeforePct < 0 || in.ErrorRateBeforePct > 100 {
		return ErrInvalidInput
	}
	return nil
}

func round2(v float64) float64 {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return 0
	}
	return math.Round(v*100) / 100
}

func FormatMoney(v float64) string {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return "—"
	}
	return fmt.Sprintf("%.0f", v)
}
