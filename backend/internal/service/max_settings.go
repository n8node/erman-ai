package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidMaxSettings = errors.New("invalid max settings")

type MaxSettingsService struct {
	repo    *repository.MaxSettingsRepository
	runtime func() model.MaxBotRuntimeStatus
}

func NewMaxSettingsService(repo *repository.MaxSettingsRepository) *MaxSettingsService {
	return &MaxSettingsService{repo: repo}
}

func (s *MaxSettingsService) BindRuntimeStatus(fn func() model.MaxBotRuntimeStatus) {
	s.runtime = fn
}

func (s *MaxSettingsService) currentRuntime() model.MaxBotRuntimeStatus {
	if s.runtime != nil {
		return s.runtime()
	}
	return model.MaxBotRuntimeStatus{Status: model.MaxBotStatusMisconfigured, Message: "Статус недоступен"}
}

func (s *MaxSettingsService) GetStored(ctx context.Context) (*model.MaxSettingsRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := model.DefaultMaxSettings()
			return &model.MaxSettingsRecord{Config: def}, nil
		}
		return nil, err
	}
	return rec, nil
}

func (s *MaxSettingsService) GetEffective(ctx context.Context) (model.MaxSettings, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return model.MaxSettings{}, err
	}
	return rec.Config, nil
}

func (s *MaxSettingsService) GetAdminView(ctx context.Context) (*model.MaxAdminView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	pub := rec.Config
	pub.BotToken = ""
	return &model.MaxAdminView{
		Settings:     pub,
		BotTokenSet:  strings.TrimSpace(rec.Config.BotToken) != "",
		BotTokenHint: maskSecret(rec.Config.BotToken),
		UpdatedAt:    rec.UpdatedAt,
		Runtime:      s.currentRuntime(),
	}, nil
}

func (s *MaxSettingsService) Update(ctx context.Context, req model.MaxAdminUpdateRequest) (*model.MaxAdminView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	cfg := req.Settings
	if strings.TrimSpace(req.BotToken) != "" {
		cfg.BotToken = strings.TrimSpace(req.BotToken)
	} else {
		cfg.BotToken = rec.Config.BotToken
	}
	if err := validateMaxSettings(cfg); err != nil {
		return nil, err
	}
	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	pub := updated.Config
	pub.BotToken = ""
	return &model.MaxAdminView{
		Settings:     pub,
		BotTokenSet:  strings.TrimSpace(updated.Config.BotToken) != "",
		BotTokenHint: maskSecret(updated.Config.BotToken),
		UpdatedAt:    updated.UpdatedAt,
		Runtime:      s.currentRuntime(),
	}, nil
}

func validateMaxSettings(cfg model.MaxSettings) error {
	if !cfg.Enabled {
		return nil
	}
	if strings.TrimSpace(cfg.BotToken) == "" {
		return fmt.Errorf("%w: bot token required when MAX enabled", ErrInvalidMaxSettings)
	}
	hasUser := strings.TrimSpace(cfg.NotifyUserID) != ""
	hasChat := strings.TrimSpace(cfg.NotifyChatID) != ""
	if cfg.UrgentAlertsEnabled && !hasUser && !hasChat {
		return fmt.Errorf("%w: notify_user_id or notify_chat_id required for urgent alerts", ErrInvalidMaxSettings)
	}
	return nil
}
