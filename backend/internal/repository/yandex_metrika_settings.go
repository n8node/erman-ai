package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type YandexMetrikaSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewYandexMetrikaSettingsRepository(pool *pgxpool.Pool) *YandexMetrikaSettingsRepository {
	return &YandexMetrikaSettingsRepository{pool: pool}
}

func (r *YandexMetrikaSettingsRepository) Get(ctx context.Context) (*model.YandexMetrikaSettingsRecord, error) {
	const q = `SELECT config, updated_at FROM yandex_metrika_settings WHERE id = 1`
	var raw []byte
	var rec model.YandexMetrikaSettingsRecord
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

func (r *YandexMetrikaSettingsRepository) Update(ctx context.Context, config model.YandexMetrikaSettings) (*model.YandexMetrikaSettingsRecord, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	const q = `
		UPDATE yandex_metrika_settings SET config = $1, updated_at = NOW()
		WHERE id = 1
		RETURNING config, updated_at
	`
	var out []byte
	var rec model.YandexMetrikaSettingsRecord
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
