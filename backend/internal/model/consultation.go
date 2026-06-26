package model

import "time"

const (
	ConsultationBookingStatusPendingPayment = "pending_payment"
	ConsultationBookingStatusPaid           = "paid"
	ConsultationBookingStatusCancelled      = "cancelled"
	ConsultationBookingStatusExpired        = "expired"
	ConsultationBookingStatusCompleted      = "completed"
	ConsultationBookingStatusNoShow         = "no_show"
)

type ConsultationService struct {
	ID                  string    `json:"id"`
	Slug                string    `json:"slug"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	DurationMinutes     int       `json:"duration_minutes"`
	PriceRUB            int       `json:"price_rub"`
	IsActive            bool      `json:"is_active"`
	MinNoticeMinutes    int       `json:"min_notice_minutes"`
	MaxAdvanceDays      int       `json:"max_advance_days"`
	BufferBeforeMinutes int       `json:"buffer_before_minutes"`
	BufferAfterMinutes  int       `json:"buffer_after_minutes"`
	MeetingURL          string    `json:"meeting_url"`
	SortOrder           int       `json:"sort_order"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type ConsultationAvailabilityRule struct {
	ID              string    `json:"id"`
	ServiceID       string    `json:"service_id"`
	Weekday         int       `json:"weekday"`
	StartTime       string    `json:"start_time"`
	EndTime         string    `json:"end_time"`
	SlotStepMinutes int       `json:"slot_step_minutes"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

type ConsultationBlackout struct {
	ID        string    `json:"id"`
	ServiceID *string   `json:"service_id,omitempty"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type ConsultationBooking struct {
	ID               string     `json:"id"`
	ServiceID        string     `json:"service_id"`
	UserID           *string    `json:"user_id,omitempty"`
	CustomerName     string     `json:"customer_name"`
	CustomerEmail    string     `json:"customer_email"`
	CustomerPhone    string     `json:"customer_phone"`
	CustomerTelegram string     `json:"customer_telegram"`
	CustomerNote     string     `json:"customer_note"`
	StartsAt         time.Time  `json:"starts_at"`
	EndsAt           time.Time  `json:"ends_at"`
	Timezone         string     `json:"timezone"`
	Status           string     `json:"status"`
	AmountRUB        int        `json:"amount_rub"`
	Provider         *string    `json:"provider,omitempty"`
	ExternalID       *string    `json:"external_id,omitempty"`
	InvID            *int64     `json:"inv_id,omitempty"`
	MeetingURL       string     `json:"meeting_url"`
	ExpiresAt        time.Time  `json:"expires_at"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ConsultationBookingDetail struct {
	ConsultationBooking
	ServiceName        string `json:"service_name"`
	ServiceSlug        string `json:"service_slug"`
	DurationMinutes    int    `json:"duration_minutes"`
	ServiceDescription string `json:"service_description"`
}

type ConsultationSlot struct {
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	Available bool      `json:"available"`
}
