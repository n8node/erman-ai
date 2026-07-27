package prompts

// Consulting multi-pass section prompts (appended to base system prompt per pass).

const StrategyPassContext = `
PASS 1 — CONTEXT & EXECUTIVE (return ONLY this JSON object, no other keys):
{
  "executive_summary": "string — minimum 1800 characters, 4-5 paragraphs with numbers",
  "current_situation": "string — minimum 1500 characters, 3-4 paragraphs",
  "goals_and_rationale": "string — minimum 1000 characters, 2-3 paragraphs"
}`

const StrategyPassAnalysis = `
PASS 2 — PROCESSES & PRIORITIES (return ONLY this JSON object):
{
  "process_analysis": [{"name":"","current_state":"","pain_points":"","ai_potential":""}],
  "data_and_infrastructure": "string — minimum 1200 characters",
  "ai_use_cases": [{"title":"","description":"","priority":1,"impact":"high|medium|low"}],
  "maturity_matrix": [{"criterion":"","current_level":"","target_level":"","gap":""}],
  "priority_matrix": [{"use_case":"","impact":1,"effort":1,"score":0,"priority":1}]
}
Requirements: process_analysis 4-5 items; ai_use_cases exactly 6; maturity_matrix at least 6 rows;
priority_matrix 6 rows with impact and effort 1-5, score = impact/effort rounded to 1 decimal.`

const StrategyPassPlan = `
PASS 3 — SOLUTIONS, ROADMAP & BUDGET (return ONLY this JSON object):
{
  "recommended_solutions": [{"title":"","description":"","rationale":"","priority":1}],
  "implementation_plan": [{"title":"","duration":"","deliverables":[""]}],
  "roadmap": [{"phase":"","quarter":"","initiatives":[""]}],
  "budget_overview": {"summary":"","lines":[{"category":"","amount_range":"","notes":""}]},
  "budget_phases": [{"phase":"","capex_rub":"","opex_monthly_rub":"","cumulative_rub":""}]
}
Requirements: 3 recommended_solutions; implementation_plan 5-6 phases; roadmap 4 quarters;
budget_overview.lines at least 5; budget_phases 4-6 rows aligned with implementation_plan.`

const StrategyPassGovernance = `
PASS 4 — GOVERNANCE, METRICS & DIAGRAMS (return ONLY this JSON object):
{
  "team_and_training": "string — minimum 900 characters",
  "architecture_overview": "string — minimum 900 characters",
  "data_governance": "string — minimum 700 characters",
  "ethics_and_compliance": "string — minimum 700 characters",
  "success_metrics": [{"metric":"","target":"","timeframe":""}],
  "strategy_adjustment_plan": "string — minimum 600 characters",
  "risks": [{"risk":"","mitigation":"","severity":"low|medium|high"}],
  "next_30_days": [""],
  "tech_stack": [{"layer":"","tool":"","role":"","status":"existing|planned|evaluate"}],
  "stakeholder_plan": [{"role":"","responsibility":"","involvement":"high|medium|low"}],
  "diagrams": [{"type":"architecture|roadmap|process","title":"","mermaid":""}]
}
Requirements: success_metrics at least 8; risks at least 6; next_30_days at least 8 steps;
tech_stack 6 rows; stakeholder_plan 5 rows; diagrams 2-3 valid mermaid diagrams.` + MermaidDiagramRules

const StrategyExpandPrompt = `
EXPANSION PASS — the strategy draft below is too thin. Expand ONLY the listed fields in your JSON response.
Keep all other content implied by context. Return ONLY the fields to expand with substantially longer text (2x minimum length).
Fields to expand: %s
Draft JSON:
%s`

const StrategyExtendedFieldsPrompt = `
ADDITIONAL STRUCTURED FIELDS — include in your JSON response alongside existing schema:
  "maturity_matrix": [{"criterion":"","current_level":"","target_level":"","gap":""}] — at least 5 rows
  "priority_matrix": [{"use_case":"","impact":1,"effort":1,"score":0,"priority":1}] — 5 rows
  "tech_stack": [{"layer":"","tool":"","role":"","status":"existing|planned|evaluate"}] — 5 rows
  "stakeholder_plan": [{"role":"","responsibility":"","involvement":"high|medium|low"}] — 4 rows
  "budget_phases": [{"phase":"","capex_rub":"","opex_monthly_rub":"","cumulative_rub":""}] — 4 rows
  "diagrams": [{"type":"architecture|roadmap|process","title":"","mermaid":""}] — 2 mermaid diagrams` + MermaidDiagramRules
