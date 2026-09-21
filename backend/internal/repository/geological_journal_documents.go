package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
)

func (r *GeologicalJournalRepository) CreateDocument(ctx context.Context, ownerID, name, path, contentType string, size int64) (*model.GeologicalJournalDocument, error) {
	var item model.GeologicalJournalDocument
	err := r.pool.QueryRow(ctx, `
		INSERT INTO geological_journal_documents(owner_id,original_name,asset_path,content_type,size_bytes)
		VALUES($1,$2,$3,$4,$5)
		RETURNING id,owner_id,original_name,asset_path,content_type,size_bytes,page_count,status,is_shared,error_msg,created_at,updated_at,analysis_started_at,analysis_completed_at`,
		ownerID, name, path, contentType, size).Scan(
		&item.ID, &item.OwnerID, &item.OriginalName, &item.AssetPath, &item.ContentType, &item.SizeBytes,
		&item.PageCount, &item.Status, &item.IsShared, &item.ErrorMsg, &item.CreatedAt,
		&item.UpdatedAt, &item.AnalysisStartedAt, &item.AnalysisCompletedAt)
	return &item, err
}

func (r *GeologicalJournalRepository) GetDocument(ctx context.Context, id string) (*model.GeologicalJournalDocument, error) {
	var item model.GeologicalJournalDocument
	err := r.pool.QueryRow(ctx, `
		SELECT id,owner_id,original_name,asset_path,content_type,size_bytes,page_count,status,is_shared,error_msg,created_at,updated_at,analysis_started_at,analysis_completed_at
		FROM geological_journal_documents WHERE id=$1`, id).Scan(
		&item.ID, &item.OwnerID, &item.OriginalName, &item.AssetPath, &item.ContentType, &item.SizeBytes,
		&item.PageCount, &item.Status, &item.IsShared, &item.ErrorMsg, &item.CreatedAt,
		&item.UpdatedAt, &item.AnalysisStartedAt, &item.AnalysisCompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &item, err
}

func (r *GeologicalJournalRepository) DeleteDocument(ctx context.Context, id, ownerID string, superadmin bool) (string, error) {
	var assetPath string
	var err error
	if superadmin {
		err = r.pool.QueryRow(ctx, `DELETE FROM geological_journal_documents WHERE id=$1 RETURNING asset_path`, id).Scan(&assetPath)
	} else {
		err = r.pool.QueryRow(ctx, `DELETE FROM geological_journal_documents WHERE id=$1 AND owner_id=$2 RETURNING asset_path`, id, ownerID).Scan(&assetPath)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return assetPath, err
}

func (r *GeologicalJournalRepository) ListDocuments(ctx context.Context, userID string, includeShared bool) ([]model.GeologicalJournalDocument, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,owner_id,original_name,asset_path,content_type,size_bytes,page_count,status,is_shared,error_msg,created_at,updated_at,analysis_started_at,analysis_completed_at
		FROM geological_journal_documents WHERE owner_id=$1 OR ($2 AND is_shared)
		ORDER BY created_at DESC`, userID, includeShared)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.GeologicalJournalDocument
	for rows.Next() {
		var item model.GeologicalJournalDocument
		if err := rows.Scan(&item.ID, &item.OwnerID, &item.OriginalName, &item.AssetPath, &item.ContentType, &item.SizeBytes,
			&item.PageCount, &item.Status, &item.IsShared, &item.ErrorMsg, &item.CreatedAt, &item.UpdatedAt,
			&item.AnalysisStartedAt, &item.AnalysisCompletedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *GeologicalJournalRepository) SetDocumentShared(ctx context.Context, id string, shared bool) error {
	tag, err := r.pool.Exec(ctx, `UPDATE geological_journal_documents SET is_shared=$2,updated_at=NOW() WHERE id=$1`, id, shared)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GeologicalJournalRepository) SetDocumentPageCount(ctx context.Context, id string, count int) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_documents SET page_count=$2,status='queued',updated_at=NOW() WHERE id=$1`, id, count)
	return err
}

func (r *GeologicalJournalRepository) MarkDocumentQueued(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_documents SET status='queued',analysis_started_at=NOW(),analysis_completed_at=NULL,updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *GeologicalJournalRepository) MarkDocumentProcessing(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_documents SET status='processing',analysis_started_at=COALESCE(analysis_started_at,NOW()),updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (r *GeologicalJournalRepository) ResetDocumentJobs(ctx context.Context, documentID string) error {
	if _, err := r.pool.Exec(ctx, `
		UPDATE geological_journal_document_pages
		SET status='queued',error_msg=NULL,updated_at=NOW()
		WHERE document_id=$1 AND status <> 'done'`, documentID); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE geological_journal_document_jobs j
		SET status='queued',phase='queued',lease_until=NULL,error_msg=NULL,completed_at=NULL,updated_at=NOW()
		FROM geological_journal_document_pages p
		WHERE j.page_id=p.id AND j.document_id=$1 AND p.status <> 'done'`, documentID)
	return err
}

func (r *GeologicalJournalRepository) CancelDocumentJobs(ctx context.Context, documentID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_document_jobs SET status='cancelled',phase='cancelled',lease_until=NULL,completed_at=COALESCE(completed_at,NOW()),updated_at=NOW(),error_msg='deep analysis cancelled; choose pages explicitly' WHERE document_id=$1 AND status IN ('queued','processing')`, documentID)
	return err
}

func (r *GeologicalJournalRepository) CancelAllDocumentJobs(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_document_jobs SET status='cancelled',phase='cancelled',lease_until=NULL,completed_at=COALESCE(completed_at,NOW()),updated_at=NOW(),error_msg='deep analysis cancelled; choose pages explicitly' WHERE status IN ('queued','processing')`)
	return err
}

func (r *GeologicalJournalRepository) MarkDocumentPreviewReady(ctx context.Context, documentID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_documents SET status='preview_ready',analysis_completed_at=NOW(),updated_at=NOW() WHERE id=$1`, documentID)
	return err
}

