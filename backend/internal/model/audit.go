package model

const (
	MaxAuditProcesses = 7
	MinAuditProcesses = 2
)

type AuditProcessInput struct {
	Name                   string   `json:"name"`
	Department             string   `json:"department"`
	FrequencyRange         string   `json:"frequency_range"`
	MinutesPerCycleRange   string   `json:"minutes_per_cycle_range"`
	FTERange               string   `json:"fte_range"`
	HourlyRateRange        string   `json:"hourly_rate_range"`
	ErrorRateRange         string   `json:"error_rate_range"`
	ErrorCostRange         string   `json:"error_cost_range"`
	Systems                []string `json:"systems"`
	SystemsOther           string   `json:"systems_other,omitempty"`
	Bottleneck             string   `json:"bottleneck"`
	Standardization        string   `json:"standardization"`
	IntegrationComplexity  int      `json:"integration_complexity"`
	AutomationReadiness    string   `json:"automation_readiness"`
	ExpectedImpact         string   `json:"expected_impact"`
}

type AuditProcessScore struct {
	Name              string  `json:"name"`
	MonthlyHours      float64 `json:"monthly_hours"`
	MonthlyLaborCost  float64 `json:"monthly_labor_cost_rub"`
	MonthlyErrorCost  float64 `json:"monthly_error_cost_rub"`
	TotalMonthlyCost  float64 `json:"total_monthly_cost_rub"`
	AutomationScore   float64 `json:"automation_score"`
	PriorityRank      int     `json:"priority_rank"`
	QuickWin          bool    `json:"quick_win"`
	Rationale         string  `json:"rationale"`
}

type AuditInput struct {
	CompanyName          string              `json:"company_name"`
	PrimaryGoal          string              `json:"primary_goal"`
	PriorityCriteria     []string            `json:"priority_criteria"`
	CompanySize          string              `json:"company_size"`
	OperationalHeadcount string              `json:"operational_headcount"`
	ITSystems            []string            `json:"it_systems"`
	ITSystemsOther       string              `json:"it_systems_other,omitempty"`
	DataDuplication      string              `json:"data_duplication"`
	BudgetRange          string              `json:"budget_range"`
	DecisionTimeline     string              `json:"decision_timeline"`
	Processes            []AuditProcessInput `json:"processes"`
	Locale               string              `json:"locale"`
}

type AuditPriorityRow struct {
	Rank              int     `json:"rank"`
	ProcessName       string  `json:"process_name"`
	AutomationScore   float64 `json:"automation_score"`
	MonthlySavingsEst float64 `json:"monthly_savings_est_rub"`
	PaybackMonthsEst  float64 `json:"payback_months_est"`
	QuickWin          bool    `json:"quick_win"`
	Rationale         string  `json:"rationale"`
}

type AuditRoadmapPhase struct {
	Phase       string   `json:"phase"`
	Period      string   `json:"period"`
	Processes   []string `json:"processes"`
	Deliverables []string `json:"deliverables"`
}

type AuditRisk struct {
	Risk       string `json:"risk"`
	Mitigation string `json:"mitigation"`
	Severity   string `json:"severity"`
}

type AuditOutput struct {
	ExecutiveSummary   string              `json:"executive_summary"`
	CompanyContext     string              `json:"company_context"`
	ProcessScores      []AuditProcessScore `json:"process_scores"`
	PriorityRanking    []AuditPriorityRow  `json:"priority_ranking"`
	TotalMonthlyCost   float64             `json:"total_monthly_cost_rub"`
	TotalMonthlySavings float64            `json:"total_monthly_savings_est_rub"`
	QuickWins          []string            `json:"quick_wins"`
	Roadmap            []AuditRoadmapPhase `json:"roadmap"`
	Risks              []AuditRisk         `json:"risks"`
	NextSteps          []string            `json:"next_steps"`
	MetricsToTrack     []string            `json:"metrics_to_track"`
}
