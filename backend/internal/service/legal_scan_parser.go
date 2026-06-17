package service

import (
	"regexp"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

var (
	reINN       = regexp.MustCompile(`\b\d{10}\b|\b\d{12}\b`)
	reOGRN      = regexp.MustCompile(`\b\d{13}\b|\b\d{15}\b`)
	reEmail     = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	rePhone     = regexp.MustCompile(`(?:\+7|8)[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}`)
	reErid      = regexp.MustCompile(`(?i)(erid[=:][\w\-]+|data-erid|token=[\w\-]{10,})`)
	reAdLabel   = regexp.MustCompile(`(?i)\bреклама\b`)
	reFormHTTP  = regexp.MustCompile(`(?i)<form[^>]+action=["']http://`)
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

	trackers := detectTrackers(lower)
	hasTrackers := len(trackers) > 0
	foreignTrackers := detectForeignTrackers(lower)

	findings := model.LegalScanFindings{
		SSL:               page.HTTPS,
		PrivacyPolicy:     hasPrivacyPolicy(lower),
		CookieBanner:      hasCookieBanner(lower),
		CookiePolicy:      hasCookiePolicy(lower),
		FormConsent:       hasFormConsent(lower),
		RequisitesINN:     reINN.MatchString(html) || reOGRN.MatchString(html),
		Contacts:          reEmail.MatchString(html) || rePhone.MatchString(html),
		Offer:             hasOffer(lower),
		Terms:             hasTerms(lower),
		ConsentWithdrawal: hasConsentWithdrawal(lower),
		AdMarking:         reErid.MatchString(html) || reAdLabel.MatchString(html),
		Trackers:          trackers,
		HasTrackers:       hasTrackers,
		FormsCollectPD:    hasPDForms(lower),
		FormsUnencrypted:  reFormHTTP.MatchString(html),
		ForeignTrackers:   foreignTrackers,
	}

	if !features.Forms {
		findings.FormsCollectPD = false
	}

	checklist := buildLegalScanChecklist(findings, features)

	return model.LegalScanLayer1{
		FinalURL:  page.URL,
		Findings:  findings,
		Checklist: checklist,
	}
}

func buildLegalScanChecklist(f model.LegalScanFindings, features model.LegalScanSiteFeatures) []model.LegalScanCheckItem {
	items := make([]model.LegalScanCheckItem, 0, len(legalScanChecks))
	for _, c := range legalScanChecks {
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
			if !f.ConsentWithdrawal {
				status = "risk"
			}
		case "admark":
			if features.TrafficFromAds && !f.AdMarking {
				status = "risk"
			}
		case "trackers":
			if f.HasTrackers {
				status = "ok"
			}
		case "formenc":
			if f.FormsUnencrypted {
				status = "risk"
			}
		}
		items = append(items, model.LegalScanCheckItem{Key: c.Key, Label: c.Label, Status: status})
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
		{"yandex_metrika", "mc.yandex.ru"},
		{"ga4", "googletagmanager.com"},
		{"meta_pixel", "connect.facebook.net"},
		{"vk_pixel", "vk.com/js/api/openapi.js"},
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
