package repository

import (
	"context"
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

func (r *TooltipRepository) ListAll(ctx context.Context) ([]model.UITooltip, error) {
	const q = `
		SELECT key, label, text_ru, text_en, sort_order, updated_at
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
		var t model.UITooltip
		if err := rows.Scan(&t.Key, &t.Label, &t.TextRU, &t.TextEN, &t.SortOrder, &t.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *TooltipRepository) ListByPrefix(ctx context.Context, prefix string) ([]model.UITooltip, error) {
	const q = `
		SELECT key, label, text_ru, text_en, sort_order, updated_at
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
		var t model.UITooltip
		if err := rows.Scan(&t.Key, &t.Label, &t.TextRU, &t.TextEN, &t.SortOrder, &t.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *TooltipRepository) Update(ctx context.Context, key, textRU, textEN string) (*model.UITooltip, error) {
	const q = `
		UPDATE ui_tooltips
		SET text_ru = $2, text_en = $3, updated_at = NOW()
		WHERE key = $1
		RETURNING key, label, text_ru, text_en, sort_order, updated_at
	`
	var t model.UITooltip
	err := r.pool.QueryRow(ctx, q, key, textRU, textEN).Scan(
		&t.Key, &t.Label, &t.TextRU, &t.TextEN, &t.SortOrder, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &t, err
}
