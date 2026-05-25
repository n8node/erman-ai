package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

type RunService struct {
	runs  *repository.ToolRunRepository
	plans *repository.PlanRepository
}

func NewRunService(runs *repository.ToolRunRepository, plans *repository.PlanRepository) *RunService {
	return &RunService{runs: runs, plans: plans}
}

func HistoryRetention(planSlug string) time.Duration {
	switch planSlug {
	case "pro", "business":
		return 365 * 24 * time.Hour
	default:
		return 90 * 24 * time.Hour
	}
}

type RunListItem struct {
	ID                string    `json:"id"`
	ToolSlug          string    `json:"tool_slug"`
	ProcessName       string    `json:"process_name,omitempty"`
	NetBenefitMonthly float64   `json:"net_benefit_monthly,omitempty"`
	PaybackMonths     float64   `json:"payback_months,omitempty"`
	ROIHorizonPct     float64   `json:"roi_horizon_pct,omitempty"`
	Recommendation    string    `json:"recommendation,omitempty"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
}

func (s *RunService) List(ctx context.Context, userID, toolSlug string, limit, offset int) ([]RunListItem, int, error) {
	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	since := time.Now().UTC().Add(-HistoryRetention(up.PlanSlug))
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	total, err := s.runs.CountByUserFiltered(ctx, userID, toolSlug, since)
	if err != nil {
		return nil, 0, err
	}

	runs, err := s.runs.ListByUserFiltered(ctx, userID, toolSlug, since, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	items := make([]RunListItem, 0, len(runs))
	for _, run := range runs {
		items = append(items, summarizeRun(run))
	}
	return items, total, nil
}

func (s *RunService) Get(ctx context.Context, runID, userID string) (*model.ToolRun, error) {
	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, err
	}
	run, err := s.runs.GetByIDForUser(ctx, runID, userID)
	if err != nil {
		return nil, err
	}
	since := time.Now().UTC().Add(-HistoryRetention(up.PlanSlug))
	if run.CreatedAt.Before(since) {
		return nil, repository.ErrNotFound
	}
	return run, nil
}

func (s *RunService) Delete(ctx context.Context, runID, userID string) error {
	return s.runs.DeleteForUser(ctx, runID, userID)
}

func summarizeRun(run model.ToolRun) RunListItem {
	item := RunListItem{
		ID:        run.ID,
		ToolSlug:  run.ToolSlug,
		Status:    string(run.Status),
		CreatedAt: run.CreatedAt,
	}
	if run.ToolSlug != "calculator" {
		if run.ToolSlug == "strategy" {
			var input model.StrategyInput
			if err := json.Unmarshal(run.Input, &input); err == nil {
				item.ProcessName = input.CompanyName
			}
		}
		if run.ToolSlug == "proposal" {
			var input model.ProposalInput
			if err := json.Unmarshal(run.Input, &input); err == nil {
				item.ProcessName = input.ClientCompany
			}
		}
		if run.ToolSlug == "audit" {
			var input model.AuditInput
			if err := json.Unmarshal(run.Input, &input); err == nil {
				item.ProcessName = input.CompanyName
			}
		}
		return item
	}
	var input model.CalculatorInput
	if err := json.Unmarshal(run.Input, &input); err == nil {
		item.ProcessName = input.ProcessName
	}
	var output model.CalculatorOutput
	if err := json.Unmarshal(run.Output, &output); err == nil {
		item.NetBenefitMonthly = output.NetBenefitMonthly
		item.PaybackMonths = output.PaybackMonths
		item.ROIHorizonPct = output.ROIHorizonPct
		item.Recommendation = string(output.Recommendation)
	}
	return item
}
