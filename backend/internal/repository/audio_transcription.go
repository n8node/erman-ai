package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AudioTranscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewAudioTranscriptionRepository(pool *pgxpool.Pool) *AudioTranscriptionRepository {
	return &AudioTranscriptionRepository{pool: pool}
}

func (r *AudioTranscriptionRepository) HasExplicitAccess(ctx context.Context, userID string) (bool, error) {
	var enabled bool
	err := r.pool.QueryRow(ctx, `SELECT enabled FROM audio_transcription_access WHERE user_id=$1`, userID).Scan(&enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return enabled, err
}

func (r *AudioTranscriptionRepository) SetAccess(ctx context.Context, userID string, enabled bool) error {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO audio_transcription_access(user_id, enabled) VALUES($1,$2)
		ON CONFLICT(user_id) DO UPDATE SET enabled=EXCLUDED.enabled, updated_at=NOW()`, userID, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AudioTranscriptionRepository) ListAccessUsers(ctx context.Context) ([]model.AudioTranscriptionAccessUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id,u.email,u.role,(u.role='superadmin' OR COALESCE(a.enabled,false))
		FROM users u LEFT JOIN audio_transcription_access a ON a.user_id=u.id
		ORDER BY u.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AudioTranscriptionAccessUser
	for rows.Next() {
		var item model.AudioTranscriptionAccessUser
		if err := rows.Scan(&item.ID, &item.Email, &item.Role, &item.HasAccess); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *AudioTranscriptionRepository) GetSettings(ctx context.Context) (*model.AudioTranscriptionSettingsRecord, error) {
	var raw []byte
	var rec model.AudioTranscriptionSettingsRecord
	if err := r.pool.QueryRow(ctx, `SELECT config,updated_at FROM audio_transcription_settings WHERE id=1`).Scan(&raw, &rec.UpdatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &rec.Settings); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *AudioTranscriptionRepository) UpdateSettings(ctx context.Context, settings model.AudioTranscriptionSettings) (*model.AudioTranscriptionSettingsRecord, error) {
	raw, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	var out []byte
	var rec model.AudioTranscriptionSettingsRecord
	err = r.pool.QueryRow(ctx, `UPDATE audio_transcription_settings SET config=$1,updated_at=NOW() WHERE id=1 RETURNING config,updated_at`, raw).Scan(&out, &rec.UpdatedAt)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &rec.Settings)
	return &rec, err
}

func (r *AudioTranscriptionRepository) CreateFile(ctx context.Context, userID, originalName, assetPath, contentType string, size int64, duration *float64) (*model.AudioTranscriptionFile, error) {
	return scanAudioFile(r.pool.QueryRow(ctx, `
		INSERT INTO audio_transcription_files(user_id,original_name,asset_path,content_type,size_bytes,duration_sec)
		VALUES($1,$2,$3,$4,$5,$6)
		RETURNING id,user_id,original_name,content_type,size_bytes,duration_sec,
		          (transcript_path IS NOT NULL AND transcript_path <> ''),created_at,updated_at`,
		userID, originalName, assetPath, contentType, size, duration))
}

func (r *AudioTranscriptionRepository) ListFiles(ctx context.Context, userID string) ([]model.AudioTranscriptionFile, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT f.id,f.user_id,f.original_name,f.content_type,f.size_bytes,f.duration_sec,
		       (f.transcript_path IS NOT NULL AND f.transcript_path <> ''),f.created_at,f.updated_at,
		       lr.run_id,lr.status
		FROM audio_transcription_files f
		LEFT JOIN LATERAL (
			SELECT tr.id AS run_id, tr.status
			FROM audio_transcription_file_runs fr
			JOIN tool_runs tr ON tr.id = fr.run_id
			WHERE fr.file_id = f.id
			ORDER BY tr.created_at DESC
			LIMIT 1
		) lr ON true
		WHERE f.user_id=$1
		ORDER BY f.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AudioTranscriptionFile
	for rows.Next() {
		item, err := scanAudioFileWithRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *AudioTranscriptionRepository) GetFile(ctx context.Context, fileID, userID string) (*model.AudioTranscriptionFile, error) {
	file, err := scanAudioFile(r.pool.QueryRow(ctx, `
		SELECT id,user_id,original_name,content_type,size_bytes,duration_sec,
		       (transcript_path IS NOT NULL AND transcript_path <> ''),created_at,updated_at
		FROM audio_transcription_files WHERE id=$1 AND user_id=$2`, fileID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return file, err
}

func (r *AudioTranscriptionRepository) FileAssetPath(ctx context.Context, fileID, userID string) (string, error) {
	var path string
	err := r.pool.QueryRow(ctx, `SELECT asset_path FROM audio_transcription_files WHERE id=$1 AND user_id=$2`, fileID, userID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return path, err
}

func (r *AudioTranscriptionRepository) TranscriptPath(ctx context.Context, fileID, userID string) (string, string, error) {
	var path, name string
	err := r.pool.QueryRow(ctx, `
		SELECT transcript_path, original_name FROM audio_transcription_files
		WHERE id=$1 AND user_id=$2 AND transcript_path IS NOT NULL AND transcript_path <> ''`,
		fileID, userID).Scan(&path, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return path, name, err
}

func (r *AudioTranscriptionRepository) SetTranscriptPath(ctx context.Context, fileID, userID, transcriptPath string, duration *float64) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE audio_transcription_files
		SET transcript_path=$3, duration_sec=COALESCE($4, duration_sec), updated_at=NOW()
		WHERE id=$1 AND user_id=$2`, fileID, userID, transcriptPath, duration)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AudioTranscriptionRepository) DeleteFile(ctx context.Context, fileID, userID string) (string, *string, error) {
	var assetPath string
	var transcriptPath *string
	err := r.pool.QueryRow(ctx, `
		DELETE FROM audio_transcription_files WHERE id=$1 AND user_id=$2
		RETURNING asset_path, transcript_path`, fileID, userID).Scan(&assetPath, &transcriptPath)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, ErrNotFound
	}
	return assetPath, transcriptPath, err
}

