package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidYandexMetrikaSettings = errors.New("invalid yandex metrika settings")

type YandexMetrikaSettingsService struct {
	repo *repository.YandexMetrikaSettingsRepository
}

func NewYandexMetrikaSettingsService(repo *repository.YandexMetrikaSettingsRepository) *YandexMetrikaSettingsService {
	return &YandexMetrikaSettingsService{repo: repo}
}

func (s *YandexMetrikaSettingsService) GetStored(ctx context.Context) (*model.YandexMetrikaSettingsRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := model.DefaultYandexMetrikaSettings()
			return &model.YandexMetrikaSettingsRecord{Config: def}, nil
		}
		return nil, err
	}
	return rec, nil
}

func (s *YandexMetrikaSettingsService) GetAdminView(ctx context.Context) (*model.YandexMetrikaAdminView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	return &model.YandexMetrikaAdminView{
		Settings:  rec.Config,
		UpdatedAt: rec.UpdatedAt,
	}, nil
}

func (s *YandexMetrikaSettingsService) GetPublicView(ctx context.Context) (*model.YandexMetrikaPublicView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	cfg := rec.Config
	if !cfg.Enabled || strings.TrimSpace(cfg.CounterCode) == "" {
		return &model.YandexMetrikaPublicView{Enabled: false, CounterCode: ""}, nil
	}
	return &model.YandexMetrikaPublicView{
		Enabled:     true,
		CounterCode: cfg.CounterCode,
	}, nil
}

func (s *YandexMetrikaSettingsService) Update(ctx context.Context, req model.YandexMetrikaAdminUpdateRequest) (*model.YandexMetrikaAdminView, error) {
	cfg := req.Settings
	cfg.CounterCode = strings.TrimSpace(cfg.CounterCode)
	if err := validateYandexMetrikaSettings(cfg); err != nil {
		return nil, err
	}
	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &model.YandexMetrikaAdminView{
		Settings:  updated.Config,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}

func validateYandexMetrikaSettings(cfg model.YandexMetrikaSettings) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.CounterCode == "" {
		return fmt.Errorf("%w: counter code required when enabled", ErrInvalidYandexMetrikaSettings)
	}
	lower := strings.ToLower(cfg.CounterCode)
	if !strings.Contains(lower, "yandex") && !strings.Contains(lower, "mc.yandex.ru") {
		return fmt.Errorf("%w: counter code must contain Yandex Metrika snippet", ErrInvalidYandexMetrikaSettings)
	}
	return nil
}
