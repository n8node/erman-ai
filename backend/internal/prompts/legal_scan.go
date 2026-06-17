package prompts

import (
	"strings"

	"github.com/erman-ai/erman-ai/internal/i18n"
)

const DefaultLegalScanSystemPrompt = `You are an assistant that turns a technical website compliance scan into a clear client-facing legal risk report under Russian Federation law.

STRICT RULES:
1. Use ONLY risks, articles, and fine amounts from the provided RISK_TABLE block. Never invent or recall articles, amounts, or norms from your own knowledge.
2. If a finding has no RISK_TABLE entry — mark it as requiring manual legal review and do NOT state a fine amount.
3. Do not give legal conclusions or guarantees. Phrase as "potential risk", not "you violated the law".
4. Write in plain language, no legal jargon. One or two sentences per risk.
5. Consider industry: for medicine/finance/online education note possible additional sector requirements (do not invent specifics).
6. Sum potential fixed fines as fine_min..fine_max across matched risks. Turnover fines go in turnover_fine_note separately — do not add to fixed totals.

Write in {{LANGUAGE}}.

Return strictly JSON per OUTPUT_SCHEMA, no markdown.`

func LegalScanSystemPrompt(locale string) string {
	lang := i18n.LanguageName(locale)
	if lang == "" {
		lang = "Russian"
	}
	return strings.ReplaceAll(DefaultLegalScanSystemPrompt, "{{LANGUAGE}}", lang)
}

func LegalScanUserPrompt(industry string, hasAds bool, findingsJSON, riskTableJSON string) string {
	var b strings.Builder
	b.WriteString("INDUSTRY: ")
	b.WriteString(industry)
	b.WriteString("\nHAS_TRAFFIC_FROM_ADS: ")
	if hasAds {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
	b.WriteString("\n\nFINDINGS:\n")
	b.WriteString(findingsJSON)
	b.WriteString("\n\nRISK_TABLE (matched risks only):\n")
	b.WriteString(riskTableJSON)
	b.WriteString(`

OUTPUT_SCHEMA:
{
  "summary": {
    "risks_count": int,
    "fine_min_total": int,
    "fine_max_total": int,
    "turnover_fine_note": string
  },
  "risks": [
    {
      "risk_id": string,
      "title": string,
      "explanation": string,
      "article": string,
      "fine_text": string,
      "severity": "high"|"medium"|"low",
      "how_to_fix": string
    }
  ],
  "industry_note": string,
  "disclaimer": "Отчёт носит информационный характер и не является юридической консультацией. Рекомендуем подтвердить с юристом."
}`)
	return b.String()
}

const DefaultLegalScanEnrichSystemPrompt = `You enrich a crawled website legal scan with structured facts ONLY from the provided page snippets and URL list.

STRICT RULES:
1. Return ONLY URLs that appear in ALLOWED_URLS. Never invent URLs.
2. Return ONLY emails and phones that literally appear in PAGE_SNIPPETS. Never invent contacts.
3. Return ONLY INN/OGRN/OGRNIP numbers that appear in snippets and pass Russian checksum rules (the backend re-validates).
4. Label requisites as INN (10 or 12 digits), OGRN (13 digits), or OGRNIP (15 digits).
5. If uncertain — return empty string or empty array. Do not guess.

Write field values in {{LANGUAGE}} where applicable.

Return strictly JSON per OUTPUT_SCHEMA, no markdown.`

func LegalScanEnrichSystemPrompt(locale string) string {
	lang := i18n.LanguageName(locale)
	if lang == "" {
		lang = "Russian"
	}
	return strings.ReplaceAll(DefaultLegalScanEnrichSystemPrompt, "{{LANGUAGE}}", lang)
}

func LegalScanEnrichUserPrompt(findingsJSON, allowedURLs, snippets string) string {
	var b strings.Builder
	b.WriteString("FINDINGS:\n")
	b.WriteString(findingsJSON)
	b.WriteString("\n\nALLOWED_URLS:\n")
	b.WriteString(allowedURLs)
	b.WriteString("\n\nPAGE_SNIPPETS:\n")
	b.WriteString(snippets)
	b.WriteString(`

OUTPUT_SCHEMA:
{
  "privacy_url": string,
  "cookie_policy_url": string,
  "offer_url": string,
  "terms_url": string,
  "contacts": [{"type": "email"|"phone", "value": string}],
  "requisites": [{"type": "INN"|"OGRN"|"OGRNIP", "value": string}]
}`)
	return b.String()
}
