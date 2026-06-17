package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LegalRiskRepository struct {
	pool *pgxpool.Pool
}

func NewLegalRiskRepository(pool *pgxpool.Pool) *LegalRiskRepository {
	return &LegalRiskRepository{pool: pool}
}

const legalRiskSelect = `
	SELECT risk_id, title_ru, title_en, what_ru, article, fine_text_ru, fine_min, fine_max,
	       severity, how_to_fix_ru, how_to_fix_en, trigger_findings, trigger_flags,
	       is_turnover_fine, is_context_only, is_active, sort_order, updated_at
	FROM legal_risks
`

func scanLegalRisk(row pgx.Row) (*model.LegalRisk, error) {
	var r model.LegalRisk
	err := row.Scan(
		&r.RiskID, &r.TitleRU, &r.TitleEN, &r.WhatRU, &r.Article, &r.FineTextRU,
		&r.FineMin, &r.FineMax, &r.Severity, &r.HowToFixRU, &r.HowToFixEN,
		&r.TriggerFindings, &r.TriggerFlags,
		&r.IsTurnoverFine, &r.IsContextOnly, &r.IsActive, &r.SortOrder, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *LegalRiskRepository) ListActive(ctx context.Context) ([]model.LegalRisk, error) {
	const q = legalRiskSelect + ` WHERE is_active = true ORDER BY sort_order, risk_id`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.LegalRisk
	for rows.Next() {
		item, err := scanLegalRisk(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *LegalRiskRepository) ListAll(ctx context.Context) ([]model.LegalRisk, error) {
	const q = legalRiskSelect + ` ORDER BY sort_order, risk_id`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.LegalRisk
	for rows.Next() {
		item, err := scanLegalRisk(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func (r *LegalRiskRepository) Get(ctx context.Context, riskID string) (*model.LegalRisk, error) {
	const q = legalRiskSelect + ` WHERE risk_id = $1`
	item, err := scanLegalRisk(r.pool.QueryRow(ctx, q, riskID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return item, err
}

func (r *LegalRiskRepository) Update(ctx context.Context, riskID string, req model.LegalRiskUpdateRequest) (*model.LegalRisk, error) {
	const q = `
		UPDATE legal_risks SET
			title_ru = $2, title_en = $3, what_ru = $4, article = $5, fine_text_ru = $6,
			fine_min = $7, fine_max = $8, severity = $9, how_to_fix_ru = $10, how_to_fix_en = $11,
			trigger_findings = $12, trigger_flags = $13,
			is_turnover_fine = $14, is_context_only = $15, is_active = $16, sort_order = $17,
			updated_at = NOW()
		WHERE risk_id = $1
		RETURNING risk_id, title_ru, title_en, what_ru, article, fine_text_ru, fine_min, fine_max,
		          severity, how_to_fix_ru, how_to_fix_en, trigger_findings, trigger_flags,
		          is_turnover_fine, is_context_only, is_active, sort_order, updated_at
	`
	return scanLegalRisk(r.pool.QueryRow(ctx, q,
		riskID, req.TitleRU, req.TitleEN, req.WhatRU, req.Article, req.FineTextRU,
		req.FineMin, req.FineMax, req.Severity, req.HowToFixRU, req.HowToFixEN,
		req.TriggerFindings, req.TriggerFlags,
		req.IsTurnoverFine, req.IsContextOnly, req.IsActive, req.SortOrder,
	))
}

func (r *LegalRiskRepository) Create(ctx context.Context, riskID string, req model.LegalRiskUpdateRequest) (*model.LegalRisk, error) {
	const q = `
		INSERT INTO legal_risks (
			risk_id, title_ru, title_en, what_ru, article, fine_text_ru, fine_min, fine_max,
			severity, how_to_fix_ru, how_to_fix_en, trigger_findings, trigger_flags,
			is_turnover_fine, is_context_only, is_active, sort_order
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		RETURNING risk_id, title_ru, title_en, what_ru, article, fine_text_ru, fine_min, fine_max,
		          severity, how_to_fix_ru, how_to_fix_en, trigger_findings, trigger_flags,
		          is_turnover_fine, is_context_only, is_active, sort_order, updated_at
	`
	return scanLegalRisk(r.pool.QueryRow(ctx, q,
		riskID, req.TitleRU, req.TitleEN, req.WhatRU, req.Article, req.FineTextRU,
		req.FineMin, req.FineMax, req.Severity, req.HowToFixRU, req.HowToFixEN,
		req.TriggerFindings, req.TriggerFlags,
		req.IsTurnoverFine, req.IsContextOnly, req.IsActive, req.SortOrder,
	))
}

func (r *LegalRiskRepository) Delete(ctx context.Context, riskID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM legal_risks WHERE risk_id = $1`, riskID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type LegalScanLLMSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewLegalScanLLMSettingsRepository(pool *pgxpool.Pool) *LegalScanLLMSettingsRepository {
	return &LegalScanLLMSettingsRepository{pool: pool}
}

func (r *LegalScanLLMSettingsRepository) Get(ctx context.Context) (*model.LegalScanLLMSettingsRecord, error) {
	const q = `SELECT config, updated_at FROM legal_scan_llm_settings WHERE id = 1`
	var raw []byte
	var rec model.LegalScanLLMSettingsRecord
	err := r.pool.QueryRow(ctx, q).Scan(&raw, &rec.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &rec.Config); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *LegalScanLLMSettingsRepository) Update(ctx context.Context, config model.LegalScanLLMStoredConfig) (*model.LegalScanLLMSettingsRecord, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	const q = `
		UPDATE legal_scan_llm_settings SET config = $1, updated_at = NOW() WHERE id = 1
		RETURNING config, updated_at
	`
	var out []byte
	var rec model.LegalScanLLMSettingsRecord
	err = r.pool.QueryRow(ctx, q, raw).Scan(&out, &rec.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(out, &rec.Config); err != nil {
		return nil, err
	}
	return &rec, nil
}