func (r *GeologicalJournalRepository) MarkDocumentError(ctx context.Context, documentID, message string) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_documents SET status='error',error_msg=$2,analysis_completed_at=NOW(),updated_at=NOW() WHERE id=$1`, documentID, message)
	return err
}

func (r *GeologicalJournalRepository) CreateDocumentPage(ctx context.Context, documentID string, number int) (*model.GeologicalJournalDocumentPage, error) {
	var item model.GeologicalJournalDocumentPage
	var originalAssetPath, orientedAssetPath, preprocessedAssetPath *string
	var contentType *string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO geological_journal_document_pages(document_id,page_number)
		VALUES($1,$2) ON CONFLICT(document_id,page_number) DO UPDATE SET updated_at=NOW()
		RETURNING id,document_id,page_number,status,original_asset_path,oriented_asset_path,preprocessed_asset_path,ocr_text,table_result,analysis,content_type,orientation_degrees,orientation_confidence,table_count,text_char_count,error_msg,created_at,updated_at`, documentID, number).Scan(
		&item.ID, &item.DocumentID, &item.PageNumber, &item.Status,
		&originalAssetPath, &orientedAssetPath, &preprocessedAssetPath,
		&item.OCRText, &item.TableResult, &item.Analysis,
		&contentType, &item.OrientationDegrees, &item.OrientationConfidence, &item.TableCount,
		&item.TextCharCount, &item.ErrorMsg, &item.CreatedAt, &item.UpdatedAt)
	item.OriginalAssetPath = derefString(originalAssetPath)
	item.OrientedAssetPath = derefString(orientedAssetPath)
	item.PreprocessedAssetPath = derefString(preprocessedAssetPath)
	item.ContentType = derefString(contentType)
	return &item, err
}

func (r *GeologicalJournalRepository) ListDocumentPages(ctx context.Context, documentID string) ([]model.GeologicalJournalDocumentPage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,document_id,page_number,status,original_asset_path,oriented_asset_path,preprocessed_asset_path,ocr_text,table_result,analysis,content_type,orientation_degrees,orientation_confidence,table_count,text_char_count,error_msg,created_at,updated_at
		FROM geological_journal_document_pages WHERE document_id=$1 ORDER BY page_number`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.GeologicalJournalDocumentPage
	for rows.Next() {
		var item model.GeologicalJournalDocumentPage
		var originalAssetPath, orientedAssetPath, preprocessedAssetPath *string
		var contentType *string
		if err := rows.Scan(&item.ID, &item.DocumentID, &item.PageNumber, &item.Status, &originalAssetPath, &orientedAssetPath, &preprocessedAssetPath, &item.OCRText, &item.TableResult, &item.Analysis,
			&contentType, &item.OrientationDegrees, &item.OrientationConfidence, &item.TableCount,
			&item.TextCharCount, &item.ErrorMsg, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.OriginalAssetPath = derefString(originalAssetPath)
		item.OrientedAssetPath = derefString(orientedAssetPath)
		item.PreprocessedAssetPath = derefString(preprocessedAssetPath)
		item.ContentType = derefString(contentType)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *GeologicalJournalRepository) CreateDocumentJob(ctx context.Context, documentID, pageID string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO geological_journal_document_jobs(document_id,page_id) VALUES($1,$2) ON CONFLICT(page_id) DO NOTHING`, documentID, pageID)
	return err
}

