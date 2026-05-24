package model

import "time"

type ProposalRequestStatus string

const (
	ProposalRequestStatusNew        ProposalRequestStatus = "new"
	ProposalRequestStatusInProgress ProposalRequestStatus = "in_progress"
	ProposalRequestStatusDone       ProposalRequestStatus = "done"
)

type ProposalRequest struct {
	ID             string                `json:"id"`
	UserID         string                `json:"user_id"`
	RunID          string                `json:"run_id"`
	RequesterName  string                `json:"requester_name"`
	Telegram       string                `json:"telegram"`
	BusinessNote   string                `json:"business_note"`
	Status         ProposalRequestStatus `json:"status"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}
