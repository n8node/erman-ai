package service

import (
	"math"
	"sort"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

const auditSavingsRate = 0.7

func rangeMidpoint(rangeKey string, table map[string]float64, fallback float64) float64 {
	if v, ok := table[rangeKey]; ok {
		return v
	}
	return fallback
}

var auditFrequencyMid = map[string]float64{
	"1-10":    5,
	"11-50":   30,
	"51-200":  125,
	"201-500": 350,
	"500+":    750,
}

var auditMinutesMid = map[string]float64{
	"5-15":    10,
	"16-30":   23,
	"31-60":   45,
	"61-120":  90,
	"120+":    180,
}

var auditFTEMid = map[string]float64{
	"1":     1,
	"2-3":   2.5,
	"4-10":  7,
	"11+":   15,
}

var auditHourlyMid = map[string]float64{
	"300-500":   400,
	"501-800":   650,
	"801-1200":  1000,
	"1201-2000": 1600,
	"2000+":     2500,
}

var auditErrorRateMid = map[string]float64{
	"0-2":   1,
	"3-5":   4,
	"6-10":  8,
	"11-20": 15,
	"20+":   25,
}

var auditErrorCostMid = map[string]float64{
	"500-2000":     1250,
	"2001-5000":    3500,
	"5001-15000":   10000,
	"15001-50000":  30000,
	"50000+":       75000,
}

func computeAuditProcessScores(processes []model.AuditProcessInput) []model.AuditProcessScore {
	scores := make([]model.AuditProcessScore, 0, len(processes))
	for _, p := range processes {
		freq := rangeMidpoint(p.FrequencyRange, auditFrequencyMid, 30)
		mins := rangeMidpoint(p.MinutesPerCycleRange, auditMinutesMid, 30)
		fte := rangeMidpoint(p.FTERange, auditFTEMid, 2)
		hourly := rangeMidpoint(p.HourlyRateRange, auditHourlyMid, 650)
		errRate := rangeMidpoint(p.ErrorRateRange, auditErrorRateMid, 5) / 100
		errCost := rangeMidpoint(p.ErrorCostRange, auditErrorCostMid, 3500)

		monthlyHours := freq * (mins / 60) * fte
		laborCost := monthlyHours * hourly
		errorCost := freq * errRate * errCost
		totalCost := laborCost + errorCost

		readiness := readinessWeight(p.AutomationReadiness)
		impact := impactWeight(p.ExpectedImpact)
		std := standardizationWeight(p.Standardization)
		complexity := complexityWeight(p.IntegrationComplexity)

		automationScore := (totalCost / 10000) * readiness * impact * std * complexity
		automationScore = math.Round(automationScore*10) / 10

		quickWin := p.AutomationReadiness == "high" && p.IntegrationComplexity <= 2 && totalCost >= 50000

		scores = append(scores, model.AuditProcessScore{
			Name:             strings.TrimSpace(p.Name),
			MonthlyHours:     math.Round(monthlyHours*10) / 10,
			MonthlyLaborCost: math.Round(laborCost),
			MonthlyErrorCost: math.Round(errorCost),
			TotalMonthlyCost: math.Round(totalCost),
			AutomationScore:  automationScore,
			QuickWin:         quickWin,
			Rationale:        buildScoreRationale(p, totalCost),
		})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].AutomationScore > scores[j].AutomationScore
	})
	for i := range scores {
		scores[i].PriorityRank = i + 1
	}
	return scores
}

func readinessWeight(v string) float64 {
	switch v {
	case "high":
		return 1.2
	case "medium":
		return 1.0
	default:
		return 0.7
	}
}

func impactWeight(v string) float64 {
	switch v {
	case "high":
		return 1.3
	case "medium":
		return 1.0
	default:
		return 0.8
	}
}

func standardizationWeight(v string) float64 {
	switch v {
	case "high":
		return 1.2
	case "medium":
		return 1.0
	default:
		return 0.75
	}
}

func complexityWeight(n int) float64 {
	switch {
	case n <= 1:
		return 1.2
	case n == 2:
		return 1.0
	case n == 3:
		return 0.85
	case n == 4:
		return 0.7
	default:
		return 0.55
	}
}

func buildScoreRationale(p model.AuditProcessInput, totalCost float64) string {
	parts := []string{}
	if totalCost >= 100000 {
		parts = append(parts, "high monthly cost")
	}
	if p.AutomationReadiness == "high" {
		parts = append(parts, "high readiness")
	}
	if p.IntegrationComplexity <= 2 {
		parts = append(parts, "low integration complexity")
	}
	if p.Standardization == "high" {
		parts = append(parts, "standardized process")
	}
	if len(parts) == 0 {
		return "moderate automation potential"
	}
	return strings.Join(parts, ", ")
}

func auditTotals(scores []model.AuditProcessScore) (totalCost, totalSavings float64) {
	for _, s := range scores {
		totalCost += s.TotalMonthlyCost
		totalSavings += s.TotalMonthlyCost * auditSavingsRate
	}
	return math.Round(totalCost), math.Round(totalSavings)
}
