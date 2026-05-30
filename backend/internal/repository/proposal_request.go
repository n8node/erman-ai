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

type ProposalRequestRepository struct {
	pool *pgxpool.Pool
}

func NewProposalRequestRepository(pool *pgxpool.Pool) *ProposalRequestRepository {
	return &ProposalRequestRepository{pool: pool}
}

type AdminProposalRequestRow struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	UserEmail         string    `json:"user_email"`
	RunID             string    `json:"run_id"`
	RequesterName     string    `json:"requester_name"`
	Telegram          string    `json:"telegram"`
	BusinessNote      string    `json:"business_note"`
	ProcessName       string    `json:"process_name"`
	NetBenefitMonthly *float64  `json:"net_benefit_monthly,omitempty"`
	PaybackMonths     *float64  `json:"payback_months,omitempty"`
	Recommendation    *string   `json:"recommendation,omitempty"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type AdminProposalRequestDetail struct {
	AdminProposalRequestRow
	Input  json.RawMessage `json:"input"`
	Output json.RawMessage `json:"output"`
}

func (r *ProposalRequestRepository) Create(
	ctx context.Context,
	userID, runID, requesterName, telegram, businessNote string,
) (*model.ProposalRequest, error) {
	const q = `
		INSERT INTO proposal_requests (user_id, run_id, requester_name, telegram, business_note)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, run_id, requester_name, telegram, business_note, status, created_at, updated_at
	`
	return r.scan(r.pool.QueryRow(ctx, q, userID, runID, requesterName, telegram, businessNote))
}

func (r *ProposalRequestRepository) ListAdmin(ctx context.Context, limit, offset int) ([]AdminProposalRequestRow, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `
		SELECT pr.id, pr.user_id, u.email, pr.run_id,
		       pr.requester_name, pr.telegram, pr.business_note,
		       COALESCE(tr.input->>'process_name', '') AS process_name,
		       NULLIF(tr.output->>'net_benefit_monthly', '')::double precision,
		       NULLIF(tr.output->>'payback_months', '')::double precision,
		       NULLIF(tr.output->>'recommendation', ''),
		       pr.status, pr.created_at, pr.updated_at
		FROM proposal_requests pr
		JOIN users u ON u.id = pr.user_id
		JOIN tool_runs tr ON tr.id = pr.run_id
		ORDER BY pr.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AdminProposalRequestRow
	for rows.Next() {
		var item AdminProposalRequestRow
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.UserEmail, &item.RunID,
			&item.RequesterName, &item.Telegram, &item.BusinessNote,
			&item.ProcessName, &item.NetBenefitMonthly, &item.PaybackMonths, &item.Recommendation,
			&item.Status, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ProposalRequestRepository) GetAdminDetail(ctx context.Context, id string) (*AdminProposalRequestDetail, error) {
	const q = `
		SELECT pr.id, pr.user_id, u.email, pr.run_id,
		       pr.requester_name, pr.telegram, pr.business_note,
		       COALESCE(tr.input->>'process_name', '') AS process_name,
		       NULLIF(tr.output->>'net_benefit_monthly', '')::double precision,
		       NULLIF(tr.output->>'payback_months', '')::double precision,
		       NULLIF(tr.output->>'recommendation', ''),
		       pr.status, pr.created_at, pr.updated_at,
		       tr.input, tr.output
		FROM proposal_requests pr
		JOIN users u ON u.id = pr.user_id
		JOIN tool_runs tr ON tr.id = pr.run_id
		WHERE pr.id = $1
	`
	var item AdminProposalRequestDetail
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&item.ID, &item.UserID, &item.UserEmail, &item.RunID,
		&item.RequesterName, &item.Telegram, &item.BusinessNote,
		&item.ProcessName, &item.NetBenefitMonthly, &item.PaybackMonths, &item.Recommendation,
		&item.Status, &item.CreatedAt, &item.UpdatedAt,
		&item.Input, &item.Output,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &item, err
}

func (r *ProposalRequestRepository) ExistsByUserID(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM proposal_requests WHERE user_id = $1)`, userID).Scan(&exists)
	return exists, err
}

func (r *ProposalRequestRepository) CountAdmin(ctx context.Context) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM proposal_requests`).Scan(&total)
	return total, err
}

func (r *ProposalRequestRepository) UpdateStatus(ctx context.Context, id, status string) (*model.ProposalRequest, error) {
	const q = `
		UPDATE proposal_requests
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, run_id, requester_name, telegram, business_note, status, created_at, updated_at
	`
	req, err := r.scan(r.pool.QueryRow(ctx, q, id, status))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return req, err
}

func (r *ProposalRequestRepository) scan(row pgx.Row) (*model.ProposalRequest, error) {
	var req model.ProposalRequest
	err := row.Scan(
		&req.ID, &req.UserID, &req.RunID,
		&req.RequesterName, &req.Telegram, &req.BusinessNote,
		&req.Status, &req.CreatedAt, &req.UpdatedAt,
	)
	return &req, err
}
