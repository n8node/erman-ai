package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetTokenRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetTokenRepository(pool *pgxpool.Pool) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{pool: pool}
}

func (r *PasswordResetTokenRepository) DeleteForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM password_reset_tokens WHERE user_id = $1`, userID)
	return err
}

func (r *PasswordResetTokenRepository) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *PasswordResetTokenRepository) FindValidUserID(ctx context.Context, tokenHash string, now time.Time) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `
		SELECT user_id FROM password_reset_tokens
		WHERE token_hash = $1 AND expires_at > $2
		ORDER BY created_at DESC
		LIMIT 1
	`, tokenHash, now).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return userID, err
}

func (r *PasswordResetTokenRepository) DeleteByHash(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM password_reset_tokens WHERE token_hash = $1`, tokenHash)
	return err
}
