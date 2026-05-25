package service

import (
	"encoding/json"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

func normalizeAuditInput(input *model.AuditInput) {
	input.CompanyName = strings.TrimSpace(input.CompanyName)
	input.PrimaryGoal = strings.TrimSpace(input.PrimaryGoal)
	input.CompanySize = strings.TrimSpace(input.CompanySize)
	input.OperationalHeadcount = strings.TrimSpace(input.OperationalHeadcount)
	input.DataDuplication = strings.TrimSpace(input.DataDuplication)
	input.BudgetRange = strings.TrimSpace(input.BudgetRange)
	input.DecisionTimeline = strings.TrimSpace(input.DecisionTimeline)
	input.ITSystemsOther = strings.TrimSpace(input.ITSystemsOther)

	input.PriorityCriteria = cleanStringSlice(input.PriorityCriteria)
	input.ITSystems = cleanStringSlice(input.ITSystems)

	clean := make([]model.AuditProcessInput, 0, len(input.Processes))
	for _, p := range input.Processes {
		p.Name = strings.TrimSpace(p.Name)
		p.Department = strings.TrimSpace(p.Department)
		p.SystemsOther = strings.TrimSpace(p.SystemsOther)
		p.Bottleneck = strings.TrimSpace(p.Bottleneck)
		p.Systems = cleanStringSlice(p.Systems)
		if p.IntegrationComplexity < 1 {
			p.IntegrationComplexity = 1
		}
		if p.IntegrationComplexity > 5 {
			p.IntegrationComplexity = 5
		}
		if p.Name != "" {
			clean = append(clean, p)
		}
	}
	input.Processes = clean
}

func cleanStringSlice(items []string) []string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func validateAuditInput(in model.AuditInput) error {
	if strings.TrimSpace(in.CompanyName) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.PrimaryGoal) == "" {
		return ErrInvalidInput
	}
	if len(in.PriorityCriteria) == 0 {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.CompanySize) == "" {
		return ErrInvalidInput
	}
	if len(in.ITSystems) == 0 && strings.TrimSpace(in.ITSystemsOther) == "" {
		return ErrInvalidInput
	}
	if len(in.Processes) < model.MinAuditProcesses || len(in.Processes) > model.MaxAuditProcesses {
		return ErrInvalidInput
	}
	for _, p := range in.Processes {
		if err := validateAuditProcess(p); err != nil {
			return err
		}
	}
	return nil
}

func validateAuditProcess(p model.AuditProcessInput) error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(p.Department) == "" {
		return ErrInvalidInput
	}
	if p.FrequencyRange == "" || p.MinutesPerCycleRange == "" || p.FTERange == "" {
		return ErrInvalidInput
	}
	if p.HourlyRateRange == "" || p.ErrorRateRange == "" || p.ErrorCostRange == "" {
		return ErrInvalidInput
	}
	if len(p.Systems) == 0 && strings.TrimSpace(p.SystemsOther) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(p.Bottleneck) == "" {
		return ErrInvalidInput
	}
	if p.Standardization == "" || p.AutomationReadiness == "" || p.ExpectedImpact == "" {
		return ErrInvalidInput
	}
	return nil
}

func parseAuditOutput(content string) (*model.AuditOutput, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var out model.AuditOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.ExecutiveSummary) == "" {
		return nil, ErrInvalidInput
	}
	if len(out.PriorityRanking) == 0 {
		return nil, ErrInvalidInput
	}
	return &out, nil
}

func ParseAuditRun(run *model.ToolRun) (model.AuditInput, model.AuditOutput, error) {
	var in model.AuditInput
	var out model.AuditOutput
	if err := json.Unmarshal(run.Input, &in); err != nil {
		return in, out, err
	}
	if len(run.Output) > 0 {
		if err := json.Unmarshal(run.Output, &out); err != nil {
			return in, out, err
		}
	}
	return in, out, nil
}
