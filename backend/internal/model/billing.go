package model

type Plan struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Slug            string         `json:"slug"`
	PriceMonthlyRUB int            `json:"price_monthly_rub"`
	PriceYearlyRUB  int            `json:"price_yearly_rub"`
	ToolLimits      map[string]int  `json:"tool_limits"`
	Features        map[string]any  `json:"features"`
	SupportLevel    string         `json:"support_level"`
	IsPublic        bool           `json:"is_public"`
	IsArchived      bool           `json:"is_archived"`
}