func (r *GeologicalJournalRepository) ClaimDocumentJob(ctx context.Context, leaseSeconds int) (*model.GeologicalJournalDocumentJob, error) {
	var job model.GeologicalJournalDocumentJob
	err := r.pool.QueryRow(ctx, `
		WITH candidate AS (
			SELECT j.id FROM geological_journal_document_jobs j
			WHERE (j.status='queued' OR (j.status='processing' AND j.lease_until < NOW()))
			ORDER BY j.created_at
			FOR UPDATE SKIP LOCKED LIMIT 1
		)
		UPDATE geological_journal_document_jobs j
		SET status='processing', phase='rendering', attempts=j.attempts+1,
			lease_until=NOW() + ($1 * INTERVAL '1 second'), started_at=COALESCE(j.started_at,NOW()), updated_at=NOW()
		FROM candidate c
		WHERE j.id=c.id
		RETURNING j.id,j.document_id,j.page_id,j.status,j.phase,j.attempts,
			(SELECT asset_path FROM geological_journal_documents d WHERE d.id=j.document_id),
			(SELECT page_number FROM geological_journal_document_pages p WHERE p.id=j.page_id)`, leaseSeconds).Scan(
		&job.ID, &job.DocumentID, &job.PageID, &job.Status, &job.Phase, &job.Attempts,
		&job.DocumentPath, &job.PageNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &job, err
}

func (r *GeologicalJournalRepository) CompleteDocumentJob(ctx context.Context, jobID, pageID string, status, phase string) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_document_jobs SET status=$2,phase=$3,lease_until=NULL,completed_at=NOW(),updated_at=NOW() WHERE id=$1`, jobID, status, phase)
	return err
}

func (r *GeologicalJournalRepository) FailDocumentJob(ctx context.Context, jobID, pageID, message string) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_document_jobs SET status='error',phase='error',error_msg=$2,lease_until=NULL,completed_at=NOW(),updated_at=NOW() WHERE id=$1`, jobID, message)
	return err
}

func (r *GeologicalJournalRepository) DocumentPageAssetDir(documentID string, pageNumber int) string {
	return filepath.Join("documents", documentID, fmt.Sprintf("page-%04d", pageNumber))
}

func (r *GeologicalJournalRepository) UpdateDocumentPageAnalysis(ctx context.Context, pageID, status, contentType string, orientation int, confidence float64, tables, chars int, text string, raw json.RawMessage, errorMsg *string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE geological_journal_document_pages SET status=$2,content_type=NULLIF($3,''),orientation_degrees=$4,
		orientation_confidence=$5,table_count=$6,text_char_count=$7,ocr_text=$8,analysis=$9,error_msg=$10,updated_at=NOW()
		WHERE id=$1`, pageID, status, contentType, orientation, confidence, tables, chars, text, raw, errorMsg)
	return err
}

func (r *GeologicalJournalRepository) UpdateDocumentPageAssets(ctx context.Context, pageID, original, oriented, preprocessed string) error {
	_, err := r.pool.Exec(ctx, `UPDATE geological_journal_document_pages SET original_asset_path=$2,oriented_asset_path=$3,preprocessed_asset_path=$4,updated_at=NOW() WHERE id=$1`, pageID, original, oriented, preprocessed)
	return err
}

func (r *GeologicalJournalRepository) UpdateDocumentPageOCRText(ctx context.Context, pageID, text string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE geological_journal_document_pages
		SET ocr_text=$2,text_char_count=char_length($2),updated_at=NOW()
		WHERE id=$1`, pageID, text)
	return err
}

func (r *GeologicalJournalRepository) UpdateDocumentPageTableResult(ctx context.Context, pageID string, result *model.GeologicalJournalOutput) error {
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `UPDATE geological_journal_document_pages SET table_result=$2,updated_at=NOW() WHERE id=$1`, pageID, raw)
	return err
}

