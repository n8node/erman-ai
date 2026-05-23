package model

import (
	"encoding/json"
	"time"
)

type RunStatus string

const (
	RunStatusPending    RunStatus = "pending"
	RunStatusProcessing RunStatus = "processing"
	RunStatusDone       RunStatus = "done"
	RunStatusError      RunStatus = "error"
)

type ToolRun struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	ToolSlug    string          `json:"tool_slug"`
	PlanTier    string          `json:"plan_tier"`
	Input       json.RawMessage `json:"input"`
	Output      json.RawMessage `json:"output,omitempty"`
	ArtifactURL *string         `json:"artifact_url,omitempty"`
	TokensUsed  int64           `json:"tokens_used"`
	ModelUsed   *string         `json:"model_used,omitempty"`
	Status      RunStatus       `json:"status"`
	ErrorMsg    *string         `json:"error_msg,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}
