package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserTokenBalanceRepository struct {
	pool *pgxpool.Pool
}

func NewUserTokenBalanceRepository(pool *pgxpool.Pool) *UserTokenBalanceRepository {
	return &UserTokenBalanceRepository{pool: pool}
}

func (r *UserTokenBalanceRepository) GetBalance(ctx context.Context, userID string) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx, `
		SELECT balance FROM user_token_balances WHERE user_id = $1
	`, userID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return balance, err
}

func (r *UserTokenBalanceRepository) AddTokens(ctx context.Context, userID string, delta int64) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO user_token_balances (user_id, balance, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE
		SET balance = user_token_balances.balance + EXCLUDED.balance,
		    updated_at = NOW()
		RETURNING balance
	`, userID, delta).Scan(&balance)
	return balance, err
}
