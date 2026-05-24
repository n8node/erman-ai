package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidStrategyLLMSettings = errors.New("invalid strategy llm settings")

type StrategyLLMSettingsService struct {
	repo *repository.StrategyLLMSettingsRepository
	cfg  *config.Config
	llm  *LLMService
}

func NewStrategyLLMSettingsService(
	repo *repository.StrategyLLMSettingsRepository,
	cfg *config.Config,
	llm *LLMService,
) *StrategyLLMSettingsService {
	return &StrategyLLMSettingsService{repo: repo, cfg: cfg, llm: llm}
}

func (s *StrategyLLMSettingsService) GetStored(ctx context.Context) (*model.StrategyLLMSettingsRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := model.DefaultStrategyLLMStoredConfig()
			s.applyPromptDefault(&def.StrategyLLMSettings)
			return &model.StrategyLLMSettingsRecord{Config: def}, nil
		}
		return nil, err
	}
	s.applyPromptDefault(&rec.Config.StrategyLLMSettings)
	if rec.Config.OpenRouterModels == nil {
		rec.Config.OpenRouterModels = []string{}
	}
	if rec.Config.DeepSeekModels == nil {
		rec.Config.DeepSeekModels = []string{}
	}
	return rec, nil
}

func (s *StrategyLLMSettingsService) GetAdminView(ctx context.Context) (*model.StrategyLLMAdminView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildAdminView(rec), nil
}

func (s *StrategyLLMSettingsService) Update(ctx context.Context, req model.StrategyLLMAdminUpdateRequest) (*model.StrategyLLMAdminView, error) {
	if err := validateStrategyLLMSettings(req.Settings); err != nil {
		return nil, err
	}

	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}

	cfg := rec.Config
	cfg.StrategyLLMSettings = req.Settings
	if strings.TrimSpace(req.OpenRouterAPIKey) != "" {
		cfg.OpenRouterAPIKey = strings.TrimSpace(req.OpenRouterAPIKey)
	}
	if strings.TrimSpace(req.DeepSeekAPIKey) != "" {
		cfg.DeepSeekAPIKey = strings.TrimSpace(req.DeepSeekAPIKey)
	}

	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	s.applyPromptDefault(&updated.Config.StrategyLLMSettings)
	return s.buildAdminView(updated), nil
}

func (s *StrategyLLMSettingsService) TestConnection(ctx context.Context, provider model.LLMProvider) (*model.StrategyLLMTestConnectionResult, error) {
	if provider != model.LLMProviderOpenRouter && provider != model.LLMProviderDeepSeek {
		return nil, fmt.Errorf("%w: invalid provider", ErrInvalidStrategyLLMSettings)
	}

	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}

	apiKey := s.llm.ResolveKey(provider, rec.Config.OpenRouterAPIKey, rec.Config.DeepSeekAPIKey)
	if apiKey == "" {
		return &model.StrategyLLMTestConnectionResult{
			Provider: provider,
			OK:       false,
			Message:  "api key not configured",
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	models, err := s.llm.ListModels(ctx, provider, apiKey)
	if err != nil {
		return &model.StrategyLLMTestConnectionResult{
			Provider: provider,
			OK:       false,
			Message:  err.Error(),
		}, nil
	}

	cfg := rec.Config
	if provider == model.LLMProviderDeepSeek {
		cfg.DeepSeekModels = models
		if cfg.DeepSeekModel == "" || !contains(models, cfg.DeepSeekModel) {
			cfg.DeepSeekModel = pickDefaultModel(models, s.cfg.DeepSeekModelDefault)
		}
	} else {
		cfg.OpenRouterModels = models
		if cfg.OpenRouterModel == "" || !contains(models, cfg.OpenRouterModel) {
			cfg.OpenRouterModel = pickDefaultModel(models, s.cfg.OpenRouterModelSmart)
		}
	}

	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	_ = updated

	return &model.StrategyLLMTestConnectionResult{
		Provider: provider,
		OK:       true,
		Message:  fmt.Sprintf("connected, %d models loaded", len(models)),
		Models:   models,
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

func (s *StrategyLLMSettingsService) buildAdminView(rec *model.StrategyLLMSettingsRecord) *model.StrategyLLMAdminView {
	cfg := rec.Config
	return &model.StrategyLLMAdminView{
		Settings:            cfg.StrategyLLMSettings,
		Providers:           s.providerStatuses(cfg),
		DefaultSystemPrompt: prompts.DefaultStrategySystemPrompt,
		UpdatedAt:           rec.UpdatedAt,
	}
}

func (s *StrategyLLMSettingsService) providerStatuses(cfg model.StrategyLLMStoredConfig) []model.LLMProviderStatus {
	orKey := s.llm.ResolveOpenRouterKey(cfg.OpenRouterAPIKey)
	dsKey := s.llm.ResolveDeepSeekKey(cfg.DeepSeekAPIKey)

	return []model.LLMProviderStatus{
		{
			ID:           model.LLMProviderOpenRouter,
			Configured:   orKey != "",
			KeyHint:      model.MaskAPIKey(orKey),
			DefaultModel: s.cfg.OpenRouterModelSmart,
			Models:       cfg.OpenRouterModels,
		},
		{
			ID:           model.LLMProviderDeepSeek,
			Configured:   dsKey != "",
			KeyHint:      model.MaskAPIKey(dsKey),
			DefaultModel: s.cfg.DeepSeekModelDefault,
			Models:       cfg.DeepSeekModels,
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

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func pickDefaultModel(models []string, preferred string) string {
	if preferred != "" && contains(models, preferred) {
		return preferred
	}
	if len(models) > 0 {
		return models[0]
	}
	return preferred
}
