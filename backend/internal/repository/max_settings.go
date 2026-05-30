package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MaxSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewMaxSettingsRepository(pool *pgxpool.Pool) *MaxSettingsRepository {
	return &MaxSettingsRepository{pool: pool}
}

func (r *MaxSettingsRepository) Get(ctx context.Context) (*model.MaxSettingsRecord, error) {
	const q = `SELECT config, updated_at FROM max_settings WHERE id = 1`
	var raw []byte
	var rec model.MaxSettingsRecord
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

func (r *MaxSettingsRepository) Update(ctx context.Context, config model.MaxSettings) (*model.MaxSettingsRecord, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	const q = `
		UPDATE max_settings SET config = $1, updated_at = NOW()
		WHERE id = 1
		RETURNING config, updated_at
	`
	var out []byte
	var rec model.MaxSettingsRecord
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
