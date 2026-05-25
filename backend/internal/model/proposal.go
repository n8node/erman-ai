package model

type ProposalInput struct {
	ClientCompany       string  `json:"client_company"`
	ClientContact       string  `json:"client_contact"`
	ClientIndustry      string  `json:"client_industry"`
	ClientProblem       string  `json:"client_problem"`
	SolutionName        string  `json:"solution_name"`
	SolutionDescription string  `json:"solution_description"`
	Deliverables        []string `json:"deliverables"`
	ProjectCostRub      float64 `json:"project_cost_rub"`
	TimelineWeeks       int     `json:"timeline_weeks"`
	PaymentSchedule     string  `json:"payment_schedule"`
	SenderCompany       string  `json:"sender_company"`
	SenderContact       string  `json:"sender_contact"`
	SenderPhone         string  `json:"sender_phone"`
	SenderEmail         string  `json:"sender_email"`
	CalculatorRunID     *string                `json:"calculator_run_id,omitempty"`
	CalculatorContext   *CalculatorRunContext  `json:"calculator_context,omitempty"`
	Locale              string                 `json:"locale"`
}

type ProposalTimelinePhase struct {
	Title         string `json:"title"`
	DurationWeeks int    `json:"duration_weeks"`
	Description   string `json:"description"`
}

type ProposalOutput struct {
	Greeting          string                  `json:"greeting"`
	TaskUnderstanding string                  `json:"task_understanding"`
	ProposedSolution  string                  `json:"proposed_solution"`
	ScopeIncluded     []string                `json:"scope_included"`
	ScopeExcluded     []string                `json:"scope_excluded"`
	Timeline          []ProposalTimelinePhase `json:"timeline"`
	CostSummary       string                  `json:"cost_summary"`
	PaymentTerms      string                  `json:"payment_terms"`
	WhyUs             string                  `json:"why_us"`
	NextStep          string                  `json:"next_step"`
}
