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
	ShareToken *string         `json:"share_token,omitempty"`
	Input      json.RawMessage `json:"input,omitempty"`
	Output     json.RawMessage `json:"output,omitempty"`
}

type InquiryShareAttachRow struct {
	ID                 string
	UserID             *string
	CalculatorRunID    *string
	CalculatorSnapshot json.RawMessage
	ShareToken         *string
}

func (r *ProjectInquiryRepository) Create(
	ctx context.Context,
	userID *string,
	runID *string,
	name, email, telegram, projectTitle, projectDescription, status, locale, ipHash string,
	calculatorSnapshot json.RawMessage,
) (*model.ProjectInquiry, error) {
	const q = `
		INSERT INTO project_inquiries (
			user_id, calculator_run_id, name, email, telegram,
			project_title, project_description, status, locale, ip_hash,
			calculator_snapshot
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), $11)
		RETURNING id, user_id, calculator_run_id, name, email, telegram,
		          project_title, project_description, status, email_verified_at,
		          locale, created_at, updated_at
	`
	return r.scan(r.pool.QueryRow(ctx, q,
		userID, runID, name, email, telegram,
		projectTitle, projectDescription, status, locale, ipHash,
		nullJSON(calculatorSnapshot),
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
		       COALESCE(tr.input->>'process_name', pi.calculator_snapshot->'input'->>'process_name', '') AS process_name,
		       NULLIF(COALESCE(tr.output->>'net_benefit_monthly', pi.calculator_snapshot->'output'->>'net_benefit_monthly'), '')::double precision,
		       NULLIF(COALESCE(tr.output->>'payback_months', pi.calculator_snapshot->'output'->>'payback_months'), '')::double precision,
		       NULLIF(COALESCE(tr.output->>'recommendation', pi.calculator_snapshot->'output'->>'recommendation'), ''),
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

func (r *ProjectInquiryRepository) GetForShareAttach(ctx context.Context, id string) (*InquiryShareAttachRow, error) {
	const q = `
		SELECT id, user_id, calculator_run_id, calculator_snapshot, share_token
		FROM project_inquiries
		WHERE id = $1
	`
	var row InquiryShareAttachRow
	var snapshot []byte
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&row.ID, &row.UserID, &row.CalculatorRunID, &snapshot, &row.ShareToken,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if len(snapshot) > 0 {
		row.CalculatorSnapshot = snapshot
	}
	return &row, nil
}

func (r *ProjectInquiryRepository) SetShareAttachment(ctx context.Context, id string, runID *string, shareToken string) error {
	const q = `
		UPDATE project_inquiries
		SET calculator_run_id = COALESCE($2, calculator_run_id),
		    share_token = $3,
		    updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, q, id, runID, shareToken)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProjectInquiryRepository) GetAdminDetail(ctx context.Context, id string) (*AdminProjectInquiryDetail, error) {
	const q = `
		SELECT pi.id, pi.user_id, u.email, pi.calculator_run_id,
		       pi.name, pi.email, pi.telegram, pi.project_title, pi.project_description,
		       COALESCE(tr.input->>'process_name', pi.calculator_snapshot->'input'->>'process_name', '') AS process_name,
		       NULLIF(COALESCE(tr.output->>'net_benefit_monthly', pi.calculator_snapshot->'output'->>'net_benefit_monthly'), '')::double precision,
		       NULLIF(COALESCE(tr.output->>'payback_months', pi.calculator_snapshot->'output'->>'payback_months'), '')::double precision,
		       NULLIF(COALESCE(tr.output->>'recommendation', pi.calculator_snapshot->'output'->>'recommendation'), ''),
		       pi.status, pi.created_at, pi.updated_at, pi.share_token,
		       COALESCE(tr.input, pi.calculator_snapshot->'input'),
		       COALESCE(tr.output, pi.calculator_snapshot->'output')
		FROM project_inquiries pi
		LEFT JOIN users u ON u.id = pi.user_id
		LEFT JOIN shared_reports sr ON sr.token = pi.share_token
		LEFT JOIN tool_runs tr ON tr.id = COALESCE(pi.calculator_run_id, sr.run_id)
		WHERE pi.id = $1
	`
	var item AdminProjectInquiryDetail
	var input, output []byte
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&item.ID, &item.UserID, &item.UserEmail, &item.CalculatorRunID,
		&item.Name, &item.Email, &item.Telegram, &item.ProjectTitle, &item.ProjectDescription,
		&item.ProcessName, &item.NetBenefitMonthly, &item.PaybackMonths, &item.Recommendation,
		&item.Status, &item.CreatedAt, &item.UpdatedAt, &item.ShareToken,
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

func (r *ProjectInquiryRepository) ExistsByUserID(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM project_inquiries WHERE user_id = $1)`, userID).Scan(&exists)
	return exists, err
}

func (r *ProjectInquiryRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM project_inquiries WHERE LOWER(email) = LOWER($1))`,
		email,
	).Scan(&exists)
	return exists, err
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

func nullJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
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
