package service

import "github.com/erman-ai/erman-ai/internal/model"

func buildStrategyROISummary(contexts []model.CalculatorRunContext) *model.StrategyROISummary {
	if len(contexts) == 0 {
		return nil
	}
	lines := make([]model.StrategyROILine, 0, len(contexts))
	var totalBenefit, totalNPV float64
	for _, c := range contexts {
		lines = append(lines, model.StrategyROILine{
			RunID:             c.RunID,
			ProcessName:       c.ProcessName,
			NetBenefitMonthly: c.NetBenefitMonthly,
			PaybackMonths:     c.PaybackMonths,
			ROIHorizonPct:     c.ROIHorizonPct,
			NPV:               c.NPV,
			Capex:             c.Capex,
			Recommendation:    c.Recommendation,
		})
		totalBenefit += c.NetBenefitMonthly
		totalNPV += c.NPV
	}
	return &model.StrategyROISummary{
		Lines:               lines,
		TotalMonthlyBenefit: totalBenefit,
		TotalNPV:            totalNPV,
	}
}
