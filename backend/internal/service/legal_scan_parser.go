package service

import (
	"regexp"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

var (
	reAdLabel  = regexp.MustCompile(`(?i)\bреклама\b`)
	reFormHTTP = regexp.MustCompile(`(?i)<form[^>]+action=["']http://`)
)

var legalScanChecks = []struct {
	Key   string
	Label string
}{
	{"ssl", "SSL-сертификат и HTTPS"},
	{"privacy", "Политика обработки ПД"},
	{"cookie", "Cookie-баннер"},
	{"cookiepol", "Отдельная политика cookie"},
	{"consent", "Согласие у форм"},
	{"req", "Реквизиты (ИНН/ОГРН)"},
	{"contacts", "Контактные данные"},
	{"offer", "Публичная оферта"},
	{"terms", "Пользовательское соглашение"},
	{"withdraw", "Отзыв согласия / удаление данных"},
	{"admark", "Маркировка рекламы (erid)"},
	{"trackers", "Трекеры и аналитика"},
	{"formenc", "Шифрование форм"},
}

func runLegalScanLayer1(page *fetchedPage, features model.LegalScanSiteFeatures) model.LegalScanLayer1 {
	html := page.HTML
	lower := strings.ToLower(html)
	ev := extractScanEvidence(html, page.URL)

	trackers := detectTrackers(lower)
	hasTrackers := len(trackers) > 0
	foreignTrackers := detectForeignTrackers(lower)

	findings := model.LegalScanFindings{
		SSL:               page.HTTPS,
		PrivacyPolicy:     hasPrivacyPolicy(lower),
		CookieBanner:      hasCookieBanner(lower),
		CookiePolicy:      hasCookiePolicy(lower),
		FormConsent:       hasFormConsent(lower),
		RequisitesINN:     len(ev.INNs) > 0 || len(ev.OGRNs) > 0,
		Contacts:          len(ev.Emails) > 0 || len(ev.Phones) > 0,
		Offer:             hasOffer(lower),
		Terms:             hasTerms(lower),
		ConsentWithdrawal: hasConsentWithdrawal(lower),
		AdMarking:         len(ev.EridTokens) > 0 || reAdLabel.MatchString(html),
		Trackers:          trackers,
		HasTrackers:       hasTrackers,
		FormsCollectPD:    hasPDForms(lower),
		FormsUnencrypted:  reFormHTTP.MatchString(html),
		ForeignTrackers:   foreignTrackers,
	}

	normalizeFindingsForFeatures(&findings, features)
	checklist := buildLegalScanChecklist(findings, features, ev, page.URL)

	return model.LegalScanLayer1{
		FinalURL:  page.URL,
		Findings:  findings,
		Checklist: checklist,
	}
}

func buildLegalScanChecklist(
	f model.LegalScanFindings,
	features model.LegalScanSiteFeatures,
	ev scanEvidence,
	pageURL string,
) []model.LegalScanCheckItem {
	items := make([]model.LegalScanCheckItem, 0, len(legalScanChecks))
	for _, c := range legalScanChecks {
		if !isLegalScanCheckApplicable(c.Key, features, f) {
			continue
		}

		status := "ok"
		switch c.Key {
		case "ssl":
			if !f.SSL {
				status = "risk"
			}
		case "privacy":
			if !f.PrivacyPolicy {
				status = "risk"
			}
		case "cookie":
			if !f.CookieBanner && f.HasTrackers {
				status = "risk"
			}
		case "cookiepol":
			if !f.CookiePolicy {
				status = "risk"
			}
		case "consent":
			if features.Forms && f.FormsCollectPD && !f.FormConsent {
				status = "risk"
			}
		case "req":
			if !f.RequisitesINN {
				status = "risk"
			}
		case "contacts":
			if !f.Contacts {
				status = "risk"
			}
		case "offer":
			if features.OnlineSales && !f.Offer {
				status = "risk"
			}
		case "terms":
			if !f.Terms {
				status = "risk"
			}
		case "withdraw":
			if features.Forms && !f.ConsentWithdrawal {
				status = "risk"
			}
		case "admark":
			if features.TrafficFromAds && !f.AdMarking {
				status = "risk"
			}
		case "trackers":
			status = "ok"
		case "formenc":
			if features.Forms && f.FormsUnencrypted {
				status = "risk"
			}
		}

		evidence, pageURLs, foundData := attachCheckEvidence(c.Key, status, ev, f, pageURL)
		items = append(items, model.LegalScanCheckItem{
			Key:       c.Key,
			Label:     c.Label,
			Status:    status,
			Evidence:  evidence,
			PageURLs:  pageURLs,
			FoundData: foundData,
		})
	}
	return items
}

func hasPrivacyPolicy(lower string) bool {
	patterns := []string{"политик", "конфиденциальн", "персональн", "/privacy", "/policy", "privacy policy"}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func hasCookieBanner(lower string) bool {
	patterns := []string{"cookiebot", "cookieyes", "cookie-consent", "cookie_banner", "файлы cookie", "файлы куки", "используем cookie", "используем куки"}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func hasCookiePolicy(lower string) bool {
	if strings.Contains(lower, "cookie") || strings.Contains(lower, "куки") {
		if strings.Contains(lower, "политик") || strings.Contains(lower, "policy") {
			return true
		}
	}
	return false
}

func hasFormConsent(lower string) bool {
	if !hasPDForms(lower) {
		return true
	}
	return strings.Contains(lower, "согласи") && strings.Contains(lower, "обработк")
}

func hasPDForms(lower string) bool {
	return strings.Contains(lower, "type=\"email\"") ||
		strings.Contains(lower, "type='email'") ||
		strings.Contains(lower, "type=\"tel\"") ||
		strings.Contains(lower, "<form")
}

func hasOffer(lower string) bool {
	return strings.Contains(lower, "оферт") || strings.Contains(lower, "/offer")
}

func hasTerms(lower string) bool {
	return strings.Contains(lower, "пользовательское соглашение") || strings.Contains(lower, "/terms")
}

func hasConsentWithdrawal(lower string) bool {
	patterns := []string{"отозвать согласие", "удалить данные", "удаление данных", "запрос на удаление"}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func detectTrackers(lower string) []string {
	var out []string
	checks := []struct {
		id  string
		sig string
	}{
		{"Яндекс.Метрика", "mc.yandex.ru"},
		{"Google Analytics / GTM", "googletagmanager.com"},
		{"Meta Pixel", "connect.facebook.net"},
		{"VK Pixel", "vk.com/js/api/openapi.js"},
	}
	for _, c := range checks {
		if strings.Contains(lower, c.sig) {
			out = append(out, c.id)
		}
	}
	return out
}

func detectForeignTrackers(lower string) bool {
	foreign := []string{"googletagmanager.com", "google-analytics.com", "connect.facebook.net", "facebook.com/tr"}
	for _, s := range foreign {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}
