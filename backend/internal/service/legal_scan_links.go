package service

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	reLegalScanHref     = regexp.MustCompile(`(?i)href\s*=\s*["']([^"'#]+)`)
	reLegalScanHrefText = regexp.MustCompile(`(?i)<a[^>]+href\s*=\s*["']([^"']+)["'][^>]*>([^<]{0,200})`)
)

var legalLinkKeywords = []string{
	"политик", "конфиденциальн", "персональн", "privacy", "policy",
	"оферт", "offer", "условия продаж", "условия оказания", "договор", "oferta", "dogovor",
	"cookie", "куки",
	"соглашение", "terms", "agreement", "legal",
	"контакт", "contact", "реквизит", "about", "о компании", "о нас",
}

type scoredLink struct {
	URL      string
	Depth    int
	Priority int
}

func normalizeLegalScanURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	u.Fragment = ""
	u.RawQuery = stripTrackingQuery(u.Query())
	return strings.TrimSuffix(u.String(), "/")
}

func stripTrackingQuery(q url.Values) string {
	if len(q) == 0 {
		return ""
	}
	for k := range q {
		kl := strings.ToLower(k)
		if strings.HasPrefix(kl, "utm_") || kl == "fbclid" || kl == "gclid" {
			delete(q, k)
		}
	}
	return q.Encode()
}

func sameLegalScanHost(a, b string) bool {
	pa, err1 := url.Parse(a)
	pb, err2 := url.Parse(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return strings.EqualFold(pa.Hostname(), pb.Hostname())
}

func scoreLegalLink(href, anchor, pageURL string) int {
	combined := strings.ToLower(href + " " + anchor)
	score := 0
	for _, kw := range legalLinkKeywords {
		if strings.Contains(combined, kw) {
			score += 10
		}
	}
	path := strings.ToLower(href)
	legalPaths := []string{"/privacy", "/policy", "/offer", "/terms", "/legal", "/contact", "/about", "/cookie", "/oferta", "/dogovor", "/rekvizit"}
	for _, p := range legalPaths {
		if strings.Contains(path, p) {
			score += 15
		}
	}
	if strings.Contains(strings.ToLower(pageURL), "footer") || strings.Contains(strings.ToLower(anchor), "footer") {
		score += 3
	}
	return score
}

func extractLegalScanLinks(html, pageURL string, depth int, baseHost string, sameHostOnly bool) []scoredLink {
	seen := make(map[string]struct{})
	var out []scoredLink

	add := func(href, anchor string) {
		resolved := resolveScanURL(href, pageURL)
		if resolved == "" {
			return
		}
		parsed, err := url.Parse(resolved)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return
		}
		if sameHostOnly && !strings.EqualFold(parsed.Hostname(), baseHost) {
			return
		}
		if err := validateLegalScanHost(parsed.Hostname()); err != nil {
			return
		}
		norm := normalizeLegalScanURL(resolved)
		if norm == "" {
			return
		}
		if _, ok := seen[norm]; ok {
			return
		}
		seen[norm] = struct{}{}
		out = append(out, scoredLink{
			URL:      norm,
			Depth:    depth,
			Priority: scoreLegalLink(parsed.Path, anchor, pageURL),
		})
	}

	for _, m := range reLegalScanHrefText.FindAllStringSubmatch(html, -1) {
		add(m[1], strings.TrimSpace(m[2]))
	}
	for _, m := range reLegalScanHref.FindAllStringSubmatch(html, -1) {
		add(m[1], "")
	}

	return out
}

func sortScoredLinks(links []scoredLink) {
	// Higher priority first; stable enough via insertion sort for small slices.
	for i := 1; i < len(links); i++ {
		j := i
		for j > 0 && links[j].Priority > links[j-1].Priority {
			links[j], links[j-1] = links[j-1], links[j]
			j--
		}
	}
}
