package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InquiryVerificationTokenRepository struct {
	pool *pgxpool.Pool
}

func NewInquiryVerificationTokenRepository(pool *pgxpool.Pool) *InquiryVerificationTokenRepository {
	return &InquiryVerificationTokenRepository{pool: pool}
}

func (r *InquiryVerificationTokenRepository) Upsert(ctx context.Context, inquiryID, tokenHash string, expiresAt time.Time) error {
	const q = `
		INSERT INTO inquiry_verification_tokens (inquiry_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (inquiry_id) DO UPDATE
		SET token_hash = EXCLUDED.token_hash, expires_at = EXCLUDED.expires_at, created_at = NOW()
	`
	_, err := r.pool.Exec(ctx, q, inquiryID, tokenHash, expiresAt)
	return err
}

func (r *InquiryVerificationTokenRepository) FindInquiryID(ctx context.Context, tokenHash string, now time.Time) (string, error) {
	const q = `
		SELECT inquiry_id FROM inquiry_verification_tokens
		WHERE token_hash = $1 AND expires_at > $2
	`
	var inquiryID string
	err := r.pool.QueryRow(ctx, q, tokenHash, now).Scan(&inquiryID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return inquiryID, err
}

func (r *InquiryVerificationTokenRepository) Delete(ctx context.Context, inquiryID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM inquiry_verification_tokens WHERE inquiry_id = $1`, inquiryID)
	return err
}