func (r *GeologicalJournalRepository) RefreshDocumentStatus(ctx context.Context, documentID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE geological_journal_documents d SET
			status = CASE
				WHEN counts.total = counts.done THEN 'done'
				WHEN counts.errors > 0 AND counts.done + counts.errors = counts.total THEN 'error'
				WHEN counts.done > 0 THEN 'partially_done'
				ELSE 'processing'
			END,
			updated_at = NOW(),
			analysis_completed_at = CASE WHEN counts.done + counts.errors = counts.total THEN COALESCE(d.analysis_completed_at, NOW()) ELSE NULL END
		FROM (
			SELECT document_id, COUNT(*)::int AS total,
				COUNT(*) FILTER (WHERE status IN ('done','needs_review'))::int AS done,
				COUNT(*) FILTER (WHERE status='error')::int AS errors
			FROM geological_journal_document_pages WHERE document_id=$1 GROUP BY document_id
		) counts WHERE d.id=$1`, documentID)
	return err
}

func (r *GeologicalJournalRepository) GetDocumentPageAsset(ctx context.Context, pageID, kind string) (string, error) {
	column := map[string]string{"original": "original_asset_path", "oriented": "oriented_asset_path", "preprocessed": "preprocessed_asset_path"}[kind]
	if column == "" {
		return "", ErrNotFound
	}
	var path *string
	err := r.pool.QueryRow(ctx, `SELECT `+column+` FROM geological_journal_document_pages WHERE id=$1`, pageID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) || path == nil || *path == "" {
		return "", ErrNotFound
	}
	return *path, err
}

func (r *GeologicalJournalRepository) GetDocumentPageAssetForDocument(ctx context.Context, documentID, pageID, kind string) (string, error) {
	column := map[string]string{"original": "original_asset_path", "oriented": "oriented_asset_path", "preprocessed": "preprocessed_asset_path"}[kind]
	if column == "" {
		return "", ErrNotFound
	}
	var path *string
	err := r.pool.QueryRow(ctx, `SELECT `+column+` FROM geological_journal_document_pages WHERE id=$1 AND document_id=$2`, pageID, documentID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) || path == nil || *path == "" {
		return "", ErrNotFound
	}
	return *path, err
}

func (r *GeologicalJournalRepository) SaveDocumentLLMResult(ctx context.Context, documentID, pageID, userID, mode string, result json.RawMessage, modelUsed string) (*model.GeologicalJournalDocumentLLMResult, error) {
	var item model.GeologicalJournalDocumentLLMResult
	err := r.pool.QueryRow(ctx, `
		INSERT INTO geological_journal_document_llm_results(document_id,page_id,user_id,mode,result,model_used)
		VALUES($1,$2,$3,$4,$5,$6)
		RETURNING id,document_id,page_id,mode,result,model_used,created_at`, documentID, pageID, userID, mode, result, modelUsed).Scan(
		&item.ID, &item.DocumentID, &item.PageID, &item.Mode, &item.Result, &item.ModelUsed, &item.CreatedAt)
	return &item, err
}

func (r *GeologicalJournalRepository) CreateChatSession(ctx context.Context, documentID, userID string, pages []int, neighbors bool) (string, error) {
	raw, err := json.Marshal(pages)
	if err != nil {
		return "", err
	}
	var id string
	err = r.pool.QueryRow(ctx, `INSERT INTO geological_journal_document_chat_sessions(document_id,user_id,page_numbers,include_neighbors) VALUES($1,$2,$3,$4) RETURNING id`, documentID, userID, raw, neighbors).Scan(&id)
	return id, err
}

func (r *GeologicalJournalRepository) AddChatMessage(ctx context.Context, sessionID, role, content string, sources json.RawMessage, confidence string) (*model.GeologicalJournalDocumentChatMessage, error) {
	var item model.GeologicalJournalDocumentChatMessage
	err := r.pool.QueryRow(ctx, `
		INSERT INTO geological_journal_document_chat_messages(session_id,role,content,sources,confidence)
		VALUES($1,$2,$3,$4,$5) RETURNING id,role,content,sources,confidence,created_at`, sessionID, role, content, sources, confidence).Scan(
		&item.ID, &item.Role, &item.Content, &item.Sources, &item.Confidence, &item.CreatedAt)
	return &item, err
}

func (r *GeologicalJournalRepository) GetChatSession(ctx context.Context, id, userID string) (string, string, bool, error) {
	var documentID string
	var raw []byte
	var neighbors bool
	err := r.pool.QueryRow(ctx, `SELECT document_id,page_numbers,include_neighbors FROM geological_journal_document_chat_sessions WHERE id=$1 AND user_id=$2`, id, userID).Scan(&documentID, &raw, &neighbors)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", false, ErrNotFound
	}
	return documentID, string(raw), neighbors, err
}

func (r *GeologicalJournalRepository) DocumentPageAsset(ctx context.Context, pageID string) (string, string, string, error) {
	var original, oriented, preprocessed *string
	err := r.pool.QueryRow(ctx, `SELECT original_asset_path,oriented_asset_path,preprocessed_asset_path FROM geological_journal_document_pages WHERE id=$1`, pageID).Scan(&original, &oriented, &preprocessed)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", ErrNotFound
	}
	return derefString(original), derefString(oriented), derefString(preprocessed), err
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
