package model

import "time"

type PublicPage struct {
	ID              string    `json:"id"`
	Slug            string    `json:"slug"`
	Title           string    `json:"title"`
	ContentHTML     string    `json:"content_html"`
	MetaDescription string    `json:"meta_description"`
	IsPublished     bool      `json:"is_published"`
	SortOrder       int       `json:"sort_order"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatedAt       time.Time `json:"created_at"`
}
