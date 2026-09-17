package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GeologicalJournalRepository struct {
	pool *pgxpool.Pool
}

func NewGeologicalJournalRepository(pool *pgxpool.Pool) *GeologicalJournalRepository {
	return &GeologicalJournalRepository{pool: pool}
}

func (r *GeologicalJournalRepository) HasExplicitAccess(ctx context.Context, userID string) (bool, error) {
	var enabled bool
	err := r.pool.QueryRow(ctx, `SELECT enabled FROM geological_journal_access WHERE user_id=$1`, userID).Scan(&enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return enabled, err
}

func (r *GeologicalJournalRepository) HasDocumentAccess(ctx context.Context, documentID, userID string, role string) (bool, error) {
	if role == "superadmin" {
		return true, nil
	}
	var allowed bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM geological_journal_documents d
			WHERE d.id=$1 AND (d.owner_id=$2 OR d.is_shared)
		)`, documentID, userID).Scan(&allowed)
	return allowed, err
}

func (r *GeologicalJournalRepository) SetAccess(ctx context.Context, userID string, enabled bool) error {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO geological_journal_access(user_id, enabled) VALUES($1,$2)
		ON CONFLICT(user_id) DO UPDATE SET enabled=EXCLUDED.enabled, updated_at=NOW()`, userID, enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GeologicalJournalRepository) ListAccessUsers(ctx context.Context) ([]model.GeologicalJournalAccessUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id,u.email,u.role,(u.role='superadmin' OR COALESCE(a.enabled,false))
		FROM users u LEFT JOIN geological_journal_access a ON a.user_id=u.id
		ORDER BY u.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.GeologicalJournalAccessUser
	for rows.Next() {
		var item model.GeologicalJournalAccessUser
		if err := rows.Scan(&item.ID, &item.Email, &item.Role, &item.HasAccess); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *GeologicalJournalRepository) CreatePage(ctx context.Context, userID, originalName, assetPath, contentType string, size int64, width, height int) (*model.GeologicalJournalPage, error) {
	return scanGeologicalPage(r.pool.QueryRow(ctx, `
		INSERT INTO geological_journal_pages(user_id,original_name,asset_path,content_type,size_bytes,width,height)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		RETURNING id,user_id,original_name,content_type,size_bytes,width,height,latest_result,created_at,updated_at`,
		userID, originalName, assetPath, contentType, size, width, height))
}

func (r *GeologicalJournalRepository) ListPages(ctx context.Context, userID string) ([]model.GeologicalJournalPage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,user_id,original_name,content_type,size_bytes,width,height,latest_result,created_at,updated_at
		FROM geological_journal_pages WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.GeologicalJournalPage
	for rows.Next() {
		item, err := scanGeologicalPage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *GeologicalJournalRepository) GetPage(ctx context.Context, pageID, userID string) (*model.GeologicalJournalPage, error) {
	page, err := scanGeologicalPage(r.pool.QueryRow(ctx, `
		SELECT id,user_id,original_name,content_type,size_bytes,width,height,latest_result,created_at,updated_at
		FROM geological_journal_pages WHERE id=$1 AND user_id=$2`, pageID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return page, err
}

func (r *GeologicalJournalRepository) PageAssetPath(ctx context.Context, pageID, userID string) (string, error) {
	var path string
	err := r.pool.QueryRow(ctx, `SELECT asset_path FROM geological_journal_pages WHERE id=$1 AND user_id=$2`, pageID, userID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return path, err
}

func (r *GeologicalJournalRepository) DeletePage(ctx context.Context, pageID, userID string) (string, error) {
	var path string
	err := r.pool.QueryRow(ctx, `DELETE FROM geological_journal_pages WHERE id=$1 AND user_id=$2 RETURNING asset_path`, pageID, userID).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return path, err
}

func (r *GeologicalJournalRepository) AttachRun(ctx context.Context, pageID, runID string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO geological_journal_page_runs(page_id,run_id) VALUES($1,$2)`, pageID, runID)
	return err
}

func (r *GeologicalJournalRepository) ListPageRuns(ctx context.Context, pageID, userID string) ([]model.ToolRun, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT tr.id,tr.user_id,tr.tool_slug,tr.plan_tier,tr.input,tr.output,tr.artifact_url,tr.tokens_used,
		       tr.model_used,tr.status,tr.error_msg,tr.created_at,tr.updated_at,tr.completed_at
		FROM tool_runs tr JOIN geological_journal_page_runs pr ON pr.run_id=tr.id
		JOIN geological_journal_pages p ON p.id=pr.page_id
		WHERE pr.page_id=$1 AND p.user_id=$2 ORDER BY tr.created_at DESC`, pageID, userID)
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

func (r *GeologicalJournalRepository) SaveResult(ctx context.Context, pageID, userID string, result json.RawMessage) (*model.GeologicalJournalResultVersion, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE geological_journal_pages SET latest_result=$3,updated_at=NOW() WHERE id=$1 AND user_id=$2`, pageID, userID, result)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	var v model.GeologicalJournalResultVersion
	err = tx.QueryRow(ctx, `
		INSERT INTO geological_journal_result_versions(page_id,user_id,result) VALUES($1,$2,$3)
		RETURNING id,page_id,user_id,result,created_at`, pageID, userID, result).
		Scan(&v.ID, &v.PageID, &v.UserID, &v.Result, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *GeologicalJournalRepository) ListVersions(ctx context.Context, pageID, userID string) ([]model.GeologicalJournalResultVersion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT v.id,v.page_id,v.user_id,v.result,v.created_at FROM geological_journal_result_versions v
		JOIN geological_journal_pages p ON p.id=v.page_id
		WHERE v.page_id=$1 AND p.user_id=$2 ORDER BY v.created_at DESC`, pageID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.GeologicalJournalResultVersion
	for rows.Next() {
		var v model.GeologicalJournalResultVersion
		if err := rows.Scan(&v.ID, &v.PageID, &v.UserID, &v.Result, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *GeologicalJournalRepository) GetSettings(ctx context.Context) (*model.GeologicalJournalSettingsRecord, error) {
	var raw []byte
	var rec model.GeologicalJournalSettingsRecord
	if err := r.pool.QueryRow(ctx, `SELECT config,updated_at FROM geological_journal_settings WHERE id=1`).Scan(&raw, &rec.UpdatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &rec.Settings); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *GeologicalJournalRepository) UpdateSettings(ctx context.Context, settings model.GeologicalJournalSettings) (*model.GeologicalJournalSettingsRecord, error) {
	raw, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	var out []byte
	var rec model.GeologicalJournalSettingsRecord
	err = r.pool.QueryRow(ctx, `UPDATE geological_journal_settings SET config=$1,updated_at=NOW() WHERE id=1 RETURNING config,updated_at`, raw).Scan(&out, &rec.UpdatedAt)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(out, &rec.Settings)
	return &rec, err
}

func (r *GeologicalJournalRepository) CreateExample(ctx context.Context, title, description, assetPath, contentType string, size int64, sortOrder int, published bool) (*model.GeologicalJournalExample, error) {
	return scanGeologicalExample(r.pool.QueryRow(ctx, `
		INSERT INTO geological_journal_examples(title,description,asset_path,content_type,size_bytes,sort_order,is_published)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		RETURNING id,title,description,content_type,size_bytes,sort_order,is_published,created_at,updated_at`,
		title, description, assetPath, contentType, size, sortOrder, published))
}

func (r *GeologicalJournalRepository) ListExamples(ctx context.Context, publishedOnly bool) ([]model.GeologicalJournalExample, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id,title,description,content_type,size_bytes,sort_order,is_published,created_at,updated_at
		FROM geological_journal_examples WHERE (NOT $1 OR is_published) ORDER BY sort_order,id`, publishedOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.GeologicalJournalExample
	for rows.Next() {
		item, err := scanGeologicalExample(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *GeologicalJournalRepository) ExampleAssetPath(ctx context.Context, id string, publishedOnly bool) (string, string, error) {
	var path, contentType string
	err := r.pool.QueryRow(ctx, `SELECT asset_path,content_type FROM geological_journal_examples WHERE id=$1 AND (NOT $2 OR is_published)`, id, publishedOnly).Scan(&path, &contentType)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return path, contentType, err
}

func (r *GeologicalJournalRepository) UpdateExample(ctx context.Context, id string, meta model.GeologicalJournalExampleMetadata) (*model.GeologicalJournalExample, error) {
	item, err := scanGeologicalExample(r.pool.QueryRow(ctx, `
		UPDATE geological_journal_examples SET title=$2,description=$3,sort_order=$4,is_published=$5,updated_at=NOW()
		WHERE id=$1 RETURNING id,title,description,content_type,size_bytes,sort_order,is_published,created_at,updated_at`,
		id, meta.Title, meta.Description, meta.SortOrder, meta.IsPublished))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *GeologicalJournalRepository) DeleteExample(ctx context.Context, id string) (string, error) {
	var path string
	err := r.pool.QueryRow(ctx, `DELETE FROM geological_journal_examples WHERE id=$1 RETURNING asset_path`, id).Scan(&path)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return path, err
}

func scanGeologicalPage(row pgx.Row) (*model.GeologicalJournalPage, error) {
	var item model.GeologicalJournalPage
	err := row.Scan(&item.ID, &item.UserID, &item.OriginalName, &item.ContentType, &item.SizeBytes,
		&item.Width, &item.Height, &item.LatestResult, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func scanGeologicalExample(row pgx.Row) (*model.GeologicalJournalExample, error) {
	var item model.GeologicalJournalExample
	err := row.Scan(&item.ID, &item.Title, &item.Description, &item.ContentType, &item.SizeBytes,
		&item.SortOrder, &item.IsPublished, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}
