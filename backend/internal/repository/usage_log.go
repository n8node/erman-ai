package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsageLogRepository struct {
	pool *pgxpool.Pool
}

func NewUsageLogRepository(pool *pgxpool.Pool) *UsageLogRepository {
	return &UsageLogRepository{pool: pool}
}

func (r *UsageLogRepository) Create(
	ctx context.Context,
	userID, runID, provider, modelName string,
	promptTokens, completionTokens int,
	costUSD, costRUB float64,
) error {
	const q = `
		INSERT INTO usage_log (user_id, run_id, provider, model, prompt_tokens, completion_tokens, total_tokens, cost_usd, cost_rub)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	total := promptTokens + completionTokens
	_, err := r.pool.Exec(ctx, q, userID, runID, provider, modelName, promptTokens, completionTokens, total, costUSD, costRUB)
	return err
}
