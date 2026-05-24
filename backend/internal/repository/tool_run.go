package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ToolRunRepository struct {
	pool *pgxpool.Pool
}

func NewToolRunRepository(pool *pgxpool.Pool) *ToolRunRepository {
	return &ToolRunRepository{pool: pool}
}

func (r *ToolRunRepository) Create(ctx context.Context, userID, toolSlug, planTier string, input json.RawMessage, output json.RawMessage) (*model.ToolRun, error) {
	const q = `
		INSERT INTO tool_runs (user_id, tool_slug, plan_tier, input, output, status, completed_at)
		VALUES ($1, $2, $3, $4, $5, 'done', NOW())
		RETURNING id, user_id, tool_slug, plan_tier, input, output, artifact_url, tokens_used, model_used, status, error_msg, created_at, updated_at, completed_at
	`
	return r.scan(r.pool.QueryRow(ctx, q, userID, toolSlug, planTier, input, output))
}

func (r *ToolRunRepository) CreatePending(ctx context.Context, userID, toolSlug, planTier string, input json.RawMessage) (*model.ToolRun, error) {
	const q = `
		INSERT INTO tool_runs (user_id, tool_slug, plan_tier, input, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, user_id, tool_slug, plan_tier, input, output, artifact_url, tokens_used, model_used, status, error_msg, created_at, updated_at, completed_at
	`
	return r.scan(r.pool.QueryRow(ctx, q, userID, toolSlug, planTier, input))
}

func (r *ToolRunRepository) UpdateStatus(ctx context.Context, id string, status model.RunStatus) error {
	const q = `
		UPDATE tool_runs SET status = $2, updated_at = NOW() WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, q, id, string(status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ToolRunRepository) UpdateRunDone(ctx context.Context, id string, output json.RawMessage, tokens int64, modelUsed string) error {
	const q = `
		UPDATE tool_runs
		SET status = 'done', output = $2, tokens_used = $3, model_used = $4,
		    completed_at = NOW(), updated_at = NOW(), error_msg = NULL
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, q, id, output, tokens, modelUsed)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ToolRunRepository) UpdateRunError(ctx context.Context, id, errMsg string) error {
	const q = `
		UPDATE tool_runs
		SET status = 'error', error_msg = $2, updated_at = NOW(), completed_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, q, id, errMsg)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ToolRunRepository) GetByID(ctx context.Context, id string) (*model.ToolRun, error) {
	const q = `
		SELECT id, user_id, tool_slug, plan_tier, input, output, artifact_url, tokens_used, model_used, status, error_msg, created_at, updated_at, completed_at
		FROM tool_runs WHERE id = $1
	`
	run, err := r.scan(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return run, err
}

func (r *ToolRunRepository) GetByIDForUser(ctx context.Context, id, userID string) (*model.ToolRun, error) {
	run, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if run.UserID != userID {
		return nil, ErrNotFound
	}
	return run, nil
}

func (r *ToolRunRepository) ListByUser(ctx context.Context, userID string, limit int) ([]model.ToolRun, error) {
	if limit <= 0 {
		limit = 20
	}
	return r.ListByUserFiltered(ctx, userID, "", time.Time{}, limit, 0)
}

func (r *ToolRunRepository) ListByUserFiltered(ctx context.Context, userID, toolSlug string, since time.Time, limit, offset int) ([]model.ToolRun, error) {
	if limit <= 0 {
		limit = 20
	}
	const q = `
		SELECT id, user_id, tool_slug, plan_tier, input, output, artifact_url, tokens_used, model_used, status, error_msg, created_at, updated_at, completed_at
		FROM tool_runs
		WHERE user_id = $1
		  AND ($2 = '' OR tool_slug = $2)
		  AND ($3::timestamptz IS NULL OR created_at >= $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`
	var sinceArg any
	if !since.IsZero() {
		sinceArg = since
	}
	rows, err := r.pool.Query(ctx, q, userID, toolSlug, sinceArg, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []model.ToolRun
	for rows.Next() {
		run, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *run)
	}
	return runs, rows.Err()
}

func (r *ToolRunRepository) CountByUserFiltered(ctx context.Context, userID, toolSlug string, since time.Time) (int, error) {
	const q = `
		SELECT COUNT(*)
		FROM tool_runs
		WHERE user_id = $1
		  AND ($2 = '' OR tool_slug = $2)
		  AND ($3::timestamptz IS NULL OR created_at >= $3)
	`
	var sinceArg any
	if !since.IsZero() {
		sinceArg = since
	}
	var count int
	err := r.pool.QueryRow(ctx, q, userID, toolSlug, sinceArg).Scan(&count)
	return count, err
}

func (r *ToolRunRepository) DeleteForUser(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tool_runs WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ToolRunRepository) scan(row pgx.Row) (*model.ToolRun, error) {
	var run model.ToolRun
	var status string
	err := row.Scan(
		&run.ID, &run.UserID, &run.ToolSlug, &run.PlanTier,
		&run.Input, &run.Output, &run.ArtifactURL, &run.TokensUsed, &run.ModelUsed,
		&status, &run.ErrorMsg, &run.CreatedAt, &run.UpdatedAt, &run.CompletedAt,
	)
	run.Status = model.RunStatus(status)
	return &run, err
}

func currentPeriodMonth() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func (r *ToolRunRepository) IncrementUsageCounter(ctx context.Context, userID, toolSlug string) error {
	period := currentPeriodMonth()
	const q = `
		INSERT INTO tool_usage_counters (user_id, tool_slug, period_month, runs_count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (user_id, tool_slug, period_month)
		DO UPDATE SET runs_count = tool_usage_counters.runs_count + 1
	`
	_, err := r.pool.Exec(ctx, q, userID, toolSlug, period)
	return err
}

func (r *ToolRunRepository) GetUsageCount(ctx context.Context, userID, toolSlug string) (int, error) {
	period := currentPeriodMonth()
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT runs_count FROM tool_usage_counters WHERE user_id = $1 AND tool_slug = $2 AND period_month = $3`,
		userID, toolSlug, period,
	).Scan(&count)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return count, err
}
