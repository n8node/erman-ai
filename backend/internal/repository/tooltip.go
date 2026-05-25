package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TooltipRepository struct {
	pool *pgxpool.Pool
}

func NewTooltipRepository(pool *pgxpool.Pool) *TooltipRepository {
	return &TooltipRepository{pool: pool}
}

func scanTooltip(row pgx.Row) (model.UITooltip, error) {
	var t model.UITooltip
	var localeTexts []byte
	err := row.Scan(&t.Key, &t.Label, &t.TextRU, &t.TextEN, &localeTexts, &t.SortOrder, &t.UpdatedAt)
	if err != nil {
		return t, err
	}
	if len(localeTexts) > 0 {
		_ = json.Unmarshal(localeTexts, &t.LocaleTexts)
	}
	return t, nil
}

func (r *TooltipRepository) ListAll(ctx context.Context) ([]model.UITooltip, error) {
	const q = `
		SELECT key, label, text_ru, text_en, locale_texts, sort_order, updated_at
		FROM ui_tooltips
		ORDER BY sort_order ASC, key ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.UITooltip
	for rows.Next() {
		t, err := scanTooltip(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *TooltipRepository) ListByPrefix(ctx context.Context, prefix string) ([]model.UITooltip, error) {
	const q = `
		SELECT key, label, text_ru, text_en, locale_texts, sort_order, updated_at
		FROM ui_tooltips
		WHERE key LIKE $1
		ORDER BY sort_order ASC, key ASC
	`
	rows, err := r.pool.Query(ctx, q, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.UITooltip
	for rows.Next() {
		t, err := scanTooltip(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *TooltipRepository) Update(ctx context.Context, key, textRU, textEN string) (*model.UITooltip, error) {
	const q = `
		UPDATE ui_tooltips
		SET text_ru = $2, text_en = $3,
		    locale_texts = locale_texts || jsonb_build_object('ru', $2::text, 'en', $3::text),
		    updated_at = NOW()
		WHERE key = $1
		RETURNING key, label, text_ru, text_en, locale_texts, sort_order, updated_at
	`
	t, err := scanTooltip(r.pool.QueryRow(ctx, q, key, textRU, textEN))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}