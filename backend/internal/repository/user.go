package repository

import (
	"context"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

const userColumns = `id, email, role, plan_id, locale, account_segment, onboarding_completed, is_blocked, created_at, updated_at, last_active_at`

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string, planID *string, role, segment string, onboardingDone bool) (*model.User, error) {
	if role == "" {
		role = "user"
	}
	if segment == "" {
		segment = model.AccountSegmentPartner
	}
	const q = `
		INSERT INTO users (email, password_hash, role, plan_id, account_segment, onboarding_completed)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + userColumns
	return r.scanUser(r.pool.QueryRow(ctx, q, email, passwordHash, role, planID, segment, onboardingDone))
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, string, error) {
	const q = `
		SELECT id, email, password_hash, role, plan_id, locale, account_segment, onboarding_completed, is_blocked, created_at, updated_at, last_active_at
		FROM users WHERE email = $1
	`
	var hash string
	row := r.pool.QueryRow(ctx, q, email)
	u, err := r.scanUserRow(row, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrNotFound
	}
	return u, hash, err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	const q = `SELECT id, email, password_hash, role, plan_id, locale, account_segment, onboarding_completed, is_blocked, created_at, updated_at, last_active_at FROM users WHERE id = $1`
	var hash string
	row := r.pool.QueryRow(ctx, q, id)
	u, err := r.scanUserRow(row, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id, email, locale string) (*model.User, error) {
	const q = `
		UPDATE users SET email = $2, locale = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING ` + userColumns
	return r.scanUser(r.pool.QueryRow(ctx, q, id, email, locale))
}

func (r *UserRepository) CompleteOnboarding(ctx context.Context, id, segment string) (*model.User, error) {
	const q = `
		UPDATE users SET account_segment = $2, onboarding_completed = true, updated_at = NOW()
		WHERE id = $1
		RETURNING ` + userColumns
	return r.scanUser(r.pool.QueryRow(ctx, q, id, segment))
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, id, passwordHash)
	return err
}

func (r *UserRepository) TouchLastActive(ctx context.Context, id string) {
	_, _ = r.pool.Exec(ctx, `UPDATE users SET last_active_at = NOW() WHERE id = $1`, id)
}

func (r *UserRepository) GetPlanIDBySlug(ctx context.Context, slug string) (*string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT id FROM plans WHERE slug = $1 LIMIT 1`, slug).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (r *UserRepository) scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	err := row.Scan(
		&u.ID, &u.Email, &u.Role, &u.PlanID, &u.Locale,
		&u.AccountSegment, &u.OnboardingCompleted, &u.IsBlocked,
		&u.CreatedAt, &u.UpdatedAt, &u.LastActiveAt,
	)
	return &u, err
}

func (r *UserRepository) scanUserRow(row pgx.Row, hash *string) (*model.User, error) {
	var u model.User
	err := row.Scan(
		&u.ID, &u.Email, hash, &u.Role, &u.PlanID, &u.Locale,
		&u.AccountSegment, &u.OnboardingCompleted, &u.IsBlocked,
		&u.CreatedAt, &u.UpdatedAt, &u.LastActiveAt,
	)
	return &u, err
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT 1 FROM users WHERE email = $1`, email).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *UserRepository) GetPasswordHashByID(ctx context.Context, id string) (string, error) {
	var hash string
	err := r.pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, id).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return hash, err
}

func (r *UserRepository) SetRole(ctx context.Context, id, role string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`, id, role)
	return err
}
