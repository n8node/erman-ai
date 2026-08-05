package model

import (
	"strings"
	"time"
)

const VideoTranscriptionToolSlug = "video-transcription"

type VideoTranscriptionSettings struct {
	Model                    string  `json:"model"`
	LanguageCode             string  `json:"language_code"`
	PriceRUBPerMinute        float64 `json:"price_rub_per_minute"`
	TextNormalizationEnabled bool    `json:"text_normalization_enabled"`
	LiteratureText           bool    `json:"literature_text"`
	ProfanityFilter          bool    `json:"profanity_filter"`
}

func ApplyVideoTranscriptionDefaults(settings VideoTranscriptionSettings) VideoTranscriptionSettings {
	if strings.TrimSpace(settings.Model) == "" {
		settings.Model = "general"
	}
	if strings.TrimSpace(settings.LanguageCode) == "" {
		settings.LanguageCode = "ru-RU"
	}
	return settings
}

type VideoTranscriptionSettingsRecord struct {
	Settings  VideoTranscriptionSettings `json:"settings"`
	UpdatedAt time.Time                  `json:"updated_at"`
}

type VideoTranscriptionAccessUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	HasAccess bool   `json:"has_access"`
}

type VideoTranscriptionFile struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id,omitempty"`
	OriginalName    string     `json:"original_name"`
	ContentType     string     `json:"content_type"`
	SizeBytes       int64      `json:"size_bytes"`
	DurationSec     *float64   `json:"duration_sec,omitempty"`
	HasTranscript   bool       `json:"has_transcript"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LatestRunID     *string    `json:"latest_run_id,omitempty"`
	LatestRunStatus *RunStatus `json:"latest_run_status,omitempty"`
}

type VideoTranscriptionFileDetail struct {
	VideoTranscriptionFile
	Runs []ToolRun `json:"runs"`
}

type VideoTranscriptionRunInput struct {
	FileID string `json:"file_id"`
}

type VideoTranscriptionOutput struct {
	Text        string  `json:"text"`
	Language    string  `json:"language"`
	Model       string  `json:"model"`
	DurationSec float64 `json:"duration_sec"`
	ChunkCount  int     `json:"chunk_count"`
	Preview     string  `json:"preview,omitempty"`
}
