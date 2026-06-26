package repository

import (
	"context"
	"errors"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConsultationRepository struct {
	pool *pgxpool.Pool
}

func NewConsultationRepository(pool *pgxpool.Pool) *ConsultationRepository {
	return &ConsultationRepository{pool: pool}
}

type ConsultationServiceInput struct {
	Slug                string `json:"slug"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	DurationMinutes     int    `json:"duration_minutes"`
	PriceRUB            int    `json:"price_rub"`
	IsActive            bool   `json:"is_active"`
	MinNoticeMinutes    int    `json:"min_notice_minutes"`
	MaxAdvanceDays      int    `json:"max_advance_days"`
	BufferBeforeMinutes int    `json:"buffer_before_minutes"`
	BufferAfterMinutes  int    `json:"buffer_after_minutes"`
	MeetingURL          string `json:"meeting_url"`
	SortOrder           int    `json:"sort_order"`
}

type ConsultationAvailabilityInput struct {
	Weekday         int    `json:"weekday"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	SlotStepMinutes int    `json:"slot_step_minutes"`
	IsActive        bool   `json:"is_active"`
}

type ConsultationBookingInput struct {
	ServiceID        string
	UserID           *string
	CustomerName     string
	CustomerEmail    string
	CustomerPhone    string
	CustomerTelegram string
	CustomerNote     string
	StartsAt         time.Time
	EndsAt           time.Time
	Timezone         string
	AmountRUB        int
	MeetingURL       string
	ExpiresAt        time.Time
}

func (r *ConsultationRepository) ListServices(ctx context.Context, includeInactive bool) ([]model.ConsultationService, error) {
	q := `
		SELECT id, slug, name, description, duration_minutes, price_rub, is_active,
			min_notice_minutes, max_advance_days, buffer_before_minutes, buffer_after_minutes,
			meeting_url, sort_order, created_at, updated_at
		FROM consultation_services
		WHERE ($1 OR is_active)
		ORDER BY sort_order ASC, created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, includeInactive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.ConsultationService{}
	for rows.Next() {
		item, err := scanConsultationService(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *ConsultationRepository) GetService(ctx context.Context, id string) (*model.ConsultationService, error) {
	q := `
		SELECT id, slug, name, description, duration_minutes, price_rub, is_active,
			min_notice_minutes, max_advance_days, buffer_before_minutes, buffer_after_minutes,
			meeting_url, sort_order, created_at, updated_at
		FROM consultation_services
		WHERE id = $1
	`
	item, err := scanConsultationService(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *ConsultationRepository) GetServiceBySlug(ctx context.Context, slug string) (*model.ConsultationService, error) {
	q := `
		SELECT id, slug, name, description, duration_minutes, price_rub, is_active,
			min_notice_minutes, max_advance_days, buffer_before_minutes, buffer_after_minutes,
			meeting_url, sort_order, created_at, updated_at
		FROM consultation_services
		WHERE slug = $1
	`
	item, err := scanConsultationService(r.pool.QueryRow(ctx, q, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *ConsultationRepository) CreateService(ctx context.Context, in ConsultationServiceInput) (*model.ConsultationService, error) {
	q := `
		INSERT INTO consultation_services (
			slug, name, description, duration_minutes, price_rub, is_active,
			min_notice_minutes, max_advance_days, buffer_before_minutes, buffer_after_minutes,
			meeting_url, sort_order
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, slug, name, description, duration_minutes, price_rub, is_active,
			min_notice_minutes, max_advance_days, buffer_before_minutes, buffer_after_minutes,
			meeting_url, sort_order, created_at, updated_at
	`
	return scanConsultationService(r.pool.QueryRow(ctx, q,
		in.Slug, in.Name, in.Description, in.DurationMinutes, in.PriceRUB, in.IsActive,
		in.MinNoticeMinutes, in.MaxAdvanceDays, in.BufferBeforeMinutes, in.BufferAfterMinutes,
		in.MeetingURL, in.SortOrder,
	))
}

func (r *ConsultationRepository) UpdateService(ctx context.Context, id string, in ConsultationServiceInput) (*model.ConsultationService, error) {
	q := `
		UPDATE consultation_services SET
			slug = $2,
			name = $3,
			description = $4,
			duration_minutes = $5,
			price_rub = $6,
			is_active = $7,
			min_notice_minutes = $8,
			max_advance_days = $9,
			buffer_before_minutes = $10,
			buffer_after_minutes = $11,
			meeting_url = $12,
			sort_order = $13,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, slug, name, description, duration_minutes, price_rub, is_active,
			min_notice_minutes, max_advance_days, buffer_before_minutes, buffer_after_minutes,
			meeting_url, sort_order, created_at, updated_at
	`
	item, err := scanConsultationService(r.pool.QueryRow(ctx, q,
		id, in.Slug, in.Name, in.Description, in.DurationMinutes, in.PriceRUB, in.IsActive,
		in.MinNoticeMinutes, in.MaxAdvanceDays, in.BufferBeforeMinutes, in.BufferAfterMinutes,
		in.MeetingURL, in.SortOrder,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *ConsultationRepository) ListAvailability(ctx context.Context, serviceID string) ([]model.ConsultationAvailabilityRule, error) {
	q := `
		SELECT id, service_id, weekday, start_time::text, end_time::text, slot_step_minutes, is_active, created_at
		FROM consultation_availability_rules
		WHERE service_id = $1
		ORDER BY weekday ASC, start_time ASC
	`
	rows, err := r.pool.Query(ctx, q, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.ConsultationAvailabilityRule{}
	for rows.Next() {
		var item model.ConsultationAvailabilityRule
		if err := rows.Scan(&item.ID, &item.ServiceID, &item.Weekday, &item.StartTime, &item.EndTime, &item.SlotStepMinutes, &item.IsActive, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ConsultationRepository) ReplaceAvailability(ctx context.Context, serviceID string, rules []ConsultationAvailabilityInput) ([]model.ConsultationAvailabilityRule, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM consultation_availability_rules WHERE service_id = $1`, serviceID); err != nil {
		return nil, err
	}
	for _, rule := range rules {
		if _, err := tx.Exec(ctx, `
			INSERT INTO consultation_availability_rules (service_id, weekday, start_time, end_time, slot_step_minutes, is_active)
			VALUES ($1, $2, $3::time, $4::time, $5, $6)
		`, serviceID, rule.Weekday, rule.StartTime, rule.EndTime, rule.SlotStepMinutes, rule.IsActive); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.ListAvailability(ctx, serviceID)
}

func (r *ConsultationRepository) ListBlackouts(ctx context.Context, serviceID string, from, to time.Time) ([]model.ConsultationBlackout, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, service_id, starts_at, ends_at, reason, created_at
		FROM consultation_blackouts
		WHERE (service_id IS NULL OR service_id = $1)
			AND starts_at < $3
			AND ends_at > $2
		ORDER BY starts_at ASC
	`, serviceID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.ConsultationBlackout{}
	for rows.Next() {
		var item model.ConsultationBlackout
		if err := rows.Scan(&item.ID, &item.ServiceID, &item.StartsAt, &item.EndsAt, &item.Reason, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ConsultationRepository) CreateBooking(ctx context.Context, in ConsultationBookingInput) (*model.ConsultationBooking, error) {
	q := `
		INSERT INTO consultation_bookings (
			service_id, user_id, customer_name, customer_email, customer_phone,
			customer_telegram, customer_note, starts_at, ends_at, timezone, status,
			amount_rub, meeting_url, expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'pending_payment',$11,$12,$13)
		RETURNING id, service_id, user_id, customer_name, customer_email, customer_phone,
			customer_telegram, customer_note, starts_at, ends_at, timezone, status,
			amount_rub, provider, external_id, inv_id, meeting_url, expires_at, paid_at, created_at, updated_at
	`
	return scanConsultationBooking(r.pool.QueryRow(ctx, q,
		in.ServiceID, in.UserID, in.CustomerName, in.CustomerEmail, in.CustomerPhone,
		in.CustomerTelegram, in.CustomerNote, in.StartsAt, in.EndsAt, in.Timezone,
		in.AmountRUB, in.MeetingURL, in.ExpiresAt,
	))
}

func (r *ConsultationRepository) NextInvID(ctx context.Context) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `SELECT nextval('plan_checkout_inv_id_seq')`).Scan(&id)
	return id, err
}

func (r *ConsultationRepository) ListBookingsForService(ctx context.Context, serviceID string, from, to time.Time) ([]model.ConsultationBooking, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, service_id, user_id, customer_name, customer_email, customer_phone,
			customer_telegram, customer_note, starts_at, ends_at, timezone, status,
			amount_rub, provider, external_id, inv_id, meeting_url, expires_at, paid_at, created_at, updated_at
		FROM consultation_bookings
		WHERE service_id = $1
			AND (status = 'paid' OR (status = 'pending_payment' AND expires_at > NOW()))
			AND starts_at < $3
			AND ends_at > $2
	`, serviceID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.ConsultationBooking{}
	for rows.Next() {
		item, err := scanConsultationBooking(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *ConsultationRepository) ListBookingsAdmin(ctx context.Context, limit, offset int) ([]model.ConsultationBookingDetail, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM consultation_bookings`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT b.id, b.service_id, b.user_id, b.customer_name, b.customer_email, b.customer_phone,
			b.customer_telegram, b.customer_note, b.starts_at, b.ends_at, b.timezone, b.status,
			b.amount_rub, b.provider, b.external_id, b.inv_id, b.meeting_url, b.expires_at, b.paid_at,
			b.created_at, b.updated_at, s.name, s.slug, s.duration_minutes, s.description
		FROM consultation_bookings b
		JOIN consultation_services s ON s.id = b.service_id
		ORDER BY b.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []model.ConsultationBookingDetail{}
	for rows.Next() {
		item, err := scanConsultationBookingDetail(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (r *ConsultationRepository) GetBookingDetail(ctx context.Context, id string) (*model.ConsultationBookingDetail, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT b.id, b.service_id, b.user_id, b.customer_name, b.customer_email, b.customer_phone,
			b.customer_telegram, b.customer_note, b.starts_at, b.ends_at, b.timezone, b.status,
			b.amount_rub, b.provider, b.external_id, b.inv_id, b.meeting_url, b.expires_at, b.paid_at,
			b.created_at, b.updated_at, s.name, s.slug, s.duration_minutes, s.description
		FROM consultation_bookings b
		JOIN consultation_services s ON s.id = b.service_id
		WHERE b.id = $1
	`, id)
	item, err := scanConsultationBookingDetail(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *ConsultationRepository) GetBookingByExternal(ctx context.Context, provider, externalID string) (*model.ConsultationBooking, error) {
	item, err := scanConsultationBooking(r.pool.QueryRow(ctx, `
		SELECT id, service_id, user_id, customer_name, customer_email, customer_phone,
			customer_telegram, customer_note, starts_at, ends_at, timezone, status,
			amount_rub, provider, external_id, inv_id, meeting_url, expires_at, paid_at, created_at, updated_at
		FROM consultation_bookings
		WHERE provider = $1 AND external_id = $2
	`, provider, externalID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *ConsultationRepository) GetBookingByInvID(ctx context.Context, invID int64) (*model.ConsultationBooking, error) {
	item, err := scanConsultationBooking(r.pool.QueryRow(ctx, `
		SELECT id, service_id, user_id, customer_name, customer_email, customer_phone,
			customer_telegram, customer_note, starts_at, ends_at, timezone, status,
			amount_rub, provider, external_id, inv_id, meeting_url, expires_at, paid_at, created_at, updated_at
		FROM consultation_bookings
		WHERE inv_id = $1
	`, invID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *ConsultationRepository) SetBookingExternal(ctx context.Context, id, provider, externalID string, invID *int64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE consultation_bookings
		SET provider = $2, external_id = $3, inv_id = $4, updated_at = NOW()
		WHERE id = $1
	`, id, provider, externalID, invID)
	return err
}

func (r *ConsultationRepository) MarkBookingPaid(ctx context.Context, id string) (*model.ConsultationBookingDetail, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE consultation_bookings
		SET status = 'paid', paid_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'pending_payment'
	`, id)
	if err != nil {
		return nil, err
	}
	return r.GetBookingDetail(ctx, id)
}

func (r *ConsultationRepository) UpdateBookingStatus(ctx context.Context, id, status string) (*model.ConsultationBookingDetail, error) {
	_, err := r.pool.Exec(ctx, `
		UPDATE consultation_bookings
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, id, status)
	if err != nil {
		return nil, err
	}
	return r.GetBookingDetail(ctx, id)
}

func scanConsultationService(row pgx.Row) (*model.ConsultationService, error) {
	var item model.ConsultationService
	err := row.Scan(
		&item.ID, &item.Slug, &item.Name, &item.Description, &item.DurationMinutes, &item.PriceRUB,
		&item.IsActive, &item.MinNoticeMinutes, &item.MaxAdvanceDays, &item.BufferBeforeMinutes,
		&item.BufferAfterMinutes, &item.MeetingURL, &item.SortOrder, &item.CreatedAt, &item.UpdatedAt,
	)
	return &item, err
}

func scanConsultationBooking(row pgx.Row) (*model.ConsultationBooking, error) {
	var item model.ConsultationBooking
	err := row.Scan(
		&item.ID, &item.ServiceID, &item.UserID, &item.CustomerName, &item.CustomerEmail,
		&item.CustomerPhone, &item.CustomerTelegram, &item.CustomerNote, &item.StartsAt,
		&item.EndsAt, &item.Timezone, &item.Status, &item.AmountRUB, &item.Provider,
		&item.ExternalID, &item.InvID, &item.MeetingURL, &item.ExpiresAt, &item.PaidAt,
		&item.CreatedAt, &item.UpdatedAt,
	)
	return &item, err
}

func scanConsultationBookingDetail(row pgx.Row) (*model.ConsultationBookingDetail, error) {
	var item model.ConsultationBookingDetail
	err := row.Scan(
		&item.ID, &item.ServiceID, &item.UserID, &item.CustomerName, &item.CustomerEmail,
		&item.CustomerPhone, &item.CustomerTelegram, &item.CustomerNote, &item.StartsAt,
		&item.EndsAt, &item.Timezone, &item.Status, &item.AmountRUB, &item.Provider,
		&item.ExternalID, &item.InvID, &item.MeetingURL, &item.ExpiresAt, &item.PaidAt,
		&item.CreatedAt, &item.UpdatedAt, &item.ServiceName, &item.ServiceSlug,
		&item.DurationMinutes, &item.ServiceDescription,
	)
	return &item, err
}
