package model

import "time"

type UITranslation struct {
	Key       string    `json:"key"`
	Locale    string    `json:"locale"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TranslationListItem struct {
	Key    string            `json:"key"`
	Values map[string]string `json:"values"`
}
