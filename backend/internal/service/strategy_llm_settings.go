package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidStrategyLLMSettings = errors.New("invalid strategy llm settings")

type StrategyLLMSettingsService struct {
	repo *repository.StrategyLLMSettingsRepository
	cfg  *config.Config
}

func NewStrategyLLMSettingsService(repo *repository.StrategyLLMSettingsRepository, cfg *config.Config) *StrategyLLMSettingsService {
	return &StrategyLLMSettingsService{repo: repo, cfg: cfg}
}

func (s *StrategyLLMSettingsService) Get(ctx context.Context) (*model.StrategyLLMSettingsRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := model.DefaultStrategyLLMSettings()
			s.applyPromptDefault(&def)
			return &model.StrategyLLMSettingsRecord{Settings: def}, nil
		}
		return nil, err
	}
	s.applyPromptDefault(&rec.Settings)
	return rec, nil
}

func (s *StrategyLLMSettingsService) GetAdminView(ctx context.Context) (*model.StrategyLLMAdminView, error) {
	rec, err := s.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &model.StrategyLLMAdminView{
		Settings:            rec.Settings,
		Providers:           s.providerStatuses(),
		DefaultSystemPrompt: prompts.DefaultStrategySystemPrompt,
		UpdatedAt:           rec.UpdatedAt,
	}, nil
}

func (s *StrategyLLMSettingsService) Update(ctx context.Context, settings model.StrategyLLMSettings) (*model.StrategyLLMAdminView, error) {
	if err := validateStrategyLLMSettings(settings); err != nil {
		return nil, err
	}
	rec, err := s.repo.Update(ctx, settings)
	if err != nil {
		return nil, err
	}
	s.applyPromptDefault(&rec.Settings)
	return &model.StrategyLLMAdminView{
		Settings:            rec.Settings,
		Providers:           s.providerStatuses(),
		DefaultSystemPrompt: prompts.DefaultStrategySystemPrompt,
		UpdatedAt:           rec.UpdatedAt,
	}, nil
}

func (s *StrategyLLMSettingsService) ResolvedSystemPrompt(settings model.StrategyLLMSettings) string {
	if settings.SystemPrompt != "" {
		return settings.SystemPrompt
	}
	return prompts.DefaultStrategySystemPrompt
}

func (s *StrategyLLMSettingsService) applyPromptDefault(settings *model.StrategyLLMSettings) {
	if settings.SystemPrompt == "" {
		settings.SystemPrompt = prompts.DefaultStrategySystemPrompt
	}
}

func (s *StrategyLLMSettingsService) providerStatuses() []model.LLMProviderStatus {
	return []model.LLMProviderStatus{
		{
			ID:           model.LLMProviderOpenRouter,
			Configured:   s.cfg.OpenRouterAPIKey != "",
			DefaultModel: s.cfg.OpenRouterModelSmart,
			SuggestedModels: []string{
				"anthropic/claude-sonnet-4-5",
				"anthropic/claude-3.5-sonnet",
				"openai/gpt-4.1",
				"google/gemini-2.5-pro-preview",
			},
		},
		{
			ID:           model.LLMProviderDeepSeek,
			Configured:   s.cfg.DeepSeekAPIKey != "",
			DefaultModel: s.cfg.DeepSeekModelDefault,
			SuggestedModels: []string{
				"deepseek-chat",
				"deepseek-reasoner",
			},
		},
	}
}

func validateStrategyLLMSettings(settings model.StrategyLLMSettings) error {
	if settings.Provider != model.LLMProviderOpenRouter && settings.Provider != model.LLMProviderDeepSeek {
		return fmt.Errorf("%w: provider must be openrouter or deepseek", ErrInvalidStrategyLLMSettings)
	}
	if settings.OpenRouterModel == "" || settings.DeepSeekModel == "" {
		return fmt.Errorf("%w: model names required", ErrInvalidStrategyLLMSettings)
	}
	if settings.Temperature < 0 || settings.Temperature > 2 {
		return fmt.Errorf("%w: temperature must be 0–2", ErrInvalidStrategyLLMSettings)
	}
	if settings.MaxTokens < 256 || settings.MaxTokens > 32000 {
		return fmt.Errorf("%w: max_tokens must be 256–32000", ErrInvalidStrategyLLMSettings)
	}
	return nil
}
