package repository

import (
	"context"
	"errors"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SharedReportRepository struct {
	pool *pgxpool.Pool
}

func NewSharedReportRepository(pool *pgxpool.Pool) *SharedReportRepository {
	return &SharedReportRepository{pool: pool}
}

func (r *SharedReportRepository) Create(ctx context.Context, runID, userID, token string, expiresAt time.Time) (*model.SharedReport, error) {
	const q = `
		INSERT INTO shared_reports (run_id, user_id, token, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, run_id, user_id, token, expires_at, view_count, revoked_at, created_at
	`
	return r.scan(r.pool.QueryRow(ctx, q, runID, userID, token, expiresAt))
}

func (r *SharedReportRepository) GetByToken(ctx context.Context, token string) (*model.SharedReport, error) {
	const q = `
		SELECT id, run_id, user_id, token, expires_at, view_count, revoked_at, created_at
		FROM shared_reports WHERE token = $1
	`
	sr, err := r.scan(r.pool.QueryRow(ctx, q, token))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return sr, err
}

func (r *SharedReportRepository) IncrementView(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE shared_reports SET view_count = view_count + 1 WHERE id = $1`, id)
	return err
}

func (r *SharedReportRepository) scan(row pgx.Row) (*model.SharedReport, error) {
	var sr model.SharedReport
	err := row.Scan(&sr.ID, &sr.RunID, &sr.UserID, &sr.Token, &sr.ExpiresAt, &sr.ViewCount, &sr.RevokedAt, &sr.CreatedAt)
	return &sr, err
}

type LeadRepository struct {
	pool *pgxpool.Pool
}

func NewLeadRepository(pool *pgxpool.Pool) *LeadRepository {
	return &LeadRepository{pool: pool}
}

func (r *LeadRepository) Create(ctx context.Context, userID, runID, name, email string, company, phone, message *string) (*model.Lead, error) {
	const q = `
		INSERT INTO leads (user_id, run_id, name, company, phone, email, message)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, run_id, name, company, phone, email, message, created_at
	`
	var lead model.Lead
	err := r.pool.QueryRow(ctx, q, userID, runID, name, company, phone, email, message).Scan(
		&lead.ID, &lead.UserID, &lead.RunID, &lead.Name, &lead.Company, &lead.Phone, &lead.Email, &lead.Message, &lead.CreatedAt,
	)
	return &lead, err
}
