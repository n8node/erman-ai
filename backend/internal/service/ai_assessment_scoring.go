package service

import (
	"github.com/erman-ai/erman-ai/internal/model"
	"math"
)

func CalculateAssessmentFinancial(in model.AIAssessmentInput, automation float64) model.FinancialAssessment {
	volume, minutes, hourly := in.MonthlyVolume, in.MinutesPerUnit, in.HourlyCost
	if volume < 0 {
		volume = 0
	}
	if minutes < 0 {
		minutes = 0
	}
	if hourly < 0 {
		hourly = 0
	}
	hours := volume * minutes / 60 * automation / 100
	benefit := hours*hourly - in.MonthlyCost
	n := in.HorizonMonths
	if n <= 0 {
		n = 12
	}
	capex := math.Max(in.Capex, 0)
	payback := math.Inf(1)
	if benefit > 0 {
		payback = capex / benefit
	}
	roi := 0.0
	if capex > 0 {
		roi = (benefit*float64(n) - capex) / capex * 100
	}
	tco := capex + math.Max(in.MonthlyCost, 0)*float64(n)
	r := math.Pow(1.2, 1.0/12) - 1
	npv := -capex
	for t := 1; t <= n; t++ {
		npv += benefit / math.Pow(1+r, float64(t))
	}
	return model.FinancialAssessment{HoursSavedMonthly: hours, MonthlyBenefit: benefit, PaybackMonths: payback, ROIPct: roi, TCO: tco, NPV: npv}
}

func assessmentAutonomy(score float64, criticalRisk bool) (string, string) {
	if criticalRisk {
		return "assistant", "Из-за критической цены ошибки нужен ассистивный режим."
	}
	switch {
	case score >= 75:
		return "autonomous", "Типовые действия можно передать агенту с минимальным контролем."
	case score >= 55:
		return "human_in_loop", "Агент может готовить решения, а человек подтверждает важные действия."
	case score >= 35:
		return "assistant", "Рекомендуется использовать AI как помощника сотрудника."
	default:
		return "not_ready", "Сначала нужно подготовить процесс, данные или инфраструктуру."
	}
}
