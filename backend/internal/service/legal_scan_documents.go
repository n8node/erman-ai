package service

import "strings"

func filterEvidenceDocumentURLs(ev scanEvidence, pages []fetchedPage) scanEvidence {
	pageByURL := make(map[string]string, len(pages))
	for _, p := range pages {
		pageByURL[normalizeLegalScanURL(p.URL)] = p.HTML
	}

	ev.PrivacyURLs = filterDocumentURLs(ev.PrivacyURLs, pageByURL, pageConfirmsPrivacyDocument)
	ev.CookiePolicyURLs = filterDocumentURLs(ev.CookiePolicyURLs, pageByURL, pageConfirmsCookiePolicyDocument)
	ev.OfferURLs = filterDocumentURLs(ev.OfferURLs, pageByURL, pageConfirmsOfferDocument)
	ev.TermsURLs = filterDocumentURLs(ev.TermsURLs, pageByURL, pageConfirmsTermsDocument)
	return ev
}

func filterDocumentURLs(
	urls []string,
	pageByURL map[string]string,
	confirms func(string) bool,
) []string {
	out := make([]string, 0, len(urls))
	for _, u := range urls {
		html, ok := pageByURL[normalizeLegalScanURL(u)]
		if !ok {
			continue
		}
		if confirms(html) {
			out = appendUnique(out, u)
		}
	}
	return out
}

func pageConfirmsPrivacyDocument(html string) bool {
	lower := legalScanNormalizedText(html)
	return strings.Contains(lower, "privacy policy") ||
		strings.Contains(lower, "политика конфиденциальности") ||
		(strings.Contains(lower, "политик") &&
			strings.Contains(lower, "персональн") &&
			(strings.Contains(lower, "обработк") || strings.Contains(lower, "оператор") || strings.Contains(lower, "соглас")))
}

func pageConfirmsCookiePolicyDocument(html string) bool {
	lower := legalScanNormalizedText(html)
	hasCookie := strings.Contains(lower, "cookie") || strings.Contains(lower, "куки")
	return hasCookie && (strings.Contains(lower, "политик") || strings.Contains(lower, "соглас") || strings.Contains(lower, "обработк"))
}

func pageConfirmsOfferDocument(html string) bool {
	lower := legalScanNormalizedText(html)
	return strings.Contains(lower, "публичная оферт") ||
		strings.Contains(lower, "договор оферт") ||
		strings.Contains(lower, "договор-оферт") ||
		(strings.Contains(lower, "условия") &&
			(strings.Contains(lower, "продаж") || strings.Contains(lower, "оказания услуг") || strings.Contains(lower, "оплат") || strings.Contains(lower, "возврат")))
}

func pageConfirmsTermsDocument(html string) bool {
	lower := legalScanNormalizedText(html)
	return strings.Contains(lower, "пользовательское соглашение") ||
		strings.Contains(lower, "user agreement") ||
		strings.Contains(lower, "terms of use") ||
		strings.Contains(lower, "условия использования") ||
		strings.Contains(lower, "правила пользования")
}

func legalScanNormalizedText(html string) string {
	text := stripHTMLForTextExtraction(html)
	return strings.ToLower(strings.Join(strings.Fields(text), " "))
}
