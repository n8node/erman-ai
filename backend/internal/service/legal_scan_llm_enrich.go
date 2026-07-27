package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/i18n"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
)

type legalScanLLMEnrichment struct {
	PrivacyURL      string                `json:"privacy_url"`
	CookiePolicyURL string                `json:"cookie_policy_url"`
	OfferURL        string                `json:"offer_url"`
	TermsURL        string                `json:"terms_url"`
	Contacts        []legalScanLLMContact `json:"contacts"`
	Requisites      []legalScanLLMReq     `json:"requisites"`
}

type legalScanLLMContact struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type legalScanLLMReq struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func mergeEvidenceFromPages(pages []fetchedPage) scanEvidence {
	var ev scanEvidence
	for _, p := range pages {
		ev = mergeScanEvidence(ev, extractScanEvidence(p.HTML, p.URL))
	}
	return ev
}

func (s *LegalScanService) enrichLayer1WithLLM(
	ctx context.Context,
	run *model.ToolRun,
	pages []fetchedPage,
	layer1 model.LegalScanLayer1,
	input model.LegalScanInput,
) model.LegalScanLayer1 {
	llmStored, err := s.llmCfg.GetStored(ctx)
	if err != nil {
		return layer1
	}
	strategyStored, err := s.strategy.GetStored(ctx)
	if err != nil {
		return layer1
	}

	settings := llmStored.Config.LegalScanLLMSettings
	provider := settings.Provider
	creds := s.llm.CredentialsFromStored(strategyStored.Config)
	apiKey := s.llm.ResolveKey(provider, creds)
	if apiKey == "" {
		return layer1
	}

	fetchedURLs := []string{}
	if layer1.Crawl != nil {
		fetchedURLs = layer1.Crawl.FetchedURLs
	} else {
		for _, p := range pages {
			fetchedURLs = append(fetchedURLs, p.URL)
		}
	}

	snippets := buildLegalScanPageSnippets(pages, 3500)
	findingsJSON, _ := json.Marshal(layer1.Findings)
	locale := i18n.NormalizeLocale(input.Locale)

	req := LLMCompletionRequest{
		Provider:     provider,
		Model:        settings.ActiveModel(),
		SystemPrompt: prompts.LegalScanEnrichSystemPrompt(locale),
		UserPrompt: prompts.LegalScanEnrichUserPrompt(
			string(findingsJSON),
			strings.Join(fetchedURLs, "\n"),
			snippets,
		),
		Temperature: 0.1,
		MaxTokens:   2048,
		APIKey:      apiKey,
		BaseURL:     s.llm.BaseURL(provider),
		FolderID:    creds.YandexFolderID,
	}

	result, err := s.llm.Complete(ctx, req)
	if err != nil {
		s.logger.Warn("legal scan llm enrich failed", "run_id", run.ID, "error", err)
		return layer1
	}

	costUSD, costRUB := s.strategy.UsageCosts(strategyStored.Config, provider, result.Model, result.PromptTokens, result.CompletionTokens)
	_ = s.usageLog.Create(ctx, run.UserID, run.ID, string(provider), result.Model, result.PromptTokens, result.CompletionTokens, costUSD, costRUB)

	raw, err := parseLegalScanEnrichment(result.Content)
	if err != nil {
		s.logger.Warn("legal scan llm enrich parse failed", "run_id", run.ID, "error", err)
		return layer1
	}

	ev := mergeEvidenceFromPages(pages)
	ev = applyValidatedEnrichmentToEvidence(ev, validateLegalScanEnrichment(raw, pages, fetchedURLs))
	ev = filterEvidenceDocumentURLs(ev, pages)

	startURL := input.URL
	if layer1.Crawl != nil && layer1.Crawl.StartURL != "" {
		startURL = layer1.Crawl.StartURL
	}
	httpsOK := layer1.Findings.SSL
	merged := mergeLegalScanPages(pages, startURL, httpsOK)
	merged.evidence = ev
	return buildLegalScanLayer1(merged, input.SiteFeatures, layer1.Crawl, layer1.FinalURL)
}

