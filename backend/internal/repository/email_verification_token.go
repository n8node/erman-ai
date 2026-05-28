package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailVerificationTokenRepository struct {
	pool *pgxpool.Pool
}

func NewEmailVerificationTokenRepository(pool *pgxpool.Pool) *EmailVerificationTokenRepository {
	return &EmailVerificationTokenRepository{pool: pool}
}

func (r *EmailVerificationTokenRepository) DeleteForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM email_verification_tokens WHERE user_id = $1`, userID)
	return err
}

func (r *EmailVerificationTokenRepository) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *EmailVerificationTokenRepository) FindValidUserID(ctx context.Context, tokenHash string, now time.Time) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `
		SELECT user_id FROM email_verification_tokens
		WHERE token_hash = $1 AND expires_at > $2
		ORDER BY created_at DESC
		LIMIT 1
	`, tokenHash, now).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return userID, err
}

func (r *EmailVerificationTokenRepository) DeleteByHash(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM email_verification_tokens WHERE token_hash = $1`, tokenHash)
	return err
}
