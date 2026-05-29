package model

import "time"

type ProjectInquiryStatus string

const (
	ProjectInquiryStatusPendingEmail ProjectInquiryStatus = "pending_email"
	ProjectInquiryStatusNew          ProjectInquiryStatus = "new"
	ProjectInquiryStatusInProgress   ProjectInquiryStatus = "in_progress"
	ProjectInquiryStatusDone         ProjectInquiryStatus = "done"
	ProjectInquiryStatusSpam         ProjectInquiryStatus = "spam"
)

type ProjectInquiry struct {
	ID                string               `json:"id"`
	UserID            *string              `json:"user_id,omitempty"`
	CalculatorRunID   *string              `json:"calculator_run_id,omitempty"`
	Name              string               `json:"name"`
	Email             string               `json:"email"`
	Telegram          string               `json:"telegram"`
	ProjectTitle      string               `json:"project_title"`
	ProjectDescription string              `json:"project_description"`
	Status            ProjectInquiryStatus `json:"status"`
	EmailVerifiedAt   *time.Time           `json:"email_verified_at,omitempty"`
	Locale            string               `json:"locale"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
}
