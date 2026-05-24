package model

const MaxStrategyCalculatorRuns = 10

type CalculatorRunContext struct {
	RunID              string  `json:"run_id"`
	ProcessName        string  `json:"process_name"`
	NetBenefitMonthly  float64 `json:"net_benefit_monthly_rub"`
	PaybackMonths      float64 `json:"payback_months"`
	ROIHorizonPct      float64 `json:"roi_horizon_pct"`
	NPV                float64 `json:"npv_rub"`
	Capex              float64 `json:"capex_rub"`
	MonthlySupport     float64 `json:"monthly_support_rub"`
	Recommendation     string  `json:"recommendation"`
	RecommendationText string  `json:"recommendation_text"`
}

type StrategyInput struct {
	CompanyName         string   `json:"company_name"`
	BusinessDescription string   `json:"business_description"`
	Industry            string   `json:"industry"`
	CompanySize         string   `json:"company_size"`
	AnnualRevenueRange  string   `json:"annual_revenue_range"`
	MarketPosition      string   `json:"market_position"`
	CurrentAILevel      string   `json:"current_ai_level"`
	MainGoals           []string `json:"main_goals"`
	KeyProcesses        []string `json:"key_processes"`
	PainPoints          string   `json:"pain_points"`
	DataMaturity        string   `json:"data_maturity"`
	ChangeReadiness     string   `json:"change_readiness"`
	BudgetRange         string   `json:"budget_range"`
	Timeline            string   `json:"timeline"`
	ExistingTools       string   `json:"existing_tools"`
	CalculatorRunIDs    []string `json:"calculator_run_ids,omitempty"`
	// Deprecated: use calculator_run_ids.
	CalculatorRunID     *string                `json:"calculator_run_id,omitempty"`
	CalculatorContexts  []CalculatorRunContext `json:"calculator_contexts,omitempty"`
	Locale              string                 `json:"locale"`
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

type StrategyProcessAnalysis struct {
	Name         string `json:"name"`
	CurrentState string `json:"current_state"`
	PainPoints   string `json:"pain_points"`
	AiPotential  string `json:"ai_potential"`
}

type StrategyUseCase struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
	Impact      string `json:"impact"`
}

type StrategyImplementationPhase struct {
	Title        string   `json:"title"`
	Duration     string   `json:"duration"`
	Deliverables []string `json:"deliverables"`
}

type StrategyBudgetLine struct {
	Category    string `json:"category"`
	AmountRange string `json:"amount_range"`
	Notes       string `json:"notes"`
}

type StrategyBudgetOverview struct {
	Summary string               `json:"summary"`
	Lines   []StrategyBudgetLine `json:"lines"`
}

type StrategyOutput struct {
	ExecutiveSummary       string                        `json:"executive_summary"`
	CurrentSituation       string                        `json:"current_situation"`
	GoalsAndRationale      string                        `json:"goals_and_rationale"`
	ProcessAnalysis        []StrategyProcessAnalysis     `json:"process_analysis"`
	DataAndInfrastructure  string                        `json:"data_and_infrastructure"`
	AiUseCases             []StrategyUseCase             `json:"ai_use_cases"`
	RecommendedSolutions   []StrategySolution            `json:"recommended_solutions"`
	ImplementationPlan     []StrategyImplementationPhase `json:"implementation_plan"`
	TeamAndTraining        string                        `json:"team_and_training"`
	ArchitectureOverview   string                        `json:"architecture_overview"`
	DataGovernance           string                        `json:"data_governance"`
	EthicsAndCompliance    string                        `json:"ethics_and_compliance"`
	BudgetOverview         StrategyBudgetOverview        `json:"budget_overview"`
	SuccessMetrics         []StrategyMetric              `json:"success_metrics"`
	StrategyAdjustmentPlan string                        `json:"strategy_adjustment_plan"`
	Risks                  []StrategyRisk                `json:"risks"`
	Roadmap                []StrategyRoadmapPhase          `json:"roadmap"`
	Next30Days             []string                      `json:"next_30_days"`
}
