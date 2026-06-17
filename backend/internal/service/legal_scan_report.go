package service

import (
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

var checkKeyToRiskID = map[string]string{
	"ssl":     "no_ssl",
	"privacy": "no_privacy_policy",
	"cookie":  "no_cookie_banner",
	"consent": "no_consent",
	"req":     "no_requisites",
	"offer":   "no_offer",
	"admark":  "no_ad_marking",
}

var legalScanIndustryNotes = map[string]string{
	"medicine":  "Для медицинских услуг могут действовать дополнительные требования к обработке медданных и рекламе — уточните у профильного юриста.",
	"finance":   "Для финансовых услуг возможны отраслевые требования ЦБ и 115-ФЗ — проверьте отдельно.",
	"education": "Для онлайн-образования могут применяться дополнительные правила к договорам и обработке данных учащихся.",
}

func buildDeterministicLegalScanReport(
	layer1 model.LegalScanLayer1,
	matched []model.LegalRisk,
	industry string,
) model.LegalScanOutput {
	checklistByRisk := make(map[string]model.LegalScanCheckItem, len(layer1.Checklist))
	for _, item := range layer1.Checklist {
		if id := checkKeyToRiskID[item.Key]; id != "" {
			checklistByRisk[id] = item
		}
	}

	var risks []model.LegalScanRiskItem
	var fineMin, fineMax int
	var turnoverNote string

	for _, ref := range matched {
		if ref.IsContextOnly {
			if ref.IsTurnoverFine && turnoverNote == "" {
				turnoverNote = ref.FineTextRU
			}
			continue
		}

		item, hasItem := checklistByRisk[ref.RiskID]
		explanation := ref.WhatRU
		var pageURLs, foundData []string
		evidence := ""
		if hasItem {
			explanation = buildRiskExplanation(ref, item)
			pageURLs = item.PageURLs
			foundData = item.FoundData
			evidence = item.Evidence
		} else {
			explanation = buildExtraRiskExplanation(ref, layer1.Findings, layer1.FinalURL)
		}

		howToFix := ref.HowToFixRU
		if note := industryHowToFixNote(industry, ref.RiskID); note != "" {
			howToFix = strings.TrimSpace(ref.HowToFixRU + " " + note)
		}

		risks = append(risks, model.LegalScanRiskItem{
			RiskID:      ref.RiskID,
			Title:       ref.TitleRU,
			Explanation: explanation,
			Article:     ref.Article,
			FineText:    ref.FineTextRU,
			Severity:    ref.Severity,
			HowToFix:    howToFix,
			Evidence:    evidence,
			PageURLs:    pageURLs,
			FoundData:   foundData,
		})

		if ref.IsTurnoverFine {
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

	industryNote := legalScanIndustryNotes[industry]

	risksCount := 0
	for _, item := range layer1.Checklist {
		if item.Status == "risk" {
			risksCount++
		}
	}
	for _, ref := range matched {
		if ref.IsContextOnly {
			continue
		}
		if _, hasCheck := checklistByRisk[ref.RiskID]; !hasCheck {
			risksCount++
		}
	}

	return model.LegalScanOutput{
		Summary: model.LegalScanSummary{
			RisksCount:       risksCount,
			FineMinTotal:     fineMin,
			FineMaxTotal:     fineMax,
			TurnoverFineNote: turnoverNote,
		},
		Risks:        risks,
		IndustryNote: industryNote,
		Disclaimer:   legalScanDisclaimerRU,
	}
}

func buildRiskExplanation(ref model.LegalRisk, item model.LegalScanCheckItem) string {
	if strings.TrimSpace(item.Evidence) != "" {
		return item.Evidence
	}
	return ref.WhatRU
}

func buildExtraRiskExplanation(ref model.LegalRisk, f model.LegalScanFindings, pageURL string) string {
	switch ref.RiskID {
	case "no_rkn_notice":
		return "На сайте есть формы сбора персональных данных — проверьте, подано ли уведомление в Роскомнадзор."
	case "no_foreign_loc":
		if len(f.Trackers) > 0 {
			return fmt.Sprintf("Обнаружены иностранные трекеры на странице %s: %s.", pageURL, strings.Join(f.Trackers, ", "))
		}
		return "На сайте используются иностранные сервисы — проверьте локализацию персональных данных в РФ."
	default:
		return ref.WhatRU
	}
}

func industryHowToFixNote(industry, riskID string) string {
	switch industry {
	case "medicine":
		if riskID == "no_privacy_policy" || riskID == "no_consent" {
			return "Учтите требования к медицинским данным."
		}
	case "finance":
		if riskID == "no_privacy_policy" {
			return "Учтите требования к финансовым данным клиентов."
		}
	case "education":
		if riskID == "no_offer" {
			return "Для онлайн-курсов добавьте условия доступа и возврата."
		}
	}
	return ""
}

func attachCheckEvidence(
	key, status string,
	ev scanEvidence,
	f model.LegalScanFindings,
	pageURL string,
) (evidence string, pageURLs, foundData []string) {
	switch key {
	case "ssl":
		if status == "ok" {
			evidence = fmt.Sprintf("Сайт %s доступен по HTTPS.", pageURL)
		} else {
			evidence = fmt.Sprintf("Сайт %s открывается без HTTPS — передача данных не шифруется.", pageURL)
		}
	case "privacy":
		pageURLs = ev.PrivacyURLs
		if status == "ok" {
			if len(pageURLs) > 0 {
				evidence = "Найдена политика обработки персональных данных."
			} else {
				evidence = "На странице есть упоминание политики конфиденциальности."
			}
		} else {
			evidence = "Ссылка на политику обработки ПД не найдена на проверенной странице."
		}
	case "cookie":
		foundData = ev.Trackers
		if status == "risk" {
			if len(foundData) > 0 {
				evidence = fmt.Sprintf("Обнаружены трекеры (%s), cookie-баннер не найден на %s.", strings.Join(foundData, ", "), pageURL)
			} else {
				evidence = fmt.Sprintf("Cookie-баннер не обнаружен на %s.", pageURL)
			}
		} else {
			evidence = "Cookie-баннер обнаружен на странице."
		}
	case "cookiepol":
		pageURLs = ev.CookiePolicyURLs
		if status == "ok" && len(pageURLs) > 0 {
			evidence = "Найдена отдельная политика cookie."
		} else if status == "ok" {
			evidence = "Упоминание политики cookie найдено на странице."
		} else {
			evidence = "Отдельная политика cookie не найдена."
		}
	case "consent":
		if status == "ok" {
			evidence = "У форм сбора данных есть чекбокс согласия на обработку ПД."
		} else {
			evidence = "Формы сбора персональных данных без явного согласия на обработку."
		}
	case "req":
		foundData = appendUnique(nil, ev.INNs...)
		foundData = appendUnique(foundData, ev.OGRNs...)
		if status == "ok" && len(foundData) > 0 {
			evidence = fmt.Sprintf("На странице найдены реквизиты: %s.", strings.Join(foundData, ", "))
		} else if status == "ok" {
			evidence = "Реквизиты продавца указаны на странице."
		} else {
			evidence = "ИНН или ОГРН не найдены на проверенной странице."
		}
	case "contacts":
		foundData = appendUnique(nil, ev.Emails...)
		foundData = appendUnique(foundData, ev.Phones...)
		if status == "ok" && len(foundData) > 0 {
			evidence = fmt.Sprintf("Контакты на странице: %s.", strings.Join(foundData, ", "))
		} else if status == "ok" {
			evidence = "Контактные данные указаны на странице."
		} else {
			evidence = "Email или телефон для связи не найдены на проверенной странице."
		}
	case "offer":
		pageURLs = ev.OfferURLs
		if status == "ok" && len(pageURLs) > 0 {
			evidence = "Найдена публичная оферта."
		} else if status == "ok" {
			evidence = "Условия продажи/оферта упоминаются на странице."
		} else {
			evidence = "Публичная оферта не найдена."
		}
	case "terms":
		pageURLs = ev.TermsURLs
		if status == "ok" && len(pageURLs) > 0 {
			evidence = "Найдено пользовательское соглашение."
		} else if status == "ok" {
			evidence = "Пользовательское соглашение упоминается на странице."
		} else {
			evidence = "Пользовательское соглашение не найдено."
		}
	case "withdraw":
		if status == "ok" {
			evidence = "На сайте есть способ отозвать согласие или запросить удаление данных."
		} else {
			evidence = "Не найден механизм отзыва согласия или удаления персональных данных."
		}
	case "admark":
		foundData = ev.EridTokens
		if status == "ok" && len(foundData) > 0 {
			evidence = fmt.Sprintf("Маркировка рекламы найдена: %s.", strings.Join(foundData, ", "))
		} else if status == "ok" {
			evidence = "На странице есть пометка «Реклама» или токен erid."
		} else {
			evidence = "Маркировка рекламы (erid или пометка «Реклама») не обнаружена."
		}
	case "trackers":
		foundData = ev.Trackers
		if len(foundData) > 0 {
			evidence = fmt.Sprintf("Обнаружены трекеры: %s.", strings.Join(foundData, ", "))
		} else {
			evidence = "Сторонние трекеры аналитики не обнаружены."
		}
	case "formenc":
		if status == "ok" {
			evidence = "Формы отправляются по HTTPS."
		} else {
			evidence = "Найдена форма с action по незашифрованному HTTP."
		}
	}
	return evidence, pageURLs, foundData
}

func normalizeFindingsForFeatures(f *model.LegalScanFindings, features model.LegalScanSiteFeatures) {
	if !features.Forms {
		f.FormsCollectPD = false
		f.FormConsent = true
		f.FormsUnencrypted = false
		f.ConsentWithdrawal = true
	}
	if !features.TrafficFromAds {
		f.AdMarking = true
	}
	if !features.OnlineSales {
		f.Offer = true
	}
	if !features.ForeignServices {
		f.ForeignTrackers = false
	}
}
