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
	return r.Upsert(ctx, config)
}

func (r *YandexMetrikaSettingsRepository) Upsert(ctx context.Context, config model.YandexMetrikaSettings) (*model.YandexMetrikaSettingsRecord, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	const q = `
		INSERT INTO yandex_metrika_settings (id, config, updated_at)
		VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE
		SET config = EXCLUDED.config, updated_at = NOW()
		RETURNING config, updated_at
	`
	var out []byte
	var rec model.YandexMetrikaSettingsRecord
	err = r.pool.QueryRow(ctx, q, raw).Scan(&out, &rec.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(out, &rec.Config); err != nil {
		return nil, err
	}
	return &rec, nil
}
