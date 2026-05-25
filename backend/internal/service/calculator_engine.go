package service

import (
	"fmt"
	"math"

	"github.com/erman-ai/erman-ai/internal/i18n"
	"github.com/erman-ai/erman-ai/internal/model"
)

const hoursPerFTEMonth = 168.0

func ResolveHoursSaved(in model.CalculatorInput) float64 {
	if in.HmMode == "direct" {
		return in.HoursSavedMonth
	}
	if len(in.ProcessSteps) == 0 {
		return in.HoursSavedMonth
	}
	var totalMin float64
	for _, s := range in.ProcessSteps {
		totalMin += s.MinutesPerUnit
	}
	if in.UnitsPerMonth <= 0 || in.AutomationPct <= 0 {
		return in.HoursSavedMonth
	}
	return totalMin * in.UnitsPerMonth * (in.AutomationPct / 100) / 60
}

func TotalMinutesPerUnit(steps []model.ProcessStep) float64 {
	var total float64
	for _, s := range steps {
		total += s.MinutesPerUnit
	}
	return total
}

func CalculateExcel(in model.CalculatorInput, locale string) model.CalculatorOutput {
	locale = i18n.NormalizeLocale(locale)
	Hm := ResolveHoursSaved(in)
	L := in.Utilization
	if L <= 0 || L > 1 {
		L = 0.8
	}
	Ch := in.HourlyCostLoaded
	Sm := in.OtherBenefitMonthly
	Om := in.MonthlySolutionCost
	I0 := in.Capex
	n := in.HorizonMonths
	if n <= 0 {
		n = 12
	}
	r := in.DiscountRateAnnual
	if r <= 0 {
		r = 0.2
	}

	Bm := Hm*Ch + Sm - Om
	FTE := 0.0
	if L > 0 {
		FTE = Hm / (hoursPerFTEMonth * L)
	}

	var payback float64
	if Bm > 0 {
		payback = I0 / Bm
	} else {
		payback = math.Inf(1)
	}

	var roiPct float64
	if I0 > 0 {
		roiPct = ((Bm * float64(n)) - I0) / I0 * 100
	}

	tco := I0 + Om*float64(n)
	npv := calculateNPV(I0, Bm, n, r)
	payroll := Hm * Ch

	minPerUnit := TotalMinutesPerUnit(in.ProcessSteps)
	autoPct := in.AutomationPct
	if autoPct <= 0 {
		autoPct = 100
	}
	var monthlyHoursBefore, monthlyHoursAfter float64
	if minPerUnit > 0 && in.UnitsPerMonth > 0 {
		monthlyHoursBefore = minPerUnit * in.UnitsPerMonth / 60
		monthlyHoursAfter = monthlyHoursBefore - Hm
		if monthlyHoursAfter < 0 {
			monthlyHoursAfter = 0
		}
	}

	rec, text := recommendationExcel(payback, Bm, roiPct, locale)

	return model.CalculatorOutput{
		HoursSavedMonth:    round2(Hm),
		MinutesPerUnit:     round2(minPerUnit),
		NetBenefitMonthly:  round2(Bm),
		FTE:                round3(FTE),
		PaybackMonths:      round2(payback),
		ROIHorizonPct:      round2(roiPct),
		TCOHorizon:         round2(tco),
		NPV:                round2(npv),
		PayrollSavings:     round2(payroll),
		MonthlyHoursBefore: round2(monthlyHoursBefore),
		MonthlyHoursAfter:  round2(monthlyHoursAfter),
		Recommendation:     rec,
		RecommendationText: text,
		KPIRows:            buildKPIRows(in, locale, minPerUnit, autoPct, monthlyHoursBefore, monthlyHoursAfter, Hm, payroll, payback),
	}
}

func calculateNPV(I0, monthlyBenefit float64, months int, annualRate float64) float64 {
	rm := math.Pow(1+annualRate, 1.0/12.0) - 1
	npv := -I0
	for t := 1; t <= months; t++ {
		npv += monthlyBenefit / math.Pow(1+rm, float64(t))
	}
	return npv
}

