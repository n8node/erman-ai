package repository

import (
	"context"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenPackageCheckoutRepository struct {
	pool *pgxpool.Pool
}

func NewTokenPackageCheckoutRepository(pool *pgxpool.Pool) *TokenPackageCheckoutRepository {
	return &TokenPackageCheckoutRepository{pool: pool}
}

func (r *TokenPackageCheckoutRepository) Create(
	ctx context.Context,
	userID, packageID, provider string,
	amountRub int,
	tokensAmount int64,
) (*model.TokenPackageCheckout, error) {
	const q = `
		INSERT INTO token_package_checkouts (user_id, token_package_id, provider, amount_rub, tokens_amount, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		RETURNING id, user_id, token_package_id, provider, amount_rub, tokens_amount, status, external_id, inv_id, created_at, paid_at
	`
	return r.scan(r.pool.QueryRow(ctx, q, userID, packageID, provider, amountRub, tokensAmount))
}

func (r *TokenPackageCheckoutRepository) SetExternal(ctx context.Context, id, externalID string, invID *int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE token_package_checkouts SET external_id = $2, inv_id = $3 WHERE id = $1
	`, id, externalID, invID)
	return err
}

func (r *TokenPackageCheckoutRepository) GetByID(ctx context.Context, id string) (*model.TokenPackageCheckout, error) {
	const q = `
		SELECT id, user_id, token_package_id, provider, amount_rub, tokens_amount, status, external_id, inv_id, created_at, paid_at
		FROM token_package_checkouts WHERE id = $1
	`
	c, err := r.scan(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *TokenPackageCheckoutRepository) GetByExternal(ctx context.Context, provider, externalID string) (*model.TokenPackageCheckout, error) {
	const q = `
		SELECT id, user_id, token_package_id, provider, amount_rub, tokens_amount, status, external_id, inv_id, created_at, paid_at
		FROM token_package_checkouts WHERE provider = $1 AND external_id = $2
	`
	c, err := r.scan(r.pool.QueryRow(ctx, q, provider, externalID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *TokenPackageCheckoutRepository) GetByInvID(ctx context.Context, invID int64) (*model.TokenPackageCheckout, error) {
	const q = `
		SELECT id, user_id, token_package_id, provider, amount_rub, tokens_amount, status, external_id, inv_id, created_at, paid_at
		FROM token_package_checkouts WHERE inv_id = $1
	`
	c, err := r.scan(r.pool.QueryRow(ctx, q, invID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *TokenPackageCheckoutRepository) MarkPaid(ctx context.Context, id string) (*model.TokenPackageCheckout, error) {
	const q = `
		UPDATE token_package_checkouts
		SET status = 'paid', paid_at = NOW()
		WHERE id = $1 AND status = 'pending'
		RETURNING id, user_id, token_package_id, provider, amount_rub, tokens_amount, status, external_id, inv_id, created_at, paid_at
	`
	c, err := r.scan(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return r.GetByID(ctx, id)
	}
	return c, err
}

func (r *TokenPackageCheckoutRepository) scan(row pgx.Row) (*model.TokenPackageCheckout, error) {
	var c model.TokenPackageCheckout
	err := row.Scan(
		&c.ID, &c.UserID, &c.TokenPackageID, &c.Provider, &c.AmountRUB, &c.TokensAmount,
		&c.Status, &c.ExternalID, &c.InvID, &c.CreatedAt, &c.PaidAt,
	)
	return &c, err
}
