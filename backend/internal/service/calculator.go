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
	rate := s.cfg.AutomationSavingsRate
	if rate <= 0 || rate > 1 {
		rate = 0.7
	}

	currentMonthly := input.HoursPerMonth * input.HourlyRateRUB * input.EmployeesCount
	errorMonthly := (input.ErrorRatePct / 100) * (input.HoursPerMonth * input.EmployeesCount) * input.ErrorCostRUB
	totalCurrent := currentMonthly + errorMonthly
	monthlySavings := totalCurrent * rate
	netMonthly := monthlySavings - input.MonthlySupportRUB

	var payback float64
	if netMonthly > 0 {
		payback = input.AutomationCostRUB / netMonthly
	} else {
		payback = math.Inf(1)
	}

	rec, text := recommendation(payback, locale)

	return model.CalculatorOutput{
		CurrentMonthlyCost: round2(currentMonthly),
		ErrorMonthlyCost:     round2(errorMonthly),
		TotalCurrentCost:     round2(totalCurrent),
		MonthlySavings:       round2(monthlySavings),
		NetMonthlySavings:    round2(netMonthly),
		PaybackMonths:        round2(payback),
		AnnualSavings:        round2(netMonthly * 12),
		Recommendation:       rec,
		RecommendationText:   text,
	}
}

func (s *CalculatorService) Run(ctx context.Context, userID string, input model.CalculatorInput, locale string) (*model.ToolRun, model.CalculatorOutput, error) {
	if err := validateCalculatorInput(input); err != nil {
		return nil, model.CalculatorOutput{}, err
	}

	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, model.CalculatorOutput{}, err
	}

	output := s.Calculate(input, locale)
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
	if in.HoursPerMonth <= 0 || in.HourlyRateRUB <= 0 || in.EmployeesCount <= 0 {
		return ErrInvalidInput
	}
	if in.ErrorRatePct < 0 || in.ErrorRatePct > 100 {
		return ErrInvalidInput
	}
	if in.AutomationCostRUB < 0 || in.MonthlySupportRUB < 0 || in.ErrorCostRUB < 0 {
		return ErrInvalidInput
	}
	return nil
}

func recommendation(payback float64, locale string) (model.Recommendation, string) {
	if math.IsInf(payback, 1) || payback > 24 {
		if locale == "en" {
			return model.RecommendationNotRecommended, "Automation is not recommended at current parameters"
		}
		return model.RecommendationNotRecommended, "Автоматизация не рекомендуется при текущих параметрах"
	}
	if payback >= 12 {
		if locale == "en" {
			return model.RecommendationConsider, "Consider automation — payback period is moderate"
		}
		return model.RecommendationConsider, "Рассмотрите автоматизацию — срок окупаемости умеренный"
	}
	if locale == "en" {
		return model.RecommendationAutomate, "Automation is recommended — payback under 12 months"
	}
	return model.RecommendationAutomate, "Рекомендуется автоматизировать — окупаемость менее 12 месяцев"
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
