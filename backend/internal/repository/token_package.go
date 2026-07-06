package repository

import (
	"context"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenPackageRepository struct {
	pool *pgxpool.Pool
}

func NewTokenPackageRepository(pool *pgxpool.Pool) *TokenPackageRepository {
	return &TokenPackageRepository{pool: pool}
}

func (r *TokenPackageRepository) ListPublic(ctx context.Context) ([]model.TokenPackage, error) {
	const q = `
		SELECT id, slug, name, tokens, price_rub, sort_order, is_public, is_archived, created_at, updated_at
		FROM token_packages
		WHERE is_public = true AND is_archived = false
		ORDER BY sort_order ASC, price_rub ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTokenPackages(rows)
}

func (r *TokenPackageRepository) ListAdmin(ctx context.Context) ([]model.TokenPackage, error) {
	const q = `
		SELECT id, slug, name, tokens, price_rub, sort_order, is_public, is_archived, created_at, updated_at
		FROM token_packages
		ORDER BY sort_order ASC, price_rub ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTokenPackages(rows)
}

func (r *TokenPackageRepository) GetByID(ctx context.Context, id string) (*model.TokenPackage, error) {
	const q = `
		SELECT id, slug, name, tokens, price_rub, sort_order, is_public, is_archived, created_at, updated_at
		FROM token_packages WHERE id = $1
	`
	p, err := scanTokenPackage(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *TokenPackageRepository) Create(ctx context.Context, in model.TokenPackageUpsertInput) (*model.TokenPackage, error) {
	const q = `
		INSERT INTO token_packages (slug, name, tokens, price_rub, sort_order, is_public, is_archived)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, slug, name, tokens, price_rub, sort_order, is_public, is_archived, created_at, updated_at
	`
	return scanTokenPackage(r.pool.QueryRow(ctx, q,
		in.Slug, in.Name, in.Tokens, in.PriceRUB, in.SortOrder, in.IsPublic, in.IsArchived,
	))
}

func (r *TokenPackageRepository) Update(ctx context.Context, id string, in model.TokenPackageUpsertInput) (*model.TokenPackage, error) {
	const q = `
		UPDATE token_packages
		SET slug = $2, name = $3, tokens = $4, price_rub = $5, sort_order = $6,
		    is_public = $7, is_archived = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING id, slug, name, tokens, price_rub, sort_order, is_public, is_archived, created_at, updated_at
	`
	p, err := scanTokenPackage(r.pool.QueryRow(ctx, q,
		id, in.Slug, in.Name, in.Tokens, in.PriceRUB, in.SortOrder, in.IsPublic, in.IsArchived,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func scanTokenPackage(row pgx.Row) (*model.TokenPackage, error) {
	var p model.TokenPackage
	err := row.Scan(
		&p.ID, &p.Slug, &p.Name, &p.Tokens, &p.PriceRUB, &p.SortOrder,
		&p.IsPublic, &p.IsArchived, &p.CreatedAt, &p.UpdatedAt,
	)
	return &p, err
}

func scanTokenPackages(rows pgx.Rows) ([]model.TokenPackage, error) {
	var out []model.TokenPackage
	for rows.Next() {
		var p model.TokenPackage
		if err := rows.Scan(
			&p.ID, &p.Slug, &p.Name, &p.Tokens, &p.PriceRUB, &p.SortOrder,
			&p.IsPublic, &p.IsArchived, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
