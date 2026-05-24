package service

import (
	"testing"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCalculateExcel_demoValues(t *testing.T) {
	in := model.CalculatorInput{
		ProcessName:         "Invoice processing",
		HmMode:              "direct",
		HoursSavedMonth:     800,
		HourlyCostLoaded:    2500,
		Utilization:         0.8,
		OtherBenefitMonthly: 15000,
		MonthlySolutionCost: 10000,
		Capex:               300000,
		HorizonMonths:       12,
		DiscountRateAnnual:  0.2,
		ProcessSteps: []model.ProcessStep{
			{Operation: "Enter invoice", MinutesPerUnit: 8, RateRubPerMin: 17},
			{Operation: "Approve", MinutesPerUnit: 2, RateRubPerMin: 33},
		},
		UnitsPerMonth:            8000,
		AutomationPct:            80,
		ErrorRateBeforePct:       5,
		ThroughputBeforePerDay:   20,
	}

	out := CalculateExcel(in, "ru")

	assert.InDelta(t, 800, out.HoursSavedMonth, 1)
	assert.InDelta(t, 2005000, out.NetBenefitMonthly, 1000)
	assert.InDelta(t, 5.952, out.FTE, 0.05)
	assert.InDelta(t, 0.15, out.PaybackMonths, 0.05)
	assert.True(t, out.ROIHorizonPct > 7000)
	assert.Equal(t, model.RecommendationAutomate, out.Recommendation)
	assert.Len(t, out.KPIRows, 6)
}

func TestResolveHoursSaved_fromProcess(t *testing.T) {
	in := model.CalculatorInput{
		HmMode: "process",
		ProcessSteps: []model.ProcessStep{
			{MinutesPerUnit: 10},
		},
		UnitsPerMonth:  8000,
		AutomationPct:  80,
	}
	hm := ResolveHoursSaved(in)
	// 10 min * 8000 * 0.8 / 60 = 1066.67
	assert.InDelta(t, 1066.67, hm, 1)
}
