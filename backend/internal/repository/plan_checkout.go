package repository

import (
	"context"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlanCheckoutRepository struct {
	pool *pgxpool.Pool
}

func NewPlanCheckoutRepository(pool *pgxpool.Pool) *PlanCheckoutRepository {
	return &PlanCheckoutRepository{pool: pool}
}

func (r *PlanCheckoutRepository) Create(ctx context.Context, userID, planID, provider string, amountRub int) (*model.PlanCheckout, error) {
	const q = `
		INSERT INTO plan_checkouts (user_id, plan_id, provider, amount_rub, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, user_id, plan_id, provider, amount_rub, status, external_id, inv_id, created_at, paid_at
	`
	return r.scan(r.pool.QueryRow(ctx, q, userID, planID, provider, amountRub))
}

func (r *PlanCheckoutRepository) NextInvID(ctx context.Context) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `SELECT nextval('plan_checkout_inv_id_seq')`).Scan(&id)
	return id, err
}

func (r *PlanCheckoutRepository) SetExternal(ctx context.Context, id string, externalID string, invID *int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE plan_checkouts SET external_id = $2, inv_id = $3 WHERE id = $1
	`, id, externalID, invID)
	return err
}

func (r *PlanCheckoutRepository) GetByID(ctx context.Context, id string) (*model.PlanCheckout, error) {
	const q = `
		SELECT id, user_id, plan_id, provider, amount_rub, status, external_id, inv_id, created_at, paid_at
		FROM plan_checkouts WHERE id = $1
	`
	c, err := r.scan(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *PlanCheckoutRepository) GetByExternal(ctx context.Context, provider, externalID string) (*model.PlanCheckout, error) {
	const q = `
		SELECT id, user_id, plan_id, provider, amount_rub, status, external_id, inv_id, created_at, paid_at
		FROM plan_checkouts WHERE provider = $1 AND external_id = $2
	`
	c, err := r.scan(r.pool.QueryRow(ctx, q, provider, externalID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *PlanCheckoutRepository) GetByInvID(ctx context.Context, invID int64) (*model.PlanCheckout, error) {
	const q = `
		SELECT id, user_id, plan_id, provider, amount_rub, status, external_id, inv_id, created_at, paid_at
		FROM plan_checkouts WHERE inv_id = $1
	`
	c, err := r.scan(r.pool.QueryRow(ctx, q, invID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *PlanCheckoutRepository) MarkPaid(ctx context.Context, id string) (*model.PlanCheckout, error) {
	const q = `
		UPDATE plan_checkouts
		SET status = 'paid', paid_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id, user_id, plan_id, provider, amount_rub, status, external_id, inv_id, created_at, paid_at
	`
	c, err := r.scan(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return r.GetByID(ctx, id)
	}
	return c, err
}

func (r *PlanCheckoutRepository) CreateSubscription(ctx context.Context, userID, planID string, amountRub int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscriptions (user_id, plan_id, amount_rub, status)
		VALUES ($1, $2, $3, 'active')
	`, userID, planID, amountRub)
	return err
}

func (r *PlanCheckoutRepository) scan(row pgx.Row) (*model.PlanCheckout, error) {
	var c model.PlanCheckout
	err := row.Scan(
		&c.ID, &c.UserID, &c.PlanID, &c.Provider, &c.AmountRUB, &c.Status,
		&c.ExternalID, &c.InvID, &c.CreatedAt, &c.PaidAt,
	)
	return &c, err
}