func (r *AudioTranscriptionRepository) AttachRun(ctx context.Context, fileID, runID string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO audio_transcription_file_runs(file_id,run_id) VALUES($1,$2)`, fileID, runID)
	return err
}

func (r *AudioTranscriptionRepository) ListFileRuns(ctx context.Context, fileID, userID string) ([]model.ToolRun, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT tr.id,tr.user_id,tr.tool_slug,tr.plan_tier,tr.input,tr.output,tr.artifact_url,tr.tokens_used,
		       tr.model_used,tr.status,tr.error_msg,tr.created_at,tr.updated_at,tr.completed_at
		FROM tool_runs tr JOIN audio_transcription_file_runs fr ON fr.run_id=tr.id
		JOIN audio_transcription_files f ON f.id=fr.file_id
		WHERE fr.file_id=$1 AND f.user_id=$2 ORDER BY tr.created_at DESC`, fileID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ToolRun
	for rows.Next() {
		var item model.ToolRun
		var status string
		if err := rows.Scan(&item.ID, &item.UserID, &item.ToolSlug, &item.PlanTier, &item.Input, &item.Output,
			&item.ArtifactURL, &item.TokensUsed, &item.ModelUsed, &status, &item.ErrorMsg, &item.CreatedAt,
			&item.UpdatedAt, &item.CompletedAt); err != nil {
			return nil, err
		}
		item.Status = model.RunStatus(status)
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanAudioFile(row pgx.Row) (*model.AudioTranscriptionFile, error) {
	var item model.AudioTranscriptionFile
	err := row.Scan(&item.ID, &item.UserID, &item.OriginalName, &item.ContentType, &item.SizeBytes,
		&item.DurationSec, &item.HasTranscript, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func scanAudioFileWithRun(row pgx.Row) (*model.AudioTranscriptionFile, error) {
	var item model.AudioTranscriptionFile
	var runID *string
	var runStatus *string
	err := row.Scan(&item.ID, &item.UserID, &item.OriginalName, &item.ContentType, &item.SizeBytes,
		&item.DurationSec, &item.HasTranscript, &item.CreatedAt, &item.UpdatedAt, &runID, &runStatus)
	if err != nil {
		return nil, err
	}
	item.LatestRunID = runID
	if runStatus != nil {
		st := model.RunStatus(*runStatus)
		item.LatestRunStatus = &st
	}
	return &item, err
}
