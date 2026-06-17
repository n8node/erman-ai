package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

const legalScanDisclaimerRU = "Отчёт носит информационный характер и не является юридической консультацией. Рекомендуем подтвердить с юристом."

func validateLegalScanOutput(raw *model.LegalScanLLMOutput, matched []model.LegalRisk) model.LegalScanOutput {
	byID := make(map[string]model.LegalRisk, len(matched))
	for _, r := range matched {
		byID[r.RiskID] = r
	}

	var risks []model.LegalScanRiskItem
	var fineMin, fineMax int
	var turnoverNote string

	for _, item := range raw.Risks {
		ref, ok := byID[item.RiskID]
		if !ok {
			continue
		}
		risks = append(risks, model.LegalScanRiskItem{
			RiskID:      ref.RiskID,
			Title:       ref.TitleRU,
			Explanation: strings.TrimSpace(item.Explanation),
			Article:     ref.Article,
			FineText:    ref.FineTextRU,
			Severity:    ref.Severity,
			HowToFix:    mergeHowToFix(ref.HowToFixRU, item.HowToFix),
		})
		if ref.IsTurnoverFine || ref.IsContextOnly {
			if turnoverNote == "" {
				turnoverNote = ref.FineTextRU
			}
			continue
		}
		if ref.FineMin != nil {
			fineMin += *ref.FineMin
		}
		if ref.FineMax != nil {
			fineMax += *ref.FineMax
		}
	}

	if len(risks) == 0 {
		for _, ref := range matched {
			if ref.IsContextOnly || ref.IsTurnoverFine {
				if turnoverNote == "" {
					turnoverNote = ref.FineTextRU
				}
				continue
			}
			risks = append(risks, model.LegalScanRiskItem{
				RiskID:      ref.RiskID,
				Title:       ref.TitleRU,
				Explanation: ref.WhatRU,
				Article:     ref.Article,
				FineText:    ref.FineTextRU,
				Severity:    ref.Severity,
				HowToFix:    ref.HowToFixRU,
			})
			if ref.FineMin != nil {
				fineMin += *ref.FineMin
			}
			if ref.FineMax != nil {
				fineMax += *ref.FineMax
			}
		}
	}

	disclaimer := strings.TrimSpace(raw.Disclaimer)
	if disclaimer == "" {
		disclaimer = legalScanDisclaimerRU
	}

	return model.LegalScanOutput{
		Summary: model.LegalScanSummary{
			RisksCount:       len(risks),
			FineMinTotal:     fineMin,
			FineMaxTotal:     fineMax,
			TurnoverFineNote: turnoverNote,
		},
		Risks:        risks,
		IndustryNote: strings.TrimSpace(raw.IndustryNote),
		Disclaimer:   disclaimer,
	}
}

func mergeHowToFix(base, fromLLM string) string {
	fromLLM = strings.TrimSpace(fromLLM)
	if fromLLM == "" || fromLLM == base {
		return base
	}
	if strings.HasPrefix(fromLLM, base) {
		return fromLLM
	}
	return base + " " + fromLLM
}

func parseLegalScanLLMOutput(content string) (*model.LegalScanLLMOutput, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var out model.LegalScanLLMOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	return &out, nil
}
