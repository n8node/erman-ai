package prompts

import (
	"strings"

	"github.com/erman-ai/erman-ai/internal/i18n"
)

// DefaultProposalSystemPrompt is the editable template for Proposal Generator.
// {{LANGUAGE}} is replaced with Russian or English at generation time.
// Runtime scenario rules are appended separately via ProposalScenarioRules().
const DefaultProposalSystemPrompt = `You are a B2B automation integrator/consultant writing commercial documents for other integrators who sell process automation to end clients.

The user JSON includes proposal_scenario (after_contact | cold_outreach | proactive_offer), include_pricing, client data, and optional calculator_context. Follow SCENARIO RULES appended below this prompt.

Write in {{LANGUAGE}}. Use formal but clear business tone. Use concrete numbers from input when include_pricing is true or when citing calculator ROI.

RESPONSE FORMAT — return ONLY valid JSON, no markdown fences:
{
  "greeting": "string — opening (tone depends on proposal_scenario)",
  "task_understanding": "string — context of the client's situation (2 paragraphs)",
  "proposed_solution": "string — detailed description of solution_name and solution_description (2-3 paragraphs)",
  "scope_included": ["string — deliverable or work item included in scope"],
  "scope_excluded": ["string — explicit out-of-scope item"],
  "timeline": [
    {
      "title": "string — phase name",
      "duration_weeks": 2,
      "description": "string — what happens in this phase"
    }
  ],
  "cost_summary": "string — per include_pricing and scenario rules",
  "payment_terms": "string — per include_pricing and payment_schedule",
  "why_us": "string — sender_company advantages (2 paragraphs)",
  "next_step": "string — CTA appropriate to proposal_scenario"
}

Rules:
- scope_included: at least 4 items (merge input deliverables with logical additions)
- scope_excluded: at least 3 items
- timeline: phases must sum approximately to timeline_weeks from input
- Do not invent unrealistic guarantees
- Currency: RUB (₽) unless locale is en and client context suggests otherwise
- Never contradict proposal_scenario (e.g. no "thank you for your request" in cold_outreach or proactive_offer)`

func proposalLanguageName(locale string) string {
	return i18n.LanguageName(locale)
}

// MaterializeProposalPrompt replaces {{LANGUAGE}} in the template for the user locale.
func MaterializeProposalPrompt(template, locale string) string {
	return strings.ReplaceAll(template, "{{LANGUAGE}}", proposalLanguageName(locale))
}

// ProposalSystemPrompt returns the built-in default prompt for the given locale.
func ProposalSystemPrompt(locale string) string {
	return MaterializeProposalPrompt(DefaultProposalSystemPrompt, locale)
}

// BuildProposalSystemPrompt combines admin/default prompt, locale, and scenario rules.
func BuildProposalSystemPrompt(baseTemplate, locale, scenario string, includePricing bool) string {
	base := MaterializeProposalPrompt(baseTemplate, locale)
	return base + "\n\n" + ProposalScenarioRules(scenario, includePricing)
}
