package model

import "time"

type UITooltip struct {
	Key          string            `json:"key"`
	Label        string            `json:"label"`
	TextRU       string            `json:"text_ru"`
	TextEN       string            `json:"text_en"`
	LocaleTexts  map[string]string `json:"locale_texts,omitempty"`
	SortOrder    int               `json:"sort_order"`
	UpdatedAt    time.Time         `json:"updated_at"`
}
