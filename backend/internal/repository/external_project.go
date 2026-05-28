package repository

import (
	"context"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExternalProjectRepository struct {
	pool *pgxpool.Pool
}

func NewExternalProjectRepository(pool *pgxpool.Pool) *ExternalProjectRepository {
	return &ExternalProjectRepository{pool: pool}
}

func (r *ExternalProjectRepository) scan(row pgx.Row) (*model.ExternalProject, error) {
	var p model.ExternalProject
	err := row.Scan(
		&p.ID, &p.Title, &p.URL, &p.SortOrder, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ExternalProjectRepository) ListEnabled(ctx context.Context) ([]model.ExternalProject, error) {
	const q = `
		SELECT id, title, url, sort_order, is_enabled, created_at, updated_at
		FROM external_projects
		WHERE is_enabled = true
		ORDER BY sort_order ASC, title ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ExternalProject
	for rows.Next() {
		var p model.ExternalProject
		if err := rows.Scan(
			&p.ID, &p.Title, &p.URL, &p.SortOrder, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *ExternalProjectRepository) ListAll(ctx context.Context) ([]model.ExternalProject, error) {
	const q = `
		SELECT id, title, url, sort_order, is_enabled, created_at, updated_at
		FROM external_projects
		ORDER BY sort_order ASC, title ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ExternalProject
	for rows.Next() {
		var p model.ExternalProject
		if err := rows.Scan(
			&p.ID, &p.Title, &p.URL, &p.SortOrder, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *ExternalProjectRepository) GetByID(ctx context.Context, id string) (*model.ExternalProject, error) {
	const q = `
		SELECT id, title, url, sort_order, is_enabled, created_at, updated_at
		FROM external_projects
		WHERE id = $1
	`
	p, err := r.scan(r.pool.QueryRow(ctx, q, id))
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *ExternalProjectRepository) Create(
	ctx context.Context,
	title, url string,
	sortOrder int,
	isEnabled bool,
) (*model.ExternalProject, error) {
	const q = `
		INSERT INTO external_projects (title, url, sort_order, is_enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, url, sort_order, is_enabled, created_at, updated_at
	`
	return r.scan(r.pool.QueryRow(ctx, q, title, url, sortOrder, isEnabled))
}

func (r *ExternalProjectRepository) Update(
	ctx context.Context,
	id, title, url string,
	sortOrder int,
	isEnabled bool,
) (*model.ExternalProject, error) {
	const q = `
		UPDATE external_projects
		SET title = $2, url = $3, sort_order = $4, is_enabled = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING id, title, url, sort_order, is_enabled, created_at, updated_at
	`
	p, err := r.scan(r.pool.QueryRow(ctx, q, id, title, url, sortOrder, isEnabled))
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *ExternalProjectRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM external_projects WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
