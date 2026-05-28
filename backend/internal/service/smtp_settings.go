package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidSMTPSettings = errors.New("invalid smtp settings")

const (
	yandexPresetHost = "smtp.yandex.ru"
	yandexPresetPort = 465
)

type SMTPSettingsService struct {
	repo *repository.SMTPSettingsRepository
}

func NewSMTPSettingsService(repo *repository.SMTPSettingsRepository) *SMTPSettingsService {
	return &SMTPSettingsService{repo: repo}
}

func (s *SMTPSettingsService) GetStored(ctx context.Context) (*model.SMTPSettingsRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := model.DefaultSMTPSettings()
			return &model.SMTPSettingsRecord{Config: def}, nil
		}
		return nil, err
	}
	return rec, nil
}

func (s *SMTPSettingsService) GetEffective(ctx context.Context) (model.SMTPSettings, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return model.SMTPSettings{}, err
	}
	cfg := rec.Config
	if !cfg.Auth {
		cfg.Username = ""
		cfg.Password = ""
	}
	return cfg, nil
}

func (s *SMTPSettingsService) GetAdminView(ctx context.Context) (*model.SMTPAdminView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildAdminView(rec), nil
}

func (s *SMTPSettingsService) Update(ctx context.Context, req model.SMTPAdminUpdateRequest) (*model.SMTPAdminView, error) {
	if err := validateSMTPSettings(req.Settings); err != nil {
		return nil, err
	}

	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}

	cfg := req.Settings
	if strings.TrimSpace(req.Password) != "" {
		cfg.Password = strings.TrimSpace(req.Password)
	} else {
		cfg.Password = rec.Config.Password
	}

	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return s.buildAdminView(updated), nil
}

func (s *SMTPSettingsService) ApplyYandexPreset(settings model.SMTPSettings) model.SMTPSettings {
	settings.Host = yandexPresetHost
	settings.Port = yandexPresetPort
	settings.Encryption = model.SMTPEncryptionSSL
	settings.AutoTLS = true
	settings.Auth = true
	return settings
}

func (s *SMTPSettingsService) buildAdminView(rec *model.SMTPSettingsRecord) *model.SMTPAdminView {
	pub := rec.Config
	pub.Password = ""
	return &model.SMTPAdminView{
		Settings:         pub,
		PasswordSet:      strings.TrimSpace(rec.Config.Password) != "",
		PasswordHint:     maskSecret(rec.Config.Password),
		UpdatedAt:        rec.UpdatedAt,
		YandexPresetHost: yandexPresetHost,
		YandexPresetPort: yandexPresetPort,
	}
}

func validateSMTPSettings(cfg model.SMTPSettings) error {
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("%w: invalid port", ErrInvalidSMTPSettings)
	}
	switch cfg.Encryption {
	case model.SMTPEncryptionNone, model.SMTPEncryptionSSL, model.SMTPEncryptionTLS:
	default:
		return fmt.Errorf("%w: invalid encryption", ErrInvalidSMTPSettings)
	}
	if cfg.Enabled {
		if strings.TrimSpace(cfg.Host) == "" {
			return fmt.Errorf("%w: host required when enabled", ErrInvalidSMTPSettings)
		}
		if cfg.Auth && strings.TrimSpace(cfg.Username) == "" {
			return fmt.Errorf("%w: username required when auth enabled", ErrInvalidSMTPSettings)
		}
	}
	return nil
}

func maskSecret(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}
	if len(v) <= 4 {
		return "••••"
	}
	return v[:2] + "••••" + v[len(v)-2:]
}