func applyValidatedEnrichmentToEvidence(ev scanEvidence, en legalScanLLMEnrichment) scanEvidence {
	if en.PrivacyURL != "" {
		ev.PrivacyURLs = appendUnique(ev.PrivacyURLs, en.PrivacyURL)
	}
	if en.CookiePolicyURL != "" {
		ev.CookiePolicyURLs = appendUnique(ev.CookiePolicyURLs, en.CookiePolicyURL)
	}
	if en.OfferURL != "" {
		ev.OfferURLs = appendUnique(ev.OfferURLs, en.OfferURL)
	}
	if en.TermsURL != "" {
		ev.TermsURLs = appendUnique(ev.TermsURLs, en.TermsURL)
	}
	for _, c := range en.Contacts {
		if c.Type == "email" {
			ev.Emails = appendUnique(ev.Emails, c.Value)
		} else if c.Type == "phone" {
			ev.Phones = appendUnique(ev.Phones, c.Value)
		}
	}
	for _, r := range en.Requisites {
		switch r.Type {
		case "INN":
			ev.Requisites = appendUnique(ev.Requisites, "ИНН: "+r.Value)
		case "OGRN":
			ev.Requisites = appendUnique(ev.Requisites, "ОГРН: "+r.Value)
		case "OGRNIP":
			ev.Requisites = appendUnique(ev.Requisites, "ОГРНИП: "+r.Value)
		}
	}
	return ev
}

func buildLegalScanPageSnippets(pages []fetchedPage, maxChars int) string {
	var b strings.Builder
	for _, p := range pages {
		text := stripHTMLForTextExtraction(p.HTML)
		text = strings.Join(strings.Fields(text), " ")
		if len(text) > 1200 {
			text = text[:1200] + "…"
		}
		b.WriteString("URL: ")
		b.WriteString(p.URL)
		b.WriteString("\n")
		b.WriteString(text)
		b.WriteString("\n\n")
		if b.Len() >= maxChars {
			break
		}
	}
	out := b.String()
	if len(out) > maxChars {
		out = out[:maxChars]
	}
	return out
}

func parseLegalScanEnrichment(content string) (*legalScanLLMEnrichment, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	var out legalScanLLMEnrichment
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	return &out, nil
}

func validateLegalScanEnrichment(raw *legalScanLLMEnrichment, pages []fetchedPage, allowedURLs []string) legalScanLLMEnrichment {
	allowed := make(map[string]struct{}, len(allowedURLs))
	for _, u := range allowedURLs {
		allowed[normalizeLegalScanURL(u)] = struct{}{}
	}
	pageByURL := make(map[string]string, len(pages))
	for _, p := range pages {
		pageByURL[normalizeLegalScanURL(p.URL)] = p.HTML
	}
	isAllowedURL := func(u string) bool {
		if u == "" {
			return false
		}
		_, ok := allowed[normalizeLegalScanURL(u)]
		return ok
	}
	confirmsURL := func(u string, confirms func(string) bool) bool {
		html, ok := pageByURL[normalizeLegalScanURL(u)]
		return ok && confirms(html)
	}

	out := legalScanLLMEnrichment{
		PrivacyURL:      raw.PrivacyURL,
		CookiePolicyURL: raw.CookiePolicyURL,
		OfferURL:        raw.OfferURL,
		TermsURL:        raw.TermsURL,
	}
	if !isAllowedURL(out.PrivacyURL) || !confirmsURL(out.PrivacyURL, pageConfirmsPrivacyDocument) {
		out.PrivacyURL = ""
	}
	if !isAllowedURL(out.CookiePolicyURL) || !confirmsURL(out.CookiePolicyURL, pageConfirmsCookiePolicyDocument) {
		out.CookiePolicyURL = ""
	}
	if !isAllowedURL(out.OfferURL) || !confirmsURL(out.OfferURL, pageConfirmsOfferDocument) {
		out.OfferURL = ""
	}
	if !isAllowedURL(out.TermsURL) || !confirmsURL(out.TermsURL, pageConfirmsTermsDocument) {
		out.TermsURL = ""
	}

	for _, c := range raw.Contacts {
		switch strings.ToLower(c.Type) {
		case "email":
			if isValidContactEmail(c.Value) {
				out.Contacts = append(out.Contacts, legalScanLLMContact{Type: "email", Value: strings.ToLower(c.Value)})
			}
		case "phone":
			if norm := normalizePhone(c.Value); norm != "" {
				out.Contacts = append(out.Contacts, legalScanLLMContact{Type: "phone", Value: norm})
			}
		}
	}

	for _, r := range raw.Requisites {
		t := strings.ToUpper(strings.TrimSpace(r.Type))
		v := strings.TrimSpace(r.Value)
		switch t {
		case "INN", "ИНН":
			if (len(v) == 10 && validateINN10(v)) || (len(v) == 12 && validateINN12(v)) {
				out.Requisites = append(out.Requisites, legalScanLLMReq{Type: "INN", Value: v})
			}
		case "OGRN", "ОГРН":
			if len(v) == 13 && validateOGRN13(v) {
				out.Requisites = append(out.Requisites, legalScanLLMReq{Type: "OGRN", Value: v})
			}
		case "OGRNIP", "ОГРНИП":
			if len(v) == 15 && validateOGRN15(v) {
				out.Requisites = append(out.Requisites, legalScanLLMReq{Type: "OGRNIP", Value: v})
			}
		}
	}
	return out
}
