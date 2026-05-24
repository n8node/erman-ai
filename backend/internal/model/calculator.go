package model

type ProcessStep struct {
	Operation      string  `json:"operation"`
	Executor       string  `json:"executor,omitempty"`
	MinutesPerUnit float64 `json:"minutes_per_unit"`
	RateRubPerMin  float64 `json:"rate_rub_per_min"`
}

type CalculatorInput struct {
	ProcessName string `json:"process_name"`

	// Step 1: process assessment
	HmMode          string        `json:"hm_mode"` // "process" | "direct"
	ProcessSteps    []ProcessStep `json:"process_steps,omitempty"`
	UnitsPerMonth   float64       `json:"units_per_month"`
	AutomationPct   float64       `json:"automation_pct"`
	HoursSavedMonth float64       `json:"hours_saved_month"` // Hm — direct or overridden

	// Optional KPI helpers
	ErrorRateBeforePct    float64 `json:"error_rate_before_pct"`
	ThroughputBeforePerDay float64 `json:"throughput_before_per_day"`

	// Step 2: financial model (Excel)
	HourlyCostLoaded     float64 `json:"hourly_cost_loaded"`      // Ch
	Utilization          float64 `json:"utilization"`             // L
	OtherBenefitMonthly  float64 `json:"other_benefit_monthly"`   // Sm
	MonthlySolutionCost  float64 `json:"monthly_solution_cost"`   // Om
	Capex                float64 `json:"capex"`                   // I0
	HorizonMonths        int     `json:"horizon_months"`          // n
	DiscountRateAnnual   float64 `json:"discount_rate_annual"`    // r
}

type Recommendation string

const (
	RecommendationAutomate       Recommendation = "automate"
	RecommendationConsider       Recommendation = "consider"
	RecommendationNotRecommended Recommendation = "not_recommended"
)

type KPIRow struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Before string `json:"before"`
	After  string `json:"after"`
	Change string `json:"change"`
}

type CalculatorOutput struct {
	HoursSavedMonth   float64 `json:"hours_saved_month"`
	MinutesPerUnit    float64 `json:"minutes_per_unit"`
	NetBenefitMonthly float64 `json:"net_benefit_monthly"` // B_m
	FTE               float64 `json:"fte"`
	PaybackMonths     float64 `json:"payback_months"`
	ROIHorizonPct     float64 `json:"roi_horizon_pct"`
	TCOHorizon        float64 `json:"tco_horizon"`
	NPV               float64 `json:"npv"`
	PayrollSavings    float64 `json:"payroll_savings"`
	MonthlyHoursBefore float64 `json:"monthly_hours_before"`
	MonthlyHoursAfter  float64 `json:"monthly_hours_after"`

	Recommendation     Recommendation `json:"recommendation"`
	RecommendationText string         `json:"recommendation_text"`
	KPIRows            []KPIRow         `json:"kpi_rows"`
}
