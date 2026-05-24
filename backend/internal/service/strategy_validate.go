package service

import (
	"encoding/json"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

const (
	consultingMinTotalChars = 18000
	standardMinTotalChars   = 8000
)

func normalizeReportMode(mode string) string {
	if mode == model.StrategyReportStandard {
		return model.StrategyReportStandard
	}
	return model.StrategyReportConsulting
}

func strategyOutputCharCount(o *model.StrategyOutput) int {
	if o == nil {
		return 0
	}
	n := len(o.ExecutiveSummary) + len(o.CurrentSituation) + len(o.GoalsAndRationale) +
		len(o.DataAndInfrastructure) + len(o.TeamAndTraining) + len(o.ArchitectureOverview) +
		len(o.DataGovernance) + len(o.EthicsAndCompliance) + len(o.StrategyAdjustmentPlan) +
		len(o.BudgetOverview.Summary)
	for _, p := range o.ProcessAnalysis {
		n += len(p.Name) + len(p.CurrentState) + len(p.PainPoints) + len(p.AiPotential)
	}
	for _, u := range o.AiUseCases {
		n += len(u.Title) + len(u.Description)
	}
	for _, s := range o.RecommendedSolutions {
		n += len(s.Title) + len(s.Description) + len(s.Rationale)
	}
	for _, ph := range o.ImplementationPlan {
		n += len(ph.Title) + len(ph.Duration)
		for _, d := range ph.Deliverables {
			n += len(d)
		}
	}
	for _, step := range o.Next30Days {
		n += len(step)
	}
	for _, r := range o.Risks {
		n += len(r.Risk) + len(r.Mitigation)
	}
	for _, m := range o.MaturityMatrix {
		n += len(m.Criterion) + len(m.CurrentLevel) + len(m.TargetLevel) + len(m.Gap)
	}
	for _, d := range o.Diagrams {
		n += len(d.Title) + len(d.Mermaid)
	}
	return n
}

func thinStrategyFields(o *model.StrategyOutput, consulting bool) []string {
	var fields []string
	minExec := 1200
	if consulting {
		minExec = 1800
	}
	if len(o.ExecutiveSummary) < minExec {
		fields = append(fields, "executive_summary")
	}
	if len(o.CurrentSituation) < 1000 {
		fields = append(fields, "current_situation")
	}
	if len(o.GoalsAndRationale) < 700 {
		fields = append(fields, "goals_and_rationale")
	}
	if len(o.DataAndInfrastructure) < 800 {
		fields = append(fields, "data_and_infrastructure")
	}
	if len(o.ProcessAnalysis) < 3 {
		fields = append(fields, "process_analysis")
	}
	if len(o.AiUseCases) < 5 {
		fields = append(fields, "ai_use_cases")
	}
	if len(o.SuccessMetrics) < 6 {
		fields = append(fields, "success_metrics")
	}
	if len(o.Risks) < 5 {
		fields = append(fields, "risks")
	}
	if consulting {
		if len(o.MaturityMatrix) < 5 {
			fields = append(fields, "maturity_matrix")
		}
		if len(o.PriorityMatrix) < 5 {
			fields = append(fields, "priority_matrix")
		}
		if len(o.Diagrams) < 2 {
			fields = append(fields, "diagrams")
		}
	}
	return fields
}

func needsVolumeExpansion(o *model.StrategyOutput, mode string) bool {
	min := standardMinTotalChars
	if mode == model.StrategyReportConsulting {
		min = consultingMinTotalChars
	}
	return strategyOutputCharCount(o) < min
}

func mergeStrategyJSON(dst *model.StrategyOutput, content string) error {
	content = cleanLLMJSON(content)
	return json.Unmarshal([]byte(content), dst)
}

func cleanLLMJSON(content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}
