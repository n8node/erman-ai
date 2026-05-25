package repository

import (
	"context"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TranslationRepository struct {
	pool *pgxpool.Pool
}

func NewTranslationRepository(pool *pgxpool.Pool) *TranslationRepository {
	return &TranslationRepository{pool: pool}
}

func (r *TranslationRepository) ListByLocale(ctx context.Context, locale string) (map[string]string, error) {
	const q = `
		SELECT key, value
		FROM ui_translations
		WHERE locale = $1
		ORDER BY key ASC
	`
	rows, err := r.pool.Query(ctx, q, locale)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		out[key] = value
	}
	return out, rows.Err()
}

func (r *TranslationRepository) Search(ctx context.Context, locale, search string, limit, offset int) ([]model.UITranslation, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
		SELECT key, locale, value, updated_at
		FROM ui_translations
		WHERE ($1 = '' OR locale = $1)
		  AND ($2 = '' OR key ILIKE '%' || $2 || '%' OR value ILIKE '%' || $2 || '%')
		ORDER BY key ASC, locale ASC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, q, locale, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.UITranslation
	for rows.Next() {
		var t model.UITranslation
		if err := rows.Scan(&t.Key, &t.Locale, &t.Value, &t.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *TranslationRepository) Upsert(ctx context.Context, key, locale, value string) (*model.UITranslation, error) {
	const q = `
		INSERT INTO ui_translations (key, locale, value, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (key, locale) DO UPDATE
		SET value = EXCLUDED.value, updated_at = NOW()
		RETURNING key, locale, value, updated_at
	`
	var t model.UITranslation
	err := r.pool.QueryRow(ctx, q, key, locale, value).Scan(&t.Key, &t.Locale, &t.Value, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TranslationRepository) Delete(ctx context.Context, key, locale string) error {
	const q = `DELETE FROM ui_translations WHERE key = $1 AND locale = $2`
	tag, err := r.pool.Exec(ctx, q, key, locale)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TranslationRepository) Get(ctx context.Context, key, locale string) (*model.UITranslation, error) {
	const q = `
		SELECT key, locale, value, updated_at
		FROM ui_translations
		WHERE key = $1 AND locale = $2
	`
	var t model.UITranslation
	err := r.pool.QueryRow(ctx, q, key, locale).Scan(&t.Key, &t.Locale, &t.Value, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &t, err
}
