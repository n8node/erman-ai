package prompts

import (
	"github.com/erman-ai/erman-ai/internal/model"
)

// ProposalScenarioRules returns runtime instructions appended to the system prompt.
func ProposalScenarioRules(scenario string, includePricing bool) string {
	scenario = model.NormalizeProposalScenario(scenario)

	var rules string
	switch scenario {
	case model.ProposalScenarioColdOutreach:
		rules = `SCENARIO: cold_outreach (first contact — NO prior conversation with the client).

Tone and content rules:
- greeting: First-touch letter to the company. Do NOT thank the client for an inquiry, request, or meeting unless prior_contact_summary explicitly says so.
- Do NOT write "as we discussed", "following our call", "per your request", or similar.
- If client_contact is empty, address the company (e.g. "Dear [client_company] team"). If provided, use the name formally without implying prior dialogue.
- task_understanding: Frame as your professional assessment / industry observation using problem_source and client_problem — hypotheses, not "your brief".
- next_step: Propose a short discovery call or meeting (15–30 min). Do NOT ask to sign a contract or confirm the project yet.
- why_us: Focus on credibility for a first introduction, not on "why we won your tender".`

	case model.ProposalScenarioProactiveOffer:
		rules = `SCENARIO: proactive_offer (integrator proposes automation after analysis — client did NOT ask for a quote).

Tone and content rules:
- greeting: Proactive value proposition — "we analyzed / we propose to consider automation of…". No "thank you for contacting us".
- task_understanding: Lead with process pain and ROI from calculator_context when present. Use numbers (payback, monthly benefit, NPV). Frame as opportunity, not response to RFP.
- Reference calculator_context explicitly when available.
- next_step: Suggest diagnostic workshop, pilot scope discussion, or ROI review call — not immediate contract signing.
- Do NOT imply the client requested this proposal.`

	default: // after_contact
		rules = `SCENARIO: after_contact (follow-up to an existing conversation or client request).

Tone and content rules:
- greeting: Reference prior_contact_summary naturally ("Following our discussion…", "As agreed…").
- task_understanding: Show you understood the client's stated needs from prior_contact_summary and client_problem.
- next_step: Concrete next step toward project start (approve scope, sign agreement, kick-off date) as appropriate.
- Use client_contact by name when provided.`
	}

	if !includePricing {
		rules += `

PRICING MODE: soft (include_pricing = false):
- cost_summary: Do NOT state exact project_cost_rub. Use ranges, "after diagnostic", or qualitative investment framing.
- payment_terms: Keep high-level or defer to post-discovery discussion.
- next_step: Emphasize meeting before commercial commitment.`
	} else {
		rules += `

PRICING MODE: full (include_pricing = true):
- cost_summary: Include project_cost_rub clearly with breakdown aligned to input.
- payment_terms: Expand payment_schedule from input professionally.`
	}

	return rules
}
