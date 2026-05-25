package service

import (
	"encoding/json"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

func validateProposalInput(in model.ProposalInput) error {
	in.ProposalScenario = model.NormalizeProposalScenario(in.ProposalScenario)

	if strings.TrimSpace(in.ClientCompany) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.ClientProblem) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.SolutionName) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.SolutionDescription) == "" {
		return ErrInvalidInput
	}
	if len(in.Deliverables) == 0 {
		return ErrInvalidInput
	}
	if in.IncludePricing && in.ProjectCostRub <= 0 {
		return ErrInvalidInput
	}
	if in.TimelineWeeks <= 0 || in.TimelineWeeks > 104 {
		return ErrInvalidInput
	}
	if in.IncludePricing && strings.TrimSpace(in.PaymentSchedule) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.SenderCompany) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.SenderContact) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(in.SenderEmail) == "" {
		return ErrInvalidInput
	}

	switch in.ProposalScenario {
	case model.ProposalScenarioAfterContact:
		if strings.TrimSpace(in.PriorContactSummary) == "" {
			return ErrInvalidInput
		}
	case model.ProposalScenarioColdOutreach:
		if strings.TrimSpace(in.ProblemSource) == "" {
			return ErrInvalidInput
		}
	case model.ProposalScenarioProactiveOffer:
		// calculator link optional
	default:
		return ErrInvalidInput
	}

	return nil
}

func parseProposalOutput(content string) (*model.ProposalOutput, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var out model.ProposalOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.Greeting) == "" {
		return nil, ErrInvalidInput
	}
	if strings.TrimSpace(out.ProposedSolution) == "" {
		return nil, ErrInvalidInput
	}
	if len(out.ScopeIncluded) == 0 {
		return nil, ErrInvalidInput
	}
	return &out, nil
}

func ParseProposalRun(run *model.ToolRun) (model.ProposalInput, model.ProposalOutput, error) {
	var in model.ProposalInput
	var out model.ProposalOutput
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
