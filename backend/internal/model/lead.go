package model

import "time"

const (
	AccountSegmentPartner    = "partner"
	AccountSegmentDirectLead = "direct_lead"
)

type SharedReport struct {
	ID        string     `json:"id"`
	RunID     string     `json:"run_id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	ViewCount int        `json:"view_count"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type Lead struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	RunID     string    `json:"run_id"`
	Name      string    `json:"name"`
	Company   *string   `json:"company,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Email     string    `json:"email"`
	Message   *string   `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
