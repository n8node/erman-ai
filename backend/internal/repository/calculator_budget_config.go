package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CalculatorBudgetConfigRepository struct {
	pool *pgxpool.Pool
}

func NewCalculatorBudgetConfigRepository(pool *pgxpool.Pool) *CalculatorBudgetConfigRepository {
	return &CalculatorBudgetConfigRepository{pool: pool}
}

func (r *CalculatorBudgetConfigRepository) Get(ctx context.Context) (*model.CalculatorBudgetConfigRecord, error) {
	const q = `
		SELECT config, updated_at
		FROM calculator_budget_config
		WHERE id = 1
	`
	var raw []byte
	var rec model.CalculatorBudgetConfigRecord
	err := r.pool.QueryRow(ctx, q).Scan(&raw, &rec.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &rec.Config); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *CalculatorBudgetConfigRepository) Update(ctx context.Context, cfg model.CalculatorBudgetConfig) (*model.CalculatorBudgetConfigRecord, error) {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	const q = `
		UPDATE calculator_budget_config
		SET config = $1, updated_at = NOW()
		WHERE id = 1
		RETURNING config, updated_at
	`
	var out []byte
	var rec model.CalculatorBudgetConfigRecord
	err = r.pool.QueryRow(ctx, q, raw).Scan(&out, &rec.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(out, &rec.Config); err != nil {
		return nil, err
	}
	return &rec, nil
}
