package model

import (
	"encoding/json"
	"time"
)

const GeologicalJournalToolSlug = "geological-journal"

type GeologicalJournalRow struct {
	Date               string              `json:"date"`
	DrillingDiameterMM *float64            `json:"drilling_diameter_mm"`
	DepthFromM         *float64            `json:"depth_from_m"`
	DepthToM           *float64            `json:"depth_to_m"`
	DrillingRunM       *float64            `json:"drilling_run_m"`
	CoreRecoveryM      *float64            `json:"core_recovery_m"`
	CoreRecoveryPct    *float64            `json:"core_recovery_pct"`
	RockDescription    string              `json:"rock_description"`
	SamplingInterval   string              `json:"sampling_interval"`
	SampleNumber       string              `json:"sample_number"`
	Notes              string              `json:"notes"`
	Uncertainties      []string            `json:"uncertainties"`
	FieldIssues        map[string][]string `json:"field_issues,omitempty"`
}

type GeologicalJournalOutput struct {
	Rows []GeologicalJournalRow `json:"rows"`
}

type GeologicalJournalPage struct {
	ID                   string          `json:"id"`
	UserID               string          `json:"user_id,omitempty"`
	OriginalName         string          `json:"original_name"`
	ContentType          string          `json:"content_type"`
	SizeBytes            int64           `json:"size_bytes"`
	Width                int             `json:"width"`
	Height               int             `json:"height"`
	HasPreprocessedImage bool            `json:"has_preprocessed_image"`
	LatestResult         json.RawMessage `json:"latest_result,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type GeologicalJournalPageDetail struct {
	GeologicalJournalPage
	Runs     []ToolRun                        `json:"runs"`
	Versions []GeologicalJournalResultVersion `json:"versions"`
}

type GeologicalJournalResultVersion struct {
	ID        string          `json:"id"`
	PageID    string          `json:"page_id"`
	UserID    string          `json:"user_id"`
	Result    json.RawMessage `json:"result"`
	CreatedAt time.Time       `json:"created_at"`
}

type GeologicalJournalSettings struct {
	Provider        LLMProvider `json:"provider"`
	OpenRouterModel string      `json:"openrouter_model"`
	DeepSeekModel   string      `json:"deepseek_model"`
	YandexModel     string      `json:"yandex_model"`
	OCRModel        string      `json:"ocr_model"`
	SystemPrompt    string      `json:"system_prompt"`
	Temperature     float64     `json:"temperature"`
	MaxTokens       int         `json:"max_tokens"`
}

func (s GeologicalJournalSettings) ActiveModel() string {
	switch s.Provider {
	case LLMProviderDeepSeek:
		return s.DeepSeekModel
	case LLMProviderYandex:
		return s.YandexModel
	default:
		return s.OpenRouterModel
	}
}

type GeologicalJournalSettingsRecord struct {
	Settings  GeologicalJournalSettings `json:"settings"`
	Providers []LLMProviderStatus       `json:"providers,omitempty"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

type GeologicalJournalAccessUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	HasAccess bool   `json:"has_access"`
}

type GeologicalJournalExample struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	SortOrder   int       `json:"sort_order"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GeologicalJournalExampleMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
}

type GeologicalJournalRunPhase string

const (
	GeologicalJournalPhasePreprocessing GeologicalJournalRunPhase = "preprocessing"
	GeologicalJournalPhaseOCR           GeologicalJournalRunPhase = "ocr"
	GeologicalJournalPhaseStructuring   GeologicalJournalRunPhase = "structuring"
)

type GeologicalJournalPreprocessingInfo struct {
	Applied              bool    `json:"applied"`
	UsedForOCR           bool    `json:"used_for_ocr"`
	HasPreprocessedImage bool    `json:"has_preprocessed_image"`
	PerspectiveCorrected bool    `json:"perspective_corrected"`
	DeskewAngle          float64 `json:"deskew_angle"`
	Scale                float64 `json:"scale"`
	Width                int     `json:"width"`
	Height               int     `json:"height"`
	FallbackReason       string  `json:"fallback_reason,omitempty"`
}

type GeologicalJournalRunInput struct {
	PageID        string                              `json:"page_id"`
	Phase         GeologicalJournalRunPhase           `json:"phase,omitempty"`
	Preprocessing *GeologicalJournalPreprocessingInfo `json:"preprocessing,omitempty"`
}
