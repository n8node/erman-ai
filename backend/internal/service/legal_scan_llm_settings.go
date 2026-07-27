package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidLegalScanLLMSettings = errors.New("invalid legal scan llm settings")

type LegalScanLLMSettingsService struct {
	repo     *repository.LegalScanLLMSettingsRepository
	strategy *StrategyLLMSettingsService
	cfg      *config.Config
}

func NewLegalScanLLMSettingsService(
	repo *repository.LegalScanLLMSettingsRepository,
	strategy *StrategyLLMSettingsService,
	cfg *config.Config,
) *LegalScanLLMSettingsService {
	return &LegalScanLLMSettingsService{repo: repo, strategy: strategy, cfg: cfg}
}

func (s *LegalScanLLMSettingsService) GetStored(ctx context.Context) (*model.LegalScanLLMSettingsRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := model.DefaultLegalScanLLMStoredConfig()
			s.applyPromptDefault(&def.LegalScanLLMSettings)
			return &model.LegalScanLLMSettingsRecord{Config: def}, nil
		}
		return nil, err
	}
	s.applyPromptDefault(&rec.Config.LegalScanLLMSettings)
	return rec, nil
}

func (s *LegalScanLLMSettingsService) GetAdminView(ctx context.Context) (*model.LegalScanLLMAdminView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	strategyRec, err := s.strategy.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildAdminView(rec, strategyRec), nil
}

func (s *LegalScanLLMSettingsService) Update(ctx context.Context, req model.LegalScanLLMAdminUpdateRequest) (*model.LegalScanLLMAdminView, error) {
	if err := validateLegalScanLLMSettings(req.Settings); err != nil {
		return nil, err
	}
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	cfg := rec.Config
	cfg.LegalScanLLMSettings = req.Settings
	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	s.applyPromptDefault(&updated.Config.LegalScanLLMSettings)
	strategyRec, err := s.strategy.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildAdminView(updated, strategyRec), nil
}

func (s *LegalScanLLMSettingsService) buildAdminView(rec *model.LegalScanLLMSettingsRecord, strategyRec *model.StrategyLLMSettingsRecord) *model.LegalScanLLMAdminView {
	stored := strategyRec.Config
	orKey := firstNonEmpty(stored.OpenRouterAPIKey, s.cfg.OpenRouterAPIKey)
	dsKey := firstNonEmpty(stored.DeepSeekAPIKey, s.cfg.DeepSeekAPIKey)
	yandexKey := firstNonEmpty(stored.YandexAPIKey, s.cfg.YandexAPIKey)
	yandexFolder := firstNonEmpty(stored.YandexFolderID, s.cfg.YandexFolderID)
	providers := []model.LLMProviderStatus{
		{
			ID:           model.LLMProviderOpenRouter,
			Configured:   orKey != "",
			KeyHint:      model.MaskAPIKey(orKey),
			DefaultModel: rec.Config.OpenRouterModel,
			Models:       stored.OpenRouterModels,
		},
		{
			ID:           model.LLMProviderDeepSeek,
			Configured:   dsKey != "",
			KeyHint:      model.MaskAPIKey(dsKey),
			DefaultModel: rec.Config.DeepSeekModel,
			Models:       stored.DeepSeekModels,
		},
		{
			ID:           model.LLMProviderYandex,
			Configured:   yandexKey != "" && yandexFolder != "",
			KeyHint:      model.MaskAPIKey(yandexKey),
			FolderHint:   model.MaskFolderID(yandexFolder),
			DefaultModel: YandexModelURI(yandexFolder, s.cfg.YandexModelFast),
			Models:       stored.YandexModels,
		},
	}
	return &model.LegalScanLLMAdminView{
		Settings:            rec.Config.LegalScanLLMSettings,
		Providers:           providers,
		DefaultSystemPrompt: prompts.DefaultLegalScanSystemPrompt,
		UpdatedAt:           rec.UpdatedAt,
	}
}

func (s *LegalScanLLMSettingsService) applyPromptDefault(settings *model.LegalScanLLMSettings) {
	if settings.Provider == "" {
		settings.Provider = model.LLMProviderYandex
	}
	if settings.OpenRouterModel == "" {
		settings.OpenRouterModel = "google/gemini-flash-1.5-8b"
	}
	if settings.DeepSeekModel == "" {
		settings.DeepSeekModel = "deepseek-chat"
	}
	if settings.YandexModel == "" {
		settings.YandexModel = s.cfg.YandexModelFast
	}
	if settings.MaxTokens == 0 {
		settings.MaxTokens = 4096
	}
}

func validateLegalScanLLMSettings(s model.LegalScanLLMSettings) error {
	switch s.Provider {
	case model.LLMProviderOpenRouter, model.LLMProviderDeepSeek, model.LLMProviderYandex:
	default:
		return fmt.Errorf("%w: invalid provider", ErrInvalidLegalScanLLMSettings)
	}
	if strings.TrimSpace(s.OpenRouterModel) == "" || strings.TrimSpace(s.DeepSeekModel) == "" || strings.TrimSpace(s.YandexModel) == "" {
		return fmt.Errorf("%w: model required", ErrInvalidLegalScanLLMSettings)
	}
	if s.Temperature < 0 || s.Temperature > 2 {
		return fmt.Errorf("%w: temperature out of range", ErrInvalidLegalScanLLMSettings)
	}
	if s.MaxTokens < 256 || s.MaxTokens > 32000 {
		return fmt.Errorf("%w: max_tokens out of range", ErrInvalidLegalScanLLMSettings)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

type LegalRiskService struct {
	repo *repository.LegalRiskRepository
}

func NewLegalRiskService(repo *repository.LegalRiskRepository) *LegalRiskService {
	return &LegalRiskService{repo: repo}
}

func (s *LegalRiskService) List(ctx context.Context) ([]model.LegalRisk, error) {
	return s.repo.ListAll(ctx)
}

func (s *LegalRiskService) Update(ctx context.Context, riskID string, req model.LegalRiskUpdateRequest) (*model.LegalRisk, error) {
	if err := validateLegalRiskUpdate(req); err != nil {
		return nil, err
	}
	item, err := s.repo.Update(ctx, riskID, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrLegalRiskNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *LegalRiskService) Create(ctx context.Context, riskID string, req model.LegalRiskUpdateRequest) (*model.LegalRisk, error) {
	riskID = strings.TrimSpace(riskID)
	if riskID == "" {
		return nil, fmt.Errorf("%w: risk_id required", ErrInvalidInput)
	}
	if err := validateLegalRiskUpdate(req); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, riskID, req)
}

func validateLegalRiskUpdate(req model.LegalRiskUpdateRequest) error {
	if strings.TrimSpace(req.TitleRU) == "" || strings.TrimSpace(req.Article) == "" {
		return fmt.Errorf("%w: title and article required", ErrInvalidInput)
	}
	if req.Severity != "high" && req.Severity != "medium" && req.Severity != "low" {
		return fmt.Errorf("%w: invalid severity", ErrInvalidInput)
	}
	return nil
}
