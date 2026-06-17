package model

import (
	"encoding/json"
	"time"
)

type LegalRisk struct {
	RiskID          string          `json:"risk_id"`
	TitleRU         string          `json:"title_ru"`
	TitleEN         string          `json:"title_en"`
	WhatRU          string          `json:"what_ru"`
	Article         string          `json:"article"`
	FineTextRU      string          `json:"fine_text_ru"`
	FineMin         *int            `json:"fine_min"`
	FineMax         *int            `json:"fine_max"`
	Severity        string          `json:"severity"`
	HowToFixRU      string          `json:"how_to_fix_ru"`
	HowToFixEN      string          `json:"how_to_fix_en"`
	TriggerFindings json.RawMessage `json:"trigger_findings"`
	TriggerFlags    json.RawMessage `json:"trigger_flags"`
	IsTurnoverFine  bool            `json:"is_turnover_fine"`
	IsContextOnly   bool            `json:"is_context_only"`
	IsActive        bool            `json:"is_active"`
	SortOrder       int             `json:"sort_order"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type LegalRiskUpdateRequest struct {
	TitleRU         string          `json:"title_ru"`
	TitleEN         string          `json:"title_en"`
	WhatRU          string          `json:"what_ru"`
	Article         string          `json:"article"`
	FineTextRU      string          `json:"fine_text_ru"`
	FineMin         *int            `json:"fine_min"`
	FineMax         *int            `json:"fine_max"`
	Severity        string          `json:"severity"`
	HowToFixRU      string          `json:"how_to_fix_ru"`
	HowToFixEN      string          `json:"how_to_fix_en"`
	TriggerFindings json.RawMessage `json:"trigger_findings"`
	TriggerFlags    json.RawMessage `json:"trigger_flags"`
	IsTurnoverFine  bool            `json:"is_turnover_fine"`
	IsContextOnly   bool            `json:"is_context_only"`
	IsActive        bool            `json:"is_active"`
	SortOrder       int             `json:"sort_order"`
}
