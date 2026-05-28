package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidTelegramSettings = errors.New("invalid telegram settings")

type TelegramSettingsService struct {
	repo     *repository.TelegramSettingsRepository
	runtime  func() model.TelegramBotRuntimeStatus
}

func NewTelegramSettingsService(repo *repository.TelegramSettingsRepository) *TelegramSettingsService {
	return &TelegramSettingsService{repo: repo}
}

func (s *TelegramSettingsService) BindRuntimeStatus(fn func() model.TelegramBotRuntimeStatus) {
	s.runtime = fn
}

func (s *TelegramSettingsService) currentRuntime() model.TelegramBotRuntimeStatus {
	if s.runtime != nil {
		return s.runtime()
	}
	return model.TelegramBotRuntimeStatus{
		Status:  model.TelegramBotStatusStarting,
		Message: "Статус недоступен",
	}
}

func (s *TelegramSettingsService) GetStored(ctx context.Context) (*model.TelegramSettingsRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := model.DefaultTelegramSettings()
			return &model.TelegramSettingsRecord{Config: def}, nil
		}
		return nil, err
	}
	return rec, nil
}

func (s *TelegramSettingsService) GetEffective(ctx context.Context) (model.TelegramSettings, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return model.TelegramSettings{}, err
	}
	return rec.Config, nil
}

func (s *TelegramSettingsService) GetAdminView(ctx context.Context) (*model.TelegramAdminView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	return buildTelegramAdminView(rec, s.currentRuntime()), nil
}

func (s *TelegramSettingsService) Update(ctx context.Context, req model.TelegramAdminUpdateRequest) (*model.TelegramAdminView, error) {
	if err := validateTelegramSettings(req.Settings); err != nil {
		return nil, err
	}

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

	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return buildTelegramAdminView(updated, s.currentRuntime()), nil
}

func buildTelegramAdminView(rec *model.TelegramSettingsRecord, runtime model.TelegramBotRuntimeStatus) *model.TelegramAdminView {
	pub := rec.Config
	pub.BotToken = ""
	return &model.TelegramAdminView{
		Settings:     pub,
		BotTokenSet:  strings.TrimSpace(rec.Config.BotToken) != "",
		BotTokenHint: maskSecret(rec.Config.BotToken),
		UpdatedAt:    rec.UpdatedAt,
		Runtime:      runtime,
	}
}

func validateTelegramSettings(cfg model.TelegramSettings) error {
	if cfg.Enabled {
		if strings.TrimSpace(cfg.ChatID) == "" {
			return fmt.Errorf("%w: chat_id required when enabled", ErrInvalidTelegramSettings)
		}
	}
	return nil
}
