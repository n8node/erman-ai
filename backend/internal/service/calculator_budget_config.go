package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidBudgetConfig = errors.New("invalid budget config")

type CalculatorBudgetConfigService struct {
	repo *repository.CalculatorBudgetConfigRepository
}

func NewCalculatorBudgetConfigService(repo *repository.CalculatorBudgetConfigRepository) *CalculatorBudgetConfigService {
	return &CalculatorBudgetConfigService{repo: repo}
}

func (s *CalculatorBudgetConfigService) Get(ctx context.Context) (*model.CalculatorBudgetConfigRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := model.DefaultCalculatorBudgetConfig()
			return &model.CalculatorBudgetConfigRecord{Config: def}, nil
		}
		return nil, err
	}
	return rec, nil
}

func (s *CalculatorBudgetConfigService) Update(ctx context.Context, cfg model.CalculatorBudgetConfig) (*model.CalculatorBudgetConfigRecord, error) {
	if err := validateBudgetConfig(cfg); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, cfg)
}

func validateBudgetConfig(cfg model.CalculatorBudgetConfig) error {
	positive := []struct {
		name string
		v    float64
	}{
		{"dev_rate_rub", cfg.DevRateRub},
		{"audit_base", cfg.AuditBase},
		{"training", cfg.Training},
		{"om_base", cfg.OmBase},
		{"capex_min", cfg.CapexMin},
		{"integration_mult_simple", cfg.IntegrationMultSimple},
		{"integration_mult_standard", cfg.IntegrationMultStandard},
		{"integration_mult_complex", cfg.IntegrationMultComplex},
		{"capex_round_step", cfg.CapexRoundStep},
		{"om_round_step", cfg.OmRoundStep},
	}
	for _, p := range positive {
		if p.v <= 0 {
			return fmt.Errorf("%w: %s must be positive", ErrInvalidBudgetConfig, p.name)
		}
	}
	if cfg.TierSimpleMax < 0 || cfg.TierStandardMax <= cfg.TierSimpleMax {
		return fmt.Errorf("%w: tier thresholds invalid", ErrInvalidBudgetConfig)
	}
	return nil
}
