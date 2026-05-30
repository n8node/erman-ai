package repository

import (
	"context"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PublicPageRepository struct {
	pool *pgxpool.Pool
}

func NewPublicPageRepository(pool *pgxpool.Pool) *PublicPageRepository {
	return &PublicPageRepository{pool: pool}
}

func (r *PublicPageRepository) scan(row pgx.Row) (*model.PublicPage, error) {
	var p model.PublicPage
	err := row.Scan(
		&p.ID, &p.Slug, &p.Title, &p.ContentHTML, &p.MetaDescription,
		&p.IsPublished, &p.SortOrder, &p.UpdatedAt, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PublicPageRepository) ListPublishedSlugs(ctx context.Context) ([]string, error) {
	const q = `
		SELECT slug FROM public_pages
		WHERE is_published = true
		ORDER BY sort_order ASC, slug ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, err
		}
		slugs = append(slugs, slug)
	}
	return slugs, rows.Err()
}

func (r *PublicPageRepository) GetPublishedBySlug(ctx context.Context, slug string) (*model.PublicPage, error) {
	const q = `
		SELECT id, slug, title, content_html, meta_description, is_published, sort_order, updated_at, created_at
		FROM public_pages
		WHERE slug = $1 AND is_published = true
	`
	p, err := r.scan(r.pool.QueryRow(ctx, q, slug))
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PublicPageRepository) ListAll(ctx context.Context) ([]model.PublicPage, error) {
	const q = `
		SELECT id, slug, title, content_html, meta_description, is_published, sort_order, updated_at, created_at
		FROM public_pages
		ORDER BY sort_order ASC, slug ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.PublicPage
	for rows.Next() {
		p, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	return items, rows.Err()
}

func (r *PublicPageRepository) GetByID(ctx context.Context, id string) (*model.PublicPage, error) {
	const q = `
		SELECT id, slug, title, content_html, meta_description, is_published, sort_order, updated_at, created_at
		FROM public_pages
		WHERE id = $1
	`
	p, err := r.scan(r.pool.QueryRow(ctx, q, id))
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PublicPageRepository) SlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1 FROM public_pages WHERE slug = $1 AND ($2 = '' OR id::text <> $2)
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, q, slug, excludeID).Scan(&exists)
	return exists, err
}

func (r *PublicPageRepository) Create(
	ctx context.Context,
	slug, title, contentHTML, metaDescription string,
	isPublished bool,
	sortOrder int,
) (*model.PublicPage, error) {
	const q = `
		INSERT INTO public_pages (slug, title, content_html, meta_description, is_published, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, slug, title, content_html, meta_description, is_published, sort_order, updated_at, created_at
	`
	return r.scan(r.pool.QueryRow(ctx, q, slug, title, contentHTML, metaDescription, isPublished, sortOrder))
}

func (r *PublicPageRepository) Update(
	ctx context.Context,
	id, slug, title, contentHTML, metaDescription string,
	isPublished bool,
	sortOrder int,
) (*model.PublicPage, error) {
	const q = `
		UPDATE public_pages
		SET slug = $2, title = $3, content_html = $4, meta_description = $5,
		    is_published = $6, sort_order = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING id, slug, title, content_html, meta_description, is_published, sort_order, updated_at, created_at
	`
	p, err := r.scan(r.pool.QueryRow(ctx, q, id, slug, title, contentHTML, metaDescription, isPublished, sortOrder))
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PublicPageRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM public_pages WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
