package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlanRepository struct {
	pool *pgxpool.Pool
}

func NewPlanRepository(pool *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{pool: pool}
}

type UserPlan struct {
	Plan     model.Plan
	PlanSlug string
}

func (r *PlanRepository) GetUserPlan(ctx context.Context, userID string) (*UserPlan, error) {
	const q = `
		SELECT p.id, p.slug, p.name, p.price_monthly_rub, p.price_yearly_rub, p.tool_limits, p.features, p.support_level, p.is_public, p.is_archived
		FROM users u
		JOIN plans p ON p.id = u.plan_id
		WHERE u.id = $1
	`
	var p model.Plan
	var features, limits []byte
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.ID, &p.Slug, &p.Name, &p.PriceMonthlyRUB, &p.PriceYearlyRUB,
		&limits, &features, &p.SupportLevel, &p.IsPublic, &p.IsArchived,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return r.defaultFreePlan(ctx)
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(limits, &p.ToolLimits)
	_ = json.Unmarshal(features, &p.Features)
	return &UserPlan{Plan: p, PlanSlug: p.Slug}, nil
}

func (r *PlanRepository) defaultFreePlan(ctx context.Context) (*UserPlan, error) {
	const q = `
		SELECT id, slug, name, price_monthly_rub, price_yearly_rub, tool_limits, features, support_level, is_public, is_archived
		FROM plans WHERE slug = 'free' LIMIT 1
	`
	var p model.Plan
	var features, limits []byte
	err := r.pool.QueryRow(ctx, q).Scan(
		&p.ID, &p.Slug, &p.Name, &p.PriceMonthlyRUB, &p.PriceYearlyRUB,
		&limits, &features, &p.SupportLevel, &p.IsPublic, &p.IsArchived,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(limits, &p.ToolLimits)
	_ = json.Unmarshal(features, &p.Features)
	return &UserPlan{Plan: p, PlanSlug: p.Slug}, nil
}

func (r *PlanRepository) ListTools(ctx context.Context) ([]model.ToolConfig, error) {
	const q = `SELECT slug, name, description, enabled FROM tools WHERE enabled = true ORDER BY slug`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []model.ToolConfig
	for rows.Next() {
		var t model.ToolConfig
		var desc *string
		if err := rows.Scan(&t.Slug, &t.Name, &desc, &t.Enabled); err != nil {
			return nil, err
		}
		if desc != nil {
			t.Description = *desc
		}
		tools = append(tools, t)
	}
	return tools, rows.Err()
}
