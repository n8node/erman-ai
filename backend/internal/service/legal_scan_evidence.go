package service

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

var (
	reHref        = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)
	reHrefText    = regexp.MustCompile(`(?i)<a[^>]+href\s*=\s*["']([^"']+)["'][^>]*>([^<]{0,120})`)
	reINNFind     = regexp.MustCompile(`\b(\d{10}|\d{12})\b`)
	reOGRNFind    = regexp.MustCompile(`\b(\d{13}|\d{15})\b`)
	reEmailFind   = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	rePhoneFind   = regexp.MustCompile(`(?:\+7|8)[\s\-]?\(?\d{3}\)?[\s\-]?\d{3}[\s\-]?\d{2}[\s\-]?\d{2}`)
	reEridFind    = regexp.MustCompile(`(?i)(erid[=:][\w\-]+|data-erid[^"'\s>]+)`)
	reTrackerSigs = []struct {
		id  string
		sig string
	}{
		{"Яндекс.Метрика", "mc.yandex.ru"},
		{"Google Analytics / GTM", "googletagmanager.com"},
		{"Meta Pixel", "connect.facebook.net"},
		{"VK Pixel", "vk.com/js/api/openapi.js"},
	}
)

type scanEvidence struct {
	PrivacyURLs      []string
	CookiePolicyURLs []string
	OfferURLs        []string
	TermsURLs        []string
	INNs             []string
	OGRNs            []string
	Emails           []string
	Phones           []string
	Trackers         []string
	EridTokens       []string
}

func extractScanEvidence(html, baseURL string) scanEvidence {
	lower := strings.ToLower(html)
	ev := scanEvidence{
		INNs:       uniqueStrings(reINNFind.FindAllString(html, 5)),
		OGRNs:      uniqueStrings(reOGRNFind.FindAllString(html, 5)),
		Emails:     uniqueStrings(reEmailFind.FindAllString(html, 5)),
		Phones:     uniqueStrings(rePhoneFind.FindAllString(html, 5)),
		EridTokens: uniqueStrings(reEridFind.FindAllString(html, 5)),
	}

	for _, m := range reHrefText.FindAllStringSubmatch(html, -1) {
		href := resolveScanURL(m[1], baseURL)
		text := strings.ToLower(strings.TrimSpace(m[2]))
		if href == "" {
			continue
		}
		if linkMatchesPrivacy(href, text) {
			ev.PrivacyURLs = appendUnique(ev.PrivacyURLs, href)
		}
		if linkMatchesCookiePolicy(href, text) {
			ev.CookiePolicyURLs = appendUnique(ev.CookiePolicyURLs, href)
		}
		if linkMatchesOffer(href, text) {
			ev.OfferURLs = appendUnique(ev.OfferURLs, href)
		}
		if linkMatchesTerms(href, text) {
			ev.TermsURLs = appendUnique(ev.TermsURLs, href)
		}
	}

	for _, h := range reHref.FindAllStringSubmatch(html, -1) {
		href := resolveScanURL(h[1], baseURL)
		if href == "" {
			continue
		}
		path := strings.ToLower(href)
		if strings.Contains(path, "/privacy") || strings.Contains(path, "/policy") {
			ev.PrivacyURLs = appendUnique(ev.PrivacyURLs, href)
		}
		if strings.Contains(path, "/offer") {
			ev.OfferURLs = appendUnique(ev.OfferURLs, href)
		}
		if strings.Contains(path, "/terms") {
			ev.TermsURLs = appendUnique(ev.TermsURLs, href)
		}
		if strings.Contains(path, "cookie") || strings.Contains(path, "куки") {
			ev.CookiePolicyURLs = appendUnique(ev.CookiePolicyURLs, href)
		}
	}

	for _, t := range reTrackerSigs {
		if strings.Contains(lower, strings.ToLower(t.sig)) {
			ev.Trackers = appendUnique(ev.Trackers, t.id)
		}
	}

	return ev
}

func linkMatchesPrivacy(href, text string) bool {
	combined := strings.ToLower(href + " " + text)
	return strings.Contains(combined, "политик") ||
		strings.Contains(combined, "конфиденциальн") ||
		strings.Contains(combined, "персональн") ||
		strings.Contains(combined, "privacy")
}

func linkMatchesCookiePolicy(href, text string) bool {
	combined := strings.ToLower(href + " " + text)
	return (strings.Contains(combined, "cookie") || strings.Contains(combined, "куки")) &&
		(strings.Contains(combined, "политик") || strings.Contains(combined, "policy"))
}

func linkMatchesOffer(href, text string) bool {
	combined := strings.ToLower(href + " " + text)
	return strings.Contains(combined, "оферт") || strings.Contains(combined, "/offer")
}

func linkMatchesTerms(href, text string) bool {
	combined := strings.ToLower(href + " " + text)
	return strings.Contains(combined, "пользовательское") || strings.Contains(combined, "/terms")
}

func resolveScanURL(href, base string) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(strings.ToLower(href), "javascript:") {
		return ""
	}
	baseParsed, err := url.Parse(base)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return baseParsed.ResolveReference(ref).String()
}

func uniqueStrings(in []string) []string {
	return appendUnique(nil, in...)
}

func appendUnique(dst []string, items ...string) []string {
	seen := make(map[string]struct{}, len(dst))
	for _, s := range dst {
		seen[s] = struct{}{}
	}
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		dst = append(dst, s)
	}
	return dst
}

func isLegalScanCheckApplicable(key string, features model.LegalScanSiteFeatures, f model.LegalScanFindings) bool {
	switch key {
	case "consent", "formenc", "withdraw":
		return features.Forms
	case "offer":
		return features.OnlineSales
	case "admark":
		return features.TrafficFromAds
	case "cookie", "cookiepol", "trackers":
		return features.Forms || features.ForeignServices
	default:
		return true
	}
}