package model

import "time"

type TokenPackage struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Tokens    int64     `json:"tokens"`
	PriceRUB  int       `json:"price_rub"`
	SortOrder int       `json:"sort_order"`
	IsPublic  bool      `json:"is_public"`
	IsArchived bool     `json:"is_archived"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TokenPackageCheckout struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	TokenPackageID  string     `json:"token_package_id"`
	Provider        string     `json:"provider"`
	AmountRUB       int        `json:"amount_rub"`
	TokensAmount    int64      `json:"tokens_amount"`
	Status          string     `json:"status"`
	ExternalID      *string    `json:"external_id,omitempty"`
	InvID           *int64     `json:"inv_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
}

type TokenPackageUpsertInput struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Tokens     int64  `json:"tokens"`
	PriceRUB   int    `json:"price_rub"`
	SortOrder  int    `json:"sort_order"`
	IsPublic   bool   `json:"is_public"`
	IsArchived bool   `json:"is_archived"`
}
