package model

import "time"

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	PlanID       *string    `json:"plan_id"`
	Locale       string     `json:"locale"`
	IsBlocked    bool       `json:"is_blocked"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
}
