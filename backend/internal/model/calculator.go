package model

type CalculatorInput struct {
	ProcessName        string  `json:"process_name"`
	HoursPerMonth      float64 `json:"hours_per_month"`
	HourlyRateRUB      float64 `json:"hourly_rate_rub"`
	EmployeesCount     float64 `json:"employees_count"`
	ErrorRatePct       float64 `json:"error_rate_pct"`
	ErrorCostRUB       float64 `json:"error_cost_rub"`
	AutomationCostRUB  float64 `json:"automation_cost_rub"`
	MonthlySupportRUB  float64 `json:"monthly_support_rub"`
}

type Recommendation string

const (
	RecommendationAutomate       Recommendation = "automate"
	RecommendationConsider       Recommendation = "consider"
	RecommendationNotRecommended Recommendation = "not_recommended"
)

type CalculatorOutput struct {
	CurrentMonthlyCost  float64        `json:"current_monthly_cost"`
	ErrorMonthlyCost    float64        `json:"error_monthly_cost"`
	TotalCurrentCost    float64        `json:"total_current_cost"`
	MonthlySavings      float64        `json:"monthly_savings"`
	NetMonthlySavings   float64        `json:"net_monthly_savings"`
	PaybackMonths       float64        `json:"payback_months"`
	AnnualSavings       float64        `json:"annual_savings"`
	Recommendation      Recommendation `json:"recommendation"`
	RecommendationText  string         `json:"recommendation_text"`
}
