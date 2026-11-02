package model

type AIAssessmentInput struct {
	Description    string  `json:"description"`
	CompanyContext string  `json:"company_context,omitempty"`
	Industry       string  `json:"industry,omitempty"`
	MonthlyVolume  float64 `json:"monthly_volume,omitempty"`
	MinutesPerUnit float64 `json:"minutes_per_unit,omitempty"`
	HourlyCost     float64 `json:"hourly_cost,omitempty"`
	AutomationPct  float64 `json:"automation_pct,omitempty"`
	Capex          float64 `json:"capex,omitempty"`
	MonthlyCost    float64 `json:"monthly_cost,omitempty"`
	HorizonMonths  int     `json:"horizon_months,omitempty"`
	Locale         string  `json:"locale,omitempty"`
}

type AssessmentStep struct {
	Name           string  `json:"name"`
	Executor       string  `json:"executor,omitempty"`
	MinutesPerUnit float64 `json:"minutes_per_unit"`
	AutomationPct  float64 `json:"automation_pct"`
}

type AssessmentCriterion struct {
	Name      string  `json:"name"`
	Score     int     `json:"score"`
	Weight    float64 `json:"weight"`
	Rationale string  `json:"rationale"`
}

type AssessmentScenario struct {
	Name              string  `json:"name"`
	AutomationPct     float64 `json:"automation_pct"`
	NetBenefitMonthly float64 `json:"net_benefit_monthly"`
	PaybackMonths     float64 `json:"payback_months"`
	ROIPct            float64 `json:"roi_pct"`
}

type AIAssessmentOutput struct {
	ProcessName    string                `json:"process_name"`
	Summary        string                `json:"summary"`
	Steps          []AssessmentStep      `json:"steps"`
	Unknowns       []string              `json:"unknowns"`
	Criteria       []AssessmentCriterion `json:"criteria"`
	ReadinessScore float64               `json:"readiness_score"`
	Autonomy       string                `json:"autonomy"`
	AutonomyText   string                `json:"autonomy_text"`
	RiskLevel      string                `json:"risk_level"`
	Risks          []string              `json:"risks"`
	Financial      FinancialAssessment   `json:"financial"`
	Scenarios      []AssessmentScenario  `json:"scenarios"`
	PilotPlan      []string              `json:"pilot_plan"`
	KPIs           []string              `json:"kpis"`
	Assumptions    []string              `json:"assumptions"`
}

type FinancialAssessment struct {
	HoursSavedMonthly float64 `json:"hours_saved_monthly"`
	MonthlyBenefit    float64 `json:"monthly_benefit"`
	PaybackMonths     float64 `json:"payback_months"`
	ROIPct            float64 `json:"roi_pct"`
	TCO               float64 `json:"tco"`
	NPV               float64 `json:"npv"`
}