func buildKPIRows(
	in model.CalculatorInput,
	locale string,
	minPerUnit, autoPct, hoursBefore, hoursAfter, Hm, payroll, payback float64,
) []model.KPIRow {
	labels := i18n.CalculatorKPIFor(locale)

	minAfter := minPerUnit * (1 - autoPct/100)
	var timeChange string
	if minPerUnit > 0 {
		pct := (1 - minAfter/minPerUnit) * 100
		timeChange = formatPctChange(pct, locale)
	}

	errBefore := in.ErrorRateBeforePct
	if errBefore <= 0 {
		errBefore = 5
	}

	var volumeChange string
	tBefore := in.ThroughputBeforePerDay
	var tAfter float64
	if tBefore > 0 && minAfter > 0 && minPerUnit > 0 {
		tAfter = tBefore * (minPerUnit / minAfter)
		volumeChange = formatMultiplier(tAfter/tBefore, locale)
	}

	paybackStr := "—"
	if payback > 0 && payback < 1e6 && !math.IsInf(payback, 0) {
		paybackStr = formatNum(payback, 1) + labels.MonthForecastSuffix
	}

	rows := []model.KPIRow{
		{
			Key: labels.TimeKey, Label: labels.Time,
			Before: formatMin(minPerUnit, locale),
			After:  formatMin(minAfter, locale),
			Change: timeChange,
		},
		{
			Key: labels.VolumeKey, Label: labels.Volume,
			Before: formatThroughput(tBefore, locale),
			After:  formatThroughput(tAfter, locale),
			Change: volumeChange,
		},
		{
			Key: labels.ErrorsKey, Label: labels.Errors,
			Before: formatErrBefore(errBefore, locale),
			After:  labels.ErrorsAfter,
			Change: labels.ErrorsChange,
		},
		{
			Key: labels.HoursKey, Label: labels.Hours,
			Before: formatHours(hoursBefore, locale),
			After:  formatHours(hoursAfter, locale),
			Change: formatHoursDelta(hoursBefore-hoursAfter, autoPct, locale),
		},
		{
			Key: labels.PayrollKey, Label: labels.Payroll,
			Before: "0 ₽",
			After:  formatRub(payroll, locale),
			Change: formatRub(payroll, locale),
		},
		{
			Key: labels.PaybackKey, Label: labels.Payback,
			Before: "—",
			After:  paybackStr,
			Change: "—",
		},
	}
	return rows
}

func recommendationExcel(payback, Bm, roiPct float64, locale string) (model.Recommendation, string) {
	copy := i18n.CalculatorRecommendationFor(locale)
	if Bm <= 0 || math.IsInf(payback, 1) || payback > 24 {
		return model.RecommendationNotRecommended, copy.NotRecommended
	}
	if payback >= 12 || roiPct <= 0 {
		return model.RecommendationConsider, copy.Consider
	}
	return model.RecommendationAutomate, copy.Automate
}

func round3(v float64) float64 {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return 0
	}
	return math.Round(v*1000) / 1000
}

func formatMin(v float64, locale string) string {
	if v <= 0 {
		return "—"
	}
	copy := i18n.CalculatorKPIFor(locale)
	return formatNum(v, 0) + copy.MinSuffix
}

func formatHours(v float64, locale string) string {
	if v <= 0 {
		return "—"
	}
	copy := i18n.CalculatorKPIFor(locale)
	return formatNum(v, 0) + copy.HourSuffix
}

func formatRub(v float64, locale string) string {
	if v <= 0 {
		return "0 ₽"
	}
	copy := i18n.CalculatorKPIFor(locale)
	return formatNum(v, 0) + copy.RubSuffix
}

func formatThroughput(v float64, locale string) string {
	if v <= 0 {
		return "—"
	}
	copy := i18n.CalculatorKPIFor(locale)
	return formatNum(v, 0) + copy.ThroughputSuffix
}

func formatErrBefore(pct float64, locale string) string {
	copy := i18n.CalculatorKPIFor(locale)
	return "~" + formatNum(pct, 0) + copy.ErrBeforeSuffix
}

func formatPctChange(pct float64, locale string) string {
	if locale == "en" {
		return formatNum(-pct, 0) + "%"
	}
	return formatNum(-pct, 0) + "%"
}

func formatHoursDelta(delta, autoPct float64, locale string) string {
	if delta <= 0 {
		return "—"
	}
	copy := i18n.CalculatorKPIFor(locale)
	sep := "-"
	if locale == "ru" {
		sep = "−"
	}
	return sep + formatNum(delta, 0) + copy.HourSuffix + " (" + formatNum(autoPct, 0) + "%)"
}

func formatMultiplier(m float64, locale string) string {
	if m <= 0 {
		return "—"
	}
	pct := (m - 1) * 100
	if locale == "en" {
		return formatNum(pct, 0) + "% (" + formatNum(m, 1) + "x)"
	}
	return formatNum(pct, 0) + "% (" + formatNum(m, 1) + "×)"
}

func formatNum(v float64, dec int) string {
	p := math.Pow(10, float64(dec))
	return fmtNum(math.Round(v*p) / p, dec)
}

func fmtNum(v float64, dec int) string {
	if dec == 0 {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.*f", dec, v)
}

// fmt import needed - use fmt.Sprintf in fmtNum - add import fmt at top
