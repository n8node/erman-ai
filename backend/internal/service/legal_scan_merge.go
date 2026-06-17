package service

import (
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

type mergedLegalScanData struct {
	combinedHTML string
	evidence     scanEvidence
	startURL     string
	startHTTPS   bool
}

func mergeLegalScanPages(pages []fetchedPage, startURL string, startHTTPS bool) mergedLegalScanData {
	var ev scanEvidence
	var parts []string
	for _, p := range pages {
		parts = append(parts, p.HTML)
		pageEv := extractScanEvidence(p.HTML, p.URL)
		ev = mergeScanEvidence(ev, pageEv)
	}
	return mergedLegalScanData{
		combinedHTML: strings.Join(parts, "\n"),
		evidence:     ev,
		startURL:     startURL,
		startHTTPS:   startHTTPS,
	}
}

func mergeScanEvidence(a, b scanEvidence) scanEvidence {
	return scanEvidence{
		PrivacyURLs:      appendUnique(a.PrivacyURLs, b.PrivacyURLs...),
		CookiePolicyURLs: appendUnique(a.CookiePolicyURLs, b.CookiePolicyURLs...),
		OfferURLs:        appendUnique(a.OfferURLs, b.OfferURLs...),
		TermsURLs:        appendUnique(a.TermsURLs, b.TermsURLs...),
		INNs:             appendUnique(a.INNs, b.INNs...),
		OGRNs:            appendUnique(a.OGRNs, b.OGRNs...),
		Emails:           appendUnique(a.Emails, b.Emails...),
		Phones:           appendUnique(a.Phones, b.Phones...),
		Trackers:         appendUnique(a.Trackers, b.Trackers...),
		EridTokens:       appendUnique(a.EridTokens, b.EridTokens...),
	}
}

func runLegalScanLayer1FromPages(
	pages []fetchedPage,
	startURL string,
	startHTTPS bool,
	features model.LegalScanSiteFeatures,
	meta *model.LegalScanCrawlMeta,
) model.LegalScanLayer1 {
	merged := mergeLegalScanPages(pages, startURL, startHTTPS)
	finalURL := startURL
	if len(pages) > 0 {
		finalURL = pages[0].URL
	}
	return buildLegalScanLayer1(merged, features, meta, finalURL)
}

func buildLegalScanLayer1(
	merged mergedLegalScanData,
	features model.LegalScanSiteFeatures,
	meta *model.LegalScanCrawlMeta,
	finalURL string,
) model.LegalScanLayer1 {
	html := merged.combinedHTML
	lower := strings.ToLower(html)
	ev := merged.evidence

	trackers := ev.Trackers
	if len(trackers) == 0 {
		trackers = detectTrackers(lower)
	}
	hasTrackers := len(trackers) > 0
	foreignTrackers := detectForeignTrackers(lower)

	hasOffer := hasOfferText(lower) || len(ev.OfferURLs) > 0
	hasTerms := hasTermsText(lower) || len(ev.TermsURLs) > 0
	hasPrivacy := hasPrivacyPolicy(lower) || len(ev.PrivacyURLs) > 0
	hasCookiePol := hasCookiePolicy(lower) || len(ev.CookiePolicyURLs) > 0

	findings := model.LegalScanFindings{
		SSL:               merged.startHTTPS,
		PrivacyPolicy:     hasPrivacy,
		CookieBanner:      hasCookieBanner(lower),
		CookiePolicy:      hasCookiePol,
		FormConsent:       hasFormConsent(lower),
		RequisitesINN:     len(ev.INNs) > 0 || len(ev.OGRNs) > 0,
		Contacts:          len(ev.Emails) > 0 || len(ev.Phones) > 0,
		Offer:             hasOffer,
		Terms:             hasTerms,
		ConsentWithdrawal: hasConsentWithdrawal(lower),
		AdMarking:         len(ev.EridTokens) > 0 || reAdLabel.MatchString(html),
		Trackers:          trackers,
		HasTrackers:       hasTrackers,
		FormsCollectPD:    hasPDForms(lower),
		FormsUnencrypted:  reFormHTTP.MatchString(html),
		ForeignTrackers:   foreignTrackers,
	}

	normalizeFindingsForFeatures(&findings, features)

	checklist := buildLegalScanChecklist(findings, features, ev, finalURL)

	layer1 := model.LegalScanLayer1{
		FinalURL:  finalURL,
		Findings:  findings,
		Checklist: checklist,
	}
	if meta != nil {
		m := *meta
		layer1.Crawl = &m
	}
	return layer1
}

func hasOfferText(lower string) bool {
	patterns := []string{
		"оферт", "/offer", "условия продаж", "условия оказания",
		"правила продаж", "договор-оферт", "публичная оферт", "/oferta", "/dogovor",
	}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func hasTermsText(lower string) bool {
	patterns := []string{
		"пользовательское соглашение", "/terms", "условия использования",
		"user agreement", "/agreement", "правила пользования",
	}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}
