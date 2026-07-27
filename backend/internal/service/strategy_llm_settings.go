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
			s.applyProposalPromptDefault(&def.StrategyLLMSettings)
			s.applyYandexModelDefault(&def)
			return &model.StrategyLLMSettingsRecord{Config: def}, nil
		}
		return nil, err
	}
	s.applyPromptDefault(&rec.Config.StrategyLLMSettings)
	s.applyProposalPromptDefault(&rec.Config.StrategyLLMSettings)
	s.applyYandexModelDefault(&rec.Config)
	if rec.Config.OpenRouterModels == nil {
		rec.Config.OpenRouterModels = []string{}
	}
	if rec.Config.DeepSeekModels == nil {
		rec.Config.DeepSeekModels = []string{}
	}
	if rec.Config.YandexModels == nil {
		rec.Config.YandexModels = []string{}
	}
	if rec.Config.Pricing == nil {
		rec.Config.Pricing = map[model.LLMProvider]model.LLMProviderPricing{}
	}
	if rec.Config.ModelPricing == nil {
		rec.Config.ModelPricing = map[string]model.LLMProviderPricing{}
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
	if err := validateLLMPricing(req.Pricing); err != nil {
		return nil, err
	}
	if err := validateModelPricing(req.ModelPricing); err != nil {
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
	if strings.TrimSpace(req.YandexAPIKey) != "" {
		cfg.YandexAPIKey = strings.TrimSpace(req.YandexAPIKey)
	}
	if strings.TrimSpace(req.YandexFolderID) != "" {
		cfg.YandexFolderID = strings.TrimSpace(req.YandexFolderID)
	}
	if req.Pricing != nil {
		if cfg.Pricing == nil {
			cfg.Pricing = map[model.LLMProvider]model.LLMProviderPricing{}
		}
		for provider, pricing := range req.Pricing {
			cfg.Pricing[provider] = pricing
		}
	}
	if req.ModelPricing != nil {
		if cfg.ModelPricing == nil {
			cfg.ModelPricing = map[string]model.LLMProviderPricing{}
		}
		for modelID, pricing := range req.ModelPricing {
			cfg.ModelPricing[modelID] = pricing
		}
	}

	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	s.applyPromptDefault(&updated.Config.StrategyLLMSettings)
	s.applyProposalPromptDefault(&updated.Config.StrategyLLMSettings)
	s.applyYandexModelDefault(&updated.Config)
	return s.buildAdminView(updated), nil
}

func (s *StrategyLLMSettingsService) TestConnection(ctx context.Context, provider model.LLMProvider) (*model.StrategyLLMTestConnectionResult, error) {
	if provider != model.LLMProviderOpenRouter && provider != model.LLMProviderDeepSeek && provider != model.LLMProviderYandex {
		return nil, fmt.Errorf("%w: invalid provider", ErrInvalidStrategyLLMSettings)
	}

	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}

	creds := s.llm.CredentialsFromStored(rec.Config)
	if s.llm.ResolveKey(provider, creds) == "" {
		return &model.StrategyLLMTestConnectionResult{
			Provider: provider,
			OK:       false,
			Message:  "api key not configured",
		}, nil
	}
	if provider == model.LLMProviderYandex && creds.YandexFolderID == "" {
		return &model.StrategyLLMTestConnectionResult{
			Provider: provider,
			OK:       false,
			Message:  "yandex folder id not configured",
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	models, err := s.llm.ListModels(ctx, provider, creds)
	if err != nil {
		return &model.StrategyLLMTestConnectionResult{
			Provider: provider,
			OK:       false,
			Message:  err.Error(),
		}, nil
	}

	cfg := rec.Config
	switch provider {
	case model.LLMProviderDeepSeek:
		cfg.DeepSeekModels = models
		if cfg.DeepSeekModel == "" || !contains(models, cfg.DeepSeekModel) {
			cfg.DeepSeekModel = pickDefaultModel(models, s.cfg.DeepSeekModelDefault)
		}
	case model.LLMProviderYandex:
		cfg.YandexModels = models
		if cfg.YandexModel == "" || !IsYandexChatModel(cfg.YandexModel) || !containsYandexModel(models, cfg.YandexModel, creds.YandexFolderID) {
			cfg.YandexModel = PickYandexChatModel(models, creds.YandexFolderID, s.cfg.YandexModelSmart)
		}
	default:
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

func (s *StrategyLLMSettingsService) UsageCosts(
	cfg model.StrategyLLMStoredConfig,
	provider model.LLMProvider,
	modelName string,
	promptTokens, completionTokens int,
) (costUSD, costRUB float64) {
	return UsageLogCosts(provider, modelName, cfg.ModelPricing, cfg.Pricing, s.cfg, promptTokens, completionTokens)
}

func (s *StrategyLLMSettingsService) ResolvedSystemPrompt(settings model.StrategyLLMSettings) string {
	if settings.SystemPrompt != "" {
		return settings.SystemPrompt
	}
	return prompts.DefaultStrategySystemPrompt
}

func (s *StrategyLLMSettingsService) ResolvedProposalSystemPrompt(settings model.StrategyLLMSettings, locale string) string {
	base := settings.ProposalSystemPrompt
	if base == "" {
		base = prompts.DefaultProposalSystemPrompt
	}
	return prompts.MaterializeProposalPrompt(base, locale)
}

func (s *StrategyLLMSettingsService) applyPromptDefault(settings *model.StrategyLLMSettings) {
	if settings.SystemPrompt == "" {
		settings.SystemPrompt = prompts.DefaultStrategySystemPrompt
	}
}

func (s *StrategyLLMSettingsService) applyProposalPromptDefault(settings *model.StrategyLLMSettings) {
	if settings.ProposalSystemPrompt == "" {
		settings.ProposalSystemPrompt = prompts.DefaultProposalSystemPrompt
	}
}

func (s *StrategyLLMSettingsService) applyYandexModelDefault(cfg *model.StrategyLLMStoredConfig) {
	if cfg.YandexModel != "" {
		return
	}
	folderID := s.llm.ResolveYandexFolderID(cfg.YandexFolderID)
	cfg.YandexModel = YandexModelURI(folderID, s.cfg.YandexModelSmart)
	if folderID == "" {
		cfg.YandexModel = s.cfg.YandexModelSmart
	}
}

func (s *StrategyLLMSettingsService) buildAdminView(rec *model.StrategyLLMSettingsRecord) *model.StrategyLLMAdminView {
	cfg := rec.Config
	pricing := mergedPricing(cfg.Pricing, s.cfg)
	return &model.StrategyLLMAdminView{
		Settings:                    cfg.StrategyLLMSettings,
		Providers:                   s.providerStatuses(cfg),
		Pricing:                     pricing,
		ModelPricing:                cloneModelPricing(cfg.ModelPricing),
		DefaultSystemPrompt:         prompts.DefaultStrategySystemPrompt,
		DefaultProposalSystemPrompt: prompts.DefaultProposalSystemPrompt,
		UpdatedAt:                   rec.UpdatedAt,
	}
}

func cloneModelPricing(src map[string]model.LLMProviderPricing) map[string]model.LLMProviderPricing {
	if src == nil {
		return map[string]model.LLMProviderPricing{}
	}
	out := make(map[string]model.LLMProviderPricing, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func mergedPricing(stored map[model.LLMProvider]model.LLMProviderPricing, cfg *config.Config) map[model.LLMProvider]model.LLMProviderPricing {
	out := map[model.LLMProvider]model.LLMProviderPricing{
		model.LLMProviderOpenRouter: ResolveProviderPricing(model.LLMProviderOpenRouter, stored, cfg),
		model.LLMProviderDeepSeek:   ResolveProviderPricing(model.LLMProviderDeepSeek, stored, cfg),
		model.LLMProviderYandex:     ResolveProviderPricing(model.LLMProviderYandex, stored, cfg),
	}
	if stored != nil {
		for k, v := range stored {
			if v.InputPer1K > 0 || v.OutputPer1K > 0 {
				if v.Currency == "" {
					v.Currency = defaultCurrencyForProvider(k)
				}
				out[k] = v
			}
		}
	}
	return out
}

func (s *StrategyLLMSettingsService) providerStatuses(cfg model.StrategyLLMStoredConfig) []model.LLMProviderStatus {
	creds := s.llm.CredentialsFromStored(cfg)

	return []model.LLMProviderStatus{
		{
			ID:           model.LLMProviderOpenRouter,
			Configured:   creds.OpenRouterKey != "",
			KeyHint:      model.MaskAPIKey(creds.OpenRouterKey),
			DefaultModel: s.cfg.OpenRouterModelSmart,
			Models:       cfg.OpenRouterModels,
		},
		{
			ID:           model.LLMProviderDeepSeek,
			Configured:   creds.DeepSeekKey != "",
			KeyHint:      model.MaskAPIKey(creds.DeepSeekKey),
			DefaultModel: s.cfg.DeepSeekModelDefault,
			Models:       cfg.DeepSeekModels,
		},
		{
			ID:           model.LLMProviderYandex,
			Configured:   creds.YandexKey != "" && creds.YandexFolderID != "",
			KeyHint:      model.MaskAPIKey(creds.YandexKey),
			FolderHint:   model.MaskFolderID(creds.YandexFolderID),
			DefaultModel: YandexModelURI(creds.YandexFolderID, s.cfg.YandexModelSmart),
			Models:       MergeYandexChatModels(cfg.YandexModels, creds.YandexFolderID),
		},
	}
}

func validateStrategyLLMSettings(settings model.StrategyLLMSettings) error {
	switch settings.Provider {
	case model.LLMProviderOpenRouter, model.LLMProviderDeepSeek, model.LLMProviderYandex:
	default:
		return fmt.Errorf("%w: provider must be openrouter, deepseek or yandex", ErrInvalidStrategyLLMSettings)
	}
	if settings.OpenRouterModel == "" || settings.DeepSeekModel == "" || settings.YandexModel == "" {
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

func validateLLMPricing(pricing map[model.LLMProvider]model.LLMProviderPricing) error {
	if pricing == nil {
		return nil
	}
	for provider, p := range pricing {
		if err := validatePricingValues(string(provider), p); err != nil {
			return err
		}
	}
	return nil
}

func validateModelPricing(pricing map[string]model.LLMProviderPricing) error {
	if pricing == nil {
		return nil
	}
	for modelID, p := range pricing {
		if err := validatePricingValues(modelID, p); err != nil {
			return err
		}
	}
	return nil
}

func validatePricingValues(label string, p model.LLMProviderPricing) error {
	if p.InputPer1K < 0 || p.OutputPer1K < 0 {
		return fmt.Errorf("%w: pricing for %s must be non-negative", ErrInvalidStrategyLLMSettings, label)
	}
	if p.InputPer1K > 10000 || p.OutputPer1K > 10000 {
		return fmt.Errorf("%w: pricing for %s out of range", ErrInvalidStrategyLLMSettings, label)
	}
	cur := strings.ToUpper(strings.TrimSpace(p.Currency))
	if cur != "" && cur != "RUB" && cur != "USD" {
		return fmt.Errorf("%w: pricing currency must be RUB or USD", ErrInvalidStrategyLLMSettings)
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

func containsYandexModel(models []string, value, folderID string) bool {
	if contains(models, value) {
		return true
	}
	return contains(models, YandexModelURI(folderID, value))
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
