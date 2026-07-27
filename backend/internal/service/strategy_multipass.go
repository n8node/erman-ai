package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
)

type llmUsageAggregate struct {
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

func (a *llmUsageAggregate) add(r *LLMCompletionResult) {
	if r == nil {
		return
	}
	a.Model = r.Model
	a.PromptTokens += r.PromptTokens
	a.CompletionTokens += r.CompletionTokens
	a.TotalTokens += r.TotalTokens
}

type strategyPass struct {
	id           string
	promptSuffix string
}

var consultingPasses = []strategyPass{
	{id: "context", promptSuffix: prompts.StrategyPassContext},
	{id: "analysis", promptSuffix: prompts.StrategyPassAnalysis},
	{id: "plan", promptSuffix: prompts.StrategyPassPlan},
	{id: "governance", promptSuffix: prompts.StrategyPassGovernance},
}

func (s *StrategyService) generateStrategyOutput(
	ctx context.Context,
	runID string,
	input model.StrategyInput,
	baseSystemPrompt string,
	settings model.StrategyLLMSettings,
	creds LLMCredentials,
	baseURL string,
) (*model.StrategyOutput, llmUsageAggregate, error) {
	mode := normalizeReportMode(input.ReportMode)
	var usage llmUsageAggregate
	apiKey := s.llm.ResolveKey(settings.Provider, creds)

	userPayload, _ := json.Marshal(input)
	baseUser := string(userPayload)

	if mode == model.StrategyReportConsulting {
		var merged model.StrategyOutput
		for i, pass := range consultingPasses {
			s.publishPhase(runID, pass.id, "active")
			system := baseSystemPrompt + "\n\n" + pass.promptSuffix
			req := LLMCompletionRequest{
				Provider:     settings.Provider,
				Model:        settings.ActiveModel(),
				SystemPrompt: system,
				UserPrompt:   baseUser,
				Temperature:  settings.Temperature,
				MaxTokens:    passMaxTokens(settings.MaxTokens),
				APIKey:       apiKey,
				BaseURL:      baseURL,
				FolderID:     creds.YandexFolderID,
			}
			result, err := s.streamWithRetry(ctx, runID, req)
			if err != nil {
				return nil, usage, fmt.Errorf("pass %s: %w", pass.id, err)
			}
			usage.add(result)
			if err := mergeStrategyJSON(&merged, result.Content); err != nil {
				return nil, usage, fmt.Errorf("pass %s parse: %w", pass.id, err)
			}
			s.publishPhase(runID, pass.id, "done")
			s.logger.Info("strategy pass complete", "run_id", runID, "pass", pass.id, "index", i+1, "total", len(consultingPasses))
		}
		if err := s.postValidateAndExpand(ctx, runID, &merged, input, baseSystemPrompt, settings, creds, baseURL, mode, &usage); err != nil {
			return nil, usage, err
		}
		return &merged, usage, nil
	}

	system := baseSystemPrompt + "\n\n" + prompts.StrategyExtendedFieldsPrompt
	req := LLMCompletionRequest{
		Provider:     settings.Provider,
		Model:        settings.ActiveModel(),
		SystemPrompt: system,
		UserPrompt:   baseUser,
		Temperature:  settings.Temperature,
		MaxTokens:    settings.MaxTokens,
		APIKey:       apiKey,
		BaseURL:      baseURL,
		FolderID:     creds.YandexFolderID,
	}
	result, err := s.streamWithRetry(ctx, runID, req)
	if err != nil {
		return nil, usage, err
	}
	usage.add(result)
	output, err := parseStrategyOutput(result.Content)
	if err != nil {
		return nil, usage, err
	}
	if err := s.postValidateAndExpand(ctx, runID, output, input, baseSystemPrompt, settings, creds, baseURL, mode, &usage); err != nil {
		return nil, usage, err
	}
	return output, usage, nil
}

func passMaxTokens(settingsMax int) int {
	if settingsMax <= 0 {
		return 8192
	}
	perPass := settingsMax / 2
	if perPass < 8192 {
		perPass = 8192
	}
	if perPass > 16384 {
		perPass = 16384
	}
	return perPass
}

func (s *StrategyService) postValidateAndExpand(
	ctx context.Context,
	runID string,
	output *model.StrategyOutput,
	input model.StrategyInput,
	baseSystemPrompt string,
	settings model.StrategyLLMSettings,
	creds LLMCredentials,
	baseURL, mode string,
	usage *llmUsageAggregate,
) error {
	consulting := mode == model.StrategyReportConsulting
	if !needsVolumeExpansion(output, mode) && len(thinStrategyFields(output, consulting)) == 0 {
		return nil
	}
	thin := thinStrategyFields(output, consulting)
	if len(thin) == 0 && needsVolumeExpansion(output, mode) {
		thin = []string{"executive_summary", "current_situation", "process_analysis", "ai_use_cases"}
	}
	s.publishPhase(runID, "pass_expand", "active")
	draft, _ := json.Marshal(output)
	expandPrompt := fmt.Sprintf(prompts.StrategyExpandPrompt, strings.Join(thin, ", "), string(draft))
	req := LLMCompletionRequest{
		Provider:     settings.Provider,
		Model:        settings.ActiveModel(),
		SystemPrompt: baseSystemPrompt,
		UserPrompt:   expandPrompt,
		Temperature:  settings.Temperature,
		MaxTokens:    passMaxTokens(settings.MaxTokens),
		APIKey:       s.llm.ResolveKey(settings.Provider, creds),
		BaseURL:      baseURL,
		FolderID:     creds.YandexFolderID,
	}
	result, err := s.streamWithRetry(ctx, runID, req)
	if err != nil {
		s.logger.Warn("strategy expansion pass failed, using draft", "run_id", runID, "error", err)
		s.publishPhase(runID, "pass_expand", "done")
		return nil
	}
	usage.add(result)
	if err := mergeStrategyJSON(output, result.Content); err != nil {
		s.logger.Warn("strategy expansion parse failed", "run_id", runID, "error", err)
	}
	s.publishPhase(runID, "pass_expand", "done")
	return nil
}
