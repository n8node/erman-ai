package service

import (
	"encoding/json"
	"sort"

	"github.com/erman-ai/erman-ai/internal/model"
)

func matchLegalRisks(risks []model.LegalRisk, findings model.LegalScanFindings, features model.LegalScanSiteFeatures) []model.LegalRisk {
	normalized := findings
	normalizeFindingsForFeatures(&normalized, features)

	flags := legalScanFlagsMap(features)
	findingsMap := legalScanFindingsMap(normalized)

	var matched []model.LegalRisk
	for _, risk := range risks {
		if !risk.IsActive {
			continue
		}
		if !riskApplicableForFeatures(risk.RiskID, features) {
			continue
		}
		if legalRiskTriggered(risk, findingsMap, flags) {
			matched = append(matched, risk)
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		si := severityOrder(matched[i].Severity)
		sj := severityOrder(matched[j].Severity)
		if si != sj {
			return si < sj
		}
		return matched[i].SortOrder < matched[j].SortOrder
	})
	return matched
}

func legalRiskTriggered(risk model.LegalRisk, findings, flags map[string]bool) bool {
	var tf map[string]bool
	if len(risk.TriggerFindings) > 0 {
		_ = json.Unmarshal(risk.TriggerFindings, &tf)
	}
	for key, expected := range tf {
		if findings[key] != expected {
			return false
		}
	}

	var tfl map[string]bool
	if len(risk.TriggerFlags) > 0 {
		_ = json.Unmarshal(risk.TriggerFlags, &tfl)
	}
	for key, expected := range tfl {
		if flags[key] != expected {
			return false
		}
	}
	return true
}

func legalScanFlagsMap(f model.LegalScanSiteFeatures) map[string]bool {
	return map[string]bool{
		"has_forms":              f.Forms,
		"has_traffic_from_ads":   f.TrafficFromAds,
		"has_online_sales":       f.OnlineSales,
		"has_foreign_services":   f.ForeignServices,
	}
}

func legalScanFindingsMap(f model.LegalScanFindings) map[string]bool {
	return map[string]bool{
		"ssl":                 f.SSL,
		"privacy_policy":      f.PrivacyPolicy,
		"cookie_banner":       f.CookieBanner,
		"cookie_policy":       f.CookiePolicy,
		"form_consent":        f.FormConsent,
		"requisites_inn":      f.RequisitesINN,
		"contacts":            f.Contacts,
		"offer":               f.Offer,
		"terms":               f.Terms,
		"consent_withdrawal":  f.ConsentWithdrawal,
		"ad_marking":          f.AdMarking,
		"has_trackers":        f.HasTrackers,
		"forms_collect_pd":    f.FormsCollectPD,
		"forms_unencrypted":   f.FormsUnencrypted,
		"foreign_trackers":    f.ForeignTrackers,
	}
}

func severityOrder(s string) int {
	switch s {
	case "high":
		return 0
	case "medium":
		return 1
	default:
		return 2
	}
}

func riskApplicableForFeatures(riskID string, features model.LegalScanSiteFeatures) bool {
	switch riskID {
	case "no_consent", "no_rkn_notice", "data_leak_exposure":
		return features.Forms
	case "no_ad_marking":
		return features.TrafficFromAds
	case "no_offer":
		return features.OnlineSales
	case "no_foreign_loc":
		return features.ForeignServices
	case "no_cookie_banner":
		return features.Forms || features.ForeignServices
	default:
		return true
	}
}

func buildRiskTablePayload(risks []model.LegalRisk) []map[string]any {
	out := make([]map[string]any, 0, len(risks))
	for _, r := range risks {
		item := map[string]any{
			"risk_id":      r.RiskID,
			"title":        r.TitleRU,
			"what":         r.WhatRU,
			"article":      r.Article,
			"fine_text":    r.FineTextRU,
			"severity":     r.Severity,
			"how_to_fix":   r.HowToFixRU,
			"is_turnover":  r.IsTurnoverFine,
		}
		if r.FineMin != nil {
			item["fine_min"] = *r.FineMin
		}
		if r.FineMax != nil {
			item["fine_max"] = *r.FineMax
		}
		out = append(out, item)
	}
	return out
}
