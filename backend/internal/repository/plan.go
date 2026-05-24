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

type AdminPlanRow struct {
	Plan      model.Plan
	UserCount int
}

func (r *PlanRepository) scanPlan(row pgx.Row) (*model.Plan, error) {
	var p model.Plan
	var features, limits []byte
	err := row.Scan(
		&p.ID, &p.Slug, &p.Name, &p.PriceMonthlyRUB, &p.PriceYearlyRUB,
		&limits, &features, &p.SupportLevel, &p.IsPublic, &p.IsArchived,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(limits, &p.ToolLimits)
	_ = json.Unmarshal(features, &p.Features)
	if p.ToolLimits == nil {
		p.ToolLimits = map[string]int{}
	}
	if p.Features == nil {
		p.Features = map[string]any{}
	}
	return &p, nil
}

func (r *PlanRepository) ListAll(ctx context.Context) ([]AdminPlanRow, error) {
	const q = `
		SELECT p.id, p.slug, p.name, p.price_monthly_rub, p.price_yearly_rub,
		       p.tool_limits, p.features, p.support_level, p.is_public, p.is_archived,
		       COUNT(u.id)::int AS user_count
		FROM plans p
		LEFT JOIN users u ON u.plan_id = p.id
		GROUP BY p.id
		ORDER BY p.price_monthly_rub ASC, p.name ASC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AdminPlanRow
	for rows.Next() {
		var p model.Plan
		var features, limits []byte
		var userCount int
		if err := rows.Scan(
			&p.ID, &p.Slug, &p.Name, &p.PriceMonthlyRUB, &p.PriceYearlyRUB,
			&limits, &features, &p.SupportLevel, &p.IsPublic, &p.IsArchived,
			&userCount,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(limits, &p.ToolLimits)
		_ = json.Unmarshal(features, &p.Features)
		if p.ToolLimits == nil {
			p.ToolLimits = map[string]int{}
		}
		if p.Features == nil {
			p.Features = map[string]any{}
		}
		out = append(out, AdminPlanRow{Plan: p, UserCount: userCount})
	}
	return out, rows.Err()
}

func (r *PlanRepository) GetByID(ctx context.Context, id string) (*model.Plan, error) {
	const q = `
		SELECT id, slug, name, price_monthly_rub, price_yearly_rub, tool_limits, features, support_level, is_public, is_archived
		FROM plans WHERE id = $1
	`
	p, err := r.scanPlan(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PlanRepository) SlugExists(ctx context.Context, slug, excludeID string) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1 FROM plans WHERE slug = $1 AND ($2 = '' OR id != $2::uuid)
		)
	`
	var exists bool
	err := r.pool.QueryRow(ctx, q, slug, excludeID).Scan(&exists)
	return exists, err
}

func (r *PlanRepository) Create(ctx context.Context, p *model.Plan) (*model.Plan, error) {
	limits, _ := json.Marshal(p.ToolLimits)
	features, _ := json.Marshal(p.Features)
	const q = `
		INSERT INTO plans (slug, name, price_monthly_rub, price_yearly_rub, tool_limits, features, support_level, is_public, is_archived)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, slug, name, price_monthly_rub, price_yearly_rub, tool_limits, features, support_level, is_public, is_archived
	`
	return r.scanPlan(r.pool.QueryRow(ctx, q,
		p.Slug, p.Name, p.PriceMonthlyRUB, p.PriceYearlyRUB,
		limits, features, p.SupportLevel, p.IsPublic, p.IsArchived,
	))
}

func (r *PlanRepository) Update(ctx context.Context, id string, p *model.Plan) (*model.Plan, error) {
	limits, _ := json.Marshal(p.ToolLimits)
	features, _ := json.Marshal(p.Features)
	const q = `
		UPDATE plans SET
			name = $2,
			price_monthly_rub = $3,
			price_yearly_rub = $4,
			tool_limits = $5,
			features = $6,
			support_level = $7,
			is_public = $8,
			is_archived = $9,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, slug, name, price_monthly_rub, price_yearly_rub, tool_limits, features, support_level, is_public, is_archived
	`
	plan, err := r.scanPlan(r.pool.QueryRow(ctx, q, id,
		p.Name, p.PriceMonthlyRUB, p.PriceYearlyRUB,
		limits, features, p.SupportLevel, p.IsPublic, p.IsArchived,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return plan, err
}

func (r *PlanRepository) ListAllTools(ctx context.Context) ([]model.ToolConfig, error) {
	const q = `SELECT slug, name, description, enabled FROM tools ORDER BY slug`
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
