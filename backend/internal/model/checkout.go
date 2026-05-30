package model

import "time"

const (
	CheckoutStatusPending = "pending"
	CheckoutStatusPaid    = "paid"
	CheckoutStatusFailed  = "failed"
)

type PlanCheckout struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	PlanID     string     `json:"plan_id"`
	Provider   string     `json:"provider"`
	AmountRUB  int        `json:"amount_rub"`
	Status     string     `json:"status"`
	ExternalID *string    `json:"external_id,omitempty"`
	InvID      *int64     `json:"inv_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
}

type CheckoutResult struct {
	CheckoutID  string `json:"checkout_id"`
	Provider    string `json:"provider"`
	CheckoutURL string `json:"checkout_url"`
}
