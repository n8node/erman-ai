package model

import (
	"encoding/json"
	"time"
)

type GeologicalJournalDocumentStatus string

const (
	GeologicalJournalDocumentUploaded      GeologicalJournalDocumentStatus = "uploaded"
	GeologicalJournalDocumentPreviewReady  GeologicalJournalDocumentStatus = "preview_ready"
	GeologicalJournalDocumentQueued        GeologicalJournalDocumentStatus = "queued"
	GeologicalJournalDocumentProcessing    GeologicalJournalDocumentStatus = "processing"
	GeologicalJournalDocumentPartiallyDone GeologicalJournalDocumentStatus = "partially_done"
	GeologicalJournalDocumentDone          GeologicalJournalDocumentStatus = "done"
	GeologicalJournalDocumentError         GeologicalJournalDocumentStatus = "error"
)

type GeologicalJournalDocumentPageStatus string

const (
	GeologicalJournalDocumentPageQueued        GeologicalJournalDocumentPageStatus = "queued"
	GeologicalJournalDocumentPageRendering     GeologicalJournalDocumentPageStatus = "rendering"
	GeologicalJournalDocumentPageOrienting     GeologicalJournalDocumentPageStatus = "orienting"
	GeologicalJournalDocumentPagePreprocessing GeologicalJournalDocumentPageStatus = "preprocessing"
	GeologicalJournalDocumentPageOCR           GeologicalJournalDocumentPageStatus = "ocr"
	GeologicalJournalDocumentPageClassifying   GeologicalJournalDocumentPageStatus = "classifying"
	GeologicalJournalDocumentPageTables        GeologicalJournalDocumentPageStatus = "detecting_tables"
	GeologicalJournalDocumentPageDone          GeologicalJournalDocumentPageStatus = "done"
	GeologicalJournalDocumentPageReview        GeologicalJournalDocumentPageStatus = "needs_review"
	GeologicalJournalDocumentPageError         GeologicalJournalDocumentPageStatus = "error"
)

type GeologicalJournalDocumentJob struct {
	ID           string `json:"id"`
	DocumentID   string `json:"document_id"`
	PageID       string `json:"page_id"`
	DocumentPath string `json:"-"`
	PageNumber   int    `json:"page_number"`
	Status       string `json:"status"`
	Phase        string `json:"phase"`
	Attempts     int    `json:"attempts"`
}

type GeologicalJournalDocument struct {
	ID                  string                          `json:"id"`
	OwnerID             string                          `json:"owner_id,omitempty"`
	OriginalName        string                          `json:"original_name"`
	ContentType         string                          `json:"content_type"`
	SizeBytes           int64                           `json:"size_bytes"`
	PageCount           int                             `json:"page_count"`
	Status              GeologicalJournalDocumentStatus `json:"status"`
	IsShared            bool                            `json:"is_shared"`
	ErrorMsg            *string                         `json:"error_msg,omitempty"`
	AssetPath           string                          `json:"-"`
	CreatedAt           time.Time                       `json:"created_at"`
	UpdatedAt           time.Time                       `json:"updated_at"`
	AnalysisStartedAt   *time.Time                      `json:"analysis_started_at,omitempty"`
	AnalysisCompletedAt *time.Time                      `json:"analysis_completed_at,omitempty"`
}

type GeologicalJournalDocumentPage struct {
	ID                    string                              `json:"id"`
	DocumentID            string                              `json:"document_id"`
	PageNumber            int                                 `json:"page_number"`
	Status                GeologicalJournalDocumentPageStatus `json:"status"`
	OCRText               string                              `json:"ocr_text,omitempty"`
	TableResult           *GeologicalJournalOutput            `json:"table_result,omitempty"`
	Analysis              json.RawMessage                     `json:"analysis,omitempty"`
	ContentType           string                              `json:"content_type,omitempty"`
	OrientationDegrees    int                                 `json:"orientation_degrees"`
	OrientationConfidence float64                             `json:"orientation_confidence"`
	TableCount            int                                 `json:"table_count"`
	TextCharCount         int                                 `json:"text_char_count"`
	ErrorMsg              *string                             `json:"error_msg,omitempty"`
	OriginalAssetPath     string                              `json:"-"`
	OrientedAssetPath     string                              `json:"-"`
	PreprocessedAssetPath string                              `json:"-"`
	CreatedAt             time.Time                           `json:"created_at"`
	UpdatedAt             time.Time                           `json:"updated_at"`
}

type GeologicalJournalDocumentDetail struct {
	Document GeologicalJournalDocument       `json:"document"`
	Pages    []GeologicalJournalDocumentPage `json:"pages"`
}

type GeologicalJournalDocumentChatMessage struct {
	ID         string          `json:"id"`
	Role       string          `json:"role"`
	Content    string          `json:"content"`
	Sources    json.RawMessage `json:"sources"`
	Confidence string          `json:"confidence,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type GeologicalJournalDocumentChatRequest struct {
	SessionID        string `json:"session_id,omitempty"`
	PageNumbers      []int  `json:"page_numbers"`
	IncludeNeighbors bool   `json:"include_neighbors"`
	Message          string `json:"message"`
}

type GeologicalJournalDocumentLLMRequest struct {
	PageNumbers      []int  `json:"page_numbers"`
	IncludeNeighbors bool   `json:"include_neighbors"`
	Mode             string `json:"mode"`
}

type GeologicalJournalDocumentPagesRequest struct {
	PageNumbers []int `json:"page_numbers"`
}

type GeologicalJournalDocumentLLMResult struct {
	ID         string          `json:"id"`
	DocumentID string          `json:"document_id"`
	PageID     string          `json:"page_id"`
	Mode       string          `json:"mode"`
	Result     json.RawMessage `json:"result"`
	ModelUsed  string          `json:"model_used,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type GeologicalJournalDocumentAnalysisSummary struct {
	DocumentID     string `json:"document_id"`
	TotalPages     int    `json:"total_pages"`
	ProcessedPages int    `json:"processed_pages"`
	QueuedPages    int    `json:"queued_pages"`
	ErrorPages     int    `json:"error_pages"`
	ReviewPages    int    `json:"review_pages"`
	TablePages     int    `json:"table_pages"`
	TextPages      int    `json:"text_pages"`
	MixedPages     int    `json:"mixed_pages"`
}
