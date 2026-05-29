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

type ProjectInquiryRepository struct {
	pool *pgxpool.Pool
}

func NewProjectInquiryRepository(pool *pgxpool.Pool) *ProjectInquiryRepository {
	return &ProjectInquiryRepository{pool: pool}
}

type AdminProjectInquiryRow struct {
	ID                 string    `json:"id"`
	UserID             *string   `json:"user_id,omitempty"`
	UserEmail          *string   `json:"user_email,omitempty"`
	CalculatorRunID    *string   `json:"calculator_run_id,omitempty"`
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	Telegram           string    `json:"telegram"`
	ProjectTitle       string    `json:"project_title"`
	ProjectDescription string    `json:"project_description"`
	ProcessName        string    `json:"process_name,omitempty"`
	NetBenefitMonthly  *float64  `json:"net_benefit_monthly,omitempty"`
	PaybackMonths      *float64  `json:"payback_months,omitempty"`
	Recommendation     *string   `json:"recommendation,omitempty"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type AdminProjectInquiryDetail struct {
	AdminProjectInquiryRow
	Input  json.RawMessage `json:"input,omitempty"`
	Output json.RawMessage `json:"output,omitempty"`
}

func (r *ProjectInquiryRepository) Create(
	ctx context.Context,
	userID *string,
	runID *string,
	name, email, telegram, projectTitle, projectDescription, status, locale, ipHash string,
) (*model.ProjectInquiry, error) {
	const q = `
		INSERT INTO project_inquiries (
			user_id, calculator_run_id, name, email, telegram,
			project_title, project_description, status, locale, ip_hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''))
		RETURNING id, user_id, calculator_run_id, name, email, telegram,
		          project_title, project_description, status, email_verified_at,
		          locale, created_at, updated_at
	`
	return r.scan(r.pool.QueryRow(ctx, q,
		userID, runID, name, email, telegram,
		projectTitle, projectDescription, status, locale, ipHash,
	))
}

func (r *ProjectInquiryRepository) GetByID(ctx context.Context, id string) (*model.ProjectInquiry, error) {
	const q = `
		SELECT id, user_id, calculator_run_id, name, email, telegram,
		       project_title, project_description, status, email_verified_at,
		       locale, created_at, updated_at
		FROM project_inquiries WHERE id = $1
	`
	inq, err := r.scan(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return inq, err
}

func (r *ProjectInquiryRepository) MarkVerified(ctx context.Context, id string) (*model.ProjectInquiry, error) {
	const q = `
		UPDATE project_inquiries
		SET status = 'new', email_verified_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'pending_email'
		RETURNING id, user_id, calculator_run_id, name, email, telegram,
		          project_title, project_description, status, email_verified_at,
		          locale, created_at, updated_at
	`
	inq, err := r.scan(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return inq, err
}

func (r *ProjectInquiryRepository) ListAdmin(ctx context.Context, limit, offset int) ([]AdminProjectInquiryRow, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
		SELECT pi.id, pi.user_id, u.email, pi.calculator_run_id,
		       pi.name, pi.email, pi.telegram, pi.project_title, pi.project_description,
		       COALESCE(tr.input->>'process_name', '') AS process_name,
		       NULLIF(tr.output->>'net_benefit_monthly', '')::double precision,
		       NULLIF(tr.output->>'payback_months', '')::double precision,
		       NULLIF(tr.output->>'recommendation', ''),
		       pi.status, pi.created_at, pi.updated_at
		FROM project_inquiries pi
		LEFT JOIN users u ON u.id = pi.user_id
		LEFT JOIN tool_runs tr ON tr.id = pi.calculator_run_id
		ORDER BY pi.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AdminProjectInquiryRow
	for rows.Next() {
		var item AdminProjectInquiryRow
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.UserEmail, &item.CalculatorRunID,
			&item.Name, &item.Email, &item.Telegram, &item.ProjectTitle, &item.ProjectDescription,
			&item.ProcessName, &item.NetBenefitMonthly, &item.PaybackMonths, &item.Recommendation,
			&item.Status, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ProjectInquiryRepository) GetAdminDetail(ctx context.Context, id string) (*AdminProjectInquiryDetail, error) {
	const q = `
		SELECT pi.id, pi.user_id, u.email, pi.calculator_run_id,
		       pi.name, pi.email, pi.telegram, pi.project_title, pi.project_description,
		       COALESCE(tr.input->>'process_name', '') AS process_name,
		       NULLIF(tr.output->>'net_benefit_monthly', '')::double precision,
		       NULLIF(tr.output->>'payback_months', '')::double precision,
		       NULLIF(tr.output->>'recommendation', ''),
		       pi.status, pi.created_at, pi.updated_at,
		       tr.input, tr.output
		FROM project_inquiries pi
		LEFT JOIN users u ON u.id = pi.user_id
		LEFT JOIN tool_runs tr ON tr.id = pi.calculator_run_id
		WHERE pi.id = $1
	`
	var item AdminProjectInquiryDetail
	var input, output []byte
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&item.ID, &item.UserID, &item.UserEmail, &item.CalculatorRunID,
		&item.Name, &item.Email, &item.Telegram, &item.ProjectTitle, &item.ProjectDescription,
		&item.ProcessName, &item.NetBenefitMonthly, &item.PaybackMonths, &item.Recommendation,
		&item.Status, &item.CreatedAt, &item.UpdatedAt,
		&input, &output,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(input) > 0 {
		item.Input = input
	}
	if len(output) > 0 {
		item.Output = output
	}
	return &item, nil
}

func (r *ProjectInquiryRepository) CountAdmin(ctx context.Context) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM project_inquiries`).Scan(&total)
	return total, err
}

func (r *ProjectInquiryRepository) UpdateStatus(ctx context.Context, id, status string) (*model.ProjectInquiry, error) {
	const q = `
		UPDATE project_inquiries SET status = $2, updated_at = NOW() WHERE id = $1
		RETURNING id, user_id, calculator_run_id, name, email, telegram,
		          project_title, project_description, status, email_verified_at,
		          locale, created_at, updated_at
	`
	inq, err := r.scan(r.pool.QueryRow(ctx, q, id, status))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return inq, err
}

func (r *ProjectInquiryRepository) scan(row pgx.Row) (*model.ProjectInquiry, error) {
	var inq model.ProjectInquiry
	err := row.Scan(
		&inq.ID, &inq.UserID, &inq.CalculatorRunID,
		&inq.Name, &inq.Email, &inq.Telegram,
		&inq.ProjectTitle, &inq.ProjectDescription, &inq.Status, &inq.EmailVerifiedAt,
		&inq.Locale, &inq.CreatedAt, &inq.UpdatedAt,
	)
	return &inq, err
}
