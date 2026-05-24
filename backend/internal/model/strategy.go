package model

type StrategyInput struct {
	CompanyName        string   `json:"company_name"`
	Industry           string   `json:"industry"`
	CompanySize        string   `json:"company_size"`
	AnnualRevenueRange string   `json:"annual_revenue_range"`
	CurrentAILevel     string   `json:"current_ai_level"`
	MainGoals          []string `json:"main_goals"`
	PainPoints         string   `json:"pain_points"`
	BudgetRange        string   `json:"budget_range"`
	Timeline           string   `json:"timeline"`
	ExistingTools      string   `json:"existing_tools"`
	CalculatorRunID    *string  `json:"calculator_run_id,omitempty"`
	Locale             string   `json:"locale"`
}

type StrategySolution struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Rationale   string `json:"rationale"`
	Priority    int    `json:"priority"`
}

type StrategyRoadmapPhase struct {
	Phase       string   `json:"phase"`
	Quarter     string   `json:"quarter"`
	Initiatives []string `json:"initiatives"`
}

type StrategyRisk struct {
	Risk       string `json:"risk"`
	Mitigation string `json:"mitigation"`
	Severity   string `json:"severity"`
}

type StrategyMetric struct {
	Metric    string `json:"metric"`
	Target    string `json:"target"`
	Timeframe string `json:"timeframe"`
}

type StrategyOutput struct {
	ExecutiveSummary     string                 `json:"executive_summary"`
	CurrentSituation     string                 `json:"current_situation"`
	RecommendedSolutions []StrategySolution     `json:"recommended_solutions"`
	Roadmap              []StrategyRoadmapPhase `json:"roadmap"`
	Risks                []StrategyRisk         `json:"risks"`
	SuccessMetrics       []StrategyMetric       `json:"success_metrics"`
	Next30Days           []string               `json:"next_30_days"`
}
