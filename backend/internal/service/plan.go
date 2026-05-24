package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var (
	ErrPlanSlugTaken   = errors.New("plan slug already exists")
	ErrPlanInvalidSlug = errors.New("invalid plan slug")
)

var planSlugRe = regexp.MustCompile(`^[a-z][a-z0-9_]{0,49}$`)

type PlanService struct {
	plans *repository.PlanRepository
}

func NewPlanService(plans *repository.PlanRepository) *PlanService {
	return &PlanService{plans: plans}
}

type AdminPlan struct {
	model.Plan
	UserCount int `json:"user_count"`
}

func (s *PlanService) ListAdmin(ctx context.Context) ([]AdminPlan, error) {
	rows, err := s.plans.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AdminPlan, len(rows))
	for i, row := range rows {
		out[i] = AdminPlan{Plan: row.Plan, UserCount: row.UserCount}
	}
	return out, nil
}

func (s *PlanService) GetAdmin(ctx context.Context, id string) (*AdminPlan, error) {
	p, err := s.plans.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	rows, err := s.plans.ListAll(ctx)
	if err != nil {
		return &AdminPlan{Plan: *p, UserCount: 0}, nil
	}
	for _, row := range rows {
		if row.Plan.ID == id {
			return &AdminPlan{Plan: *p, UserCount: row.UserCount}, nil
		}
	}
	return &AdminPlan{Plan: *p, UserCount: 0}, nil
}

func (s *PlanService) ListTools(ctx context.Context) ([]model.ToolConfig, error) {
	return s.plans.ListAllTools(ctx)
}

type PlanInput struct {
	Slug            string
	Name            string
	PriceMonthlyRUB int
	PriceYearlyRUB  int
	ToolLimits      map[string]int
	Features        map[string]any
	SupportLevel    string
	IsPublic        bool
	IsArchived      bool
}

func (s *PlanService) Create(ctx context.Context, in PlanInput) (*model.Plan, error) {
	if err := normalizePlanInput(&in, true); err != nil {
		return nil, err
	}
	exists, err := s.plans.SlugExists(ctx, in.Slug, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPlanSlugTaken
	}
	p := inputToPlan(in)
	return s.plans.Create(ctx, p)
}

func (s *PlanService) Update(ctx context.Context, id string, in PlanInput) (*model.Plan, error) {
	if err := normalizePlanInput(&in, false); err != nil {
		return nil, err
	}
	current, err := s.plans.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	p := inputToPlan(in)
	p.ID = current.ID
	p.Slug = current.Slug
	return s.plans.Update(ctx, id, p)
}

func normalizePlanInput(in *PlanInput, creating bool) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Slug = strings.TrimSpace(strings.ToLower(in.Slug))
	in.SupportLevel = strings.TrimSpace(in.SupportLevel)

	if in.Name == "" {
		return ErrInvalidInput
	}
	if creating {
		if in.Slug == "" || !planSlugRe.MatchString(in.Slug) {
			return ErrPlanInvalidSlug
		}
	}
	if in.SupportLevel == "" {
		in.SupportLevel = "community"
	}
	if in.ToolLimits == nil {
		in.ToolLimits = map[string]int{}
	}
	if in.Features == nil {
		in.Features = map[string]any{}
	}
	return nil
}

func inputToPlan(in PlanInput) *model.Plan {
	return &model.Plan{
		Slug:            in.Slug,
		Name:            in.Name,
		PriceMonthlyRUB: in.PriceMonthlyRUB,
		PriceYearlyRUB:  in.PriceYearlyRUB,
		ToolLimits:      in.ToolLimits,
		Features:        in.Features,
		SupportLevel:    in.SupportLevel,
		IsPublic:        in.IsPublic,
		IsArchived:      in.IsArchived,
	}
}

func DefaultPlanFeatures() map[string]any {
	return map[string]any{
		"export_pdf":          false,
		"export_docx":         false,
		"api_access":          false,
		"priority_queue":      false,
		"white_label":         false,
		"hide_share_promo":    false,
		"share_report":        false,
		"share_report_limit":  float64(0),
	}
}
