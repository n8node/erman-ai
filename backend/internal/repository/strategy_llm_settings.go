package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StrategyLLMSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewStrategyLLMSettingsRepository(pool *pgxpool.Pool) *StrategyLLMSettingsRepository {
	return &StrategyLLMSettingsRepository{pool: pool}
}

func (r *StrategyLLMSettingsRepository) Get(ctx context.Context) (*model.StrategyLLMSettingsRecord, error) {
	const q = `
		SELECT config, updated_at
		FROM strategy_llm_settings
		WHERE id = 1
	`
	var raw []byte
	var rec model.StrategyLLMSettingsRecord
	err := r.pool.QueryRow(ctx, q).Scan(&raw, &rec.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &rec.Settings); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *StrategyLLMSettingsRepository) Update(ctx context.Context, settings model.StrategyLLMSettings) (*model.StrategyLLMSettingsRecord, error) {
	raw, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	const q = `
		UPDATE strategy_llm_settings
		SET config = $1, updated_at = NOW()
		WHERE id = 1
		RETURNING config, updated_at
	`
	var out []byte
	var rec model.StrategyLLMSettingsRecord
	err = r.pool.QueryRow(ctx, q, raw).Scan(&out, &rec.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(out, &rec.Settings); err != nil {
		return nil, err
	}
	return &rec, nil
}
