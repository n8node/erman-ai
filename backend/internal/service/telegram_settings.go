package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidTelegramSettings = errors.New("invalid telegram settings")

type TelegramSettingsService struct {
	repo    *repository.TelegramSettingsRepository
	assets  *TelegramAssets
	env     *config.Config
	logger  *slog.Logger
	runtime func() model.TelegramBotRuntimeStatus
}

func NewTelegramSettingsService(
	repo *repository.TelegramSettingsRepository,
	assets *TelegramAssets,
	env *config.Config,
	logger *slog.Logger,
) *TelegramSettingsService {
	if logger == nil {
		logger = slog.Default()
	}
	return &TelegramSettingsService{repo: repo, assets: assets, env: env, logger: logger}
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

func (s *TelegramSettingsService) finalizeSettings(cfg model.TelegramSettings) model.TelegramSettings {
	s.applyTelegramProxyEnv(&cfg)
	normalizeTelegramSettingsFromStorage(&cfg)
	return cfg
}

func (s *TelegramSettingsService) envProxyConfigured() bool {
	if s.env == nil {
		return false
	}
	if len(parseTelegramProxyURLs(s.env.TelegramProxyURLs)) > 0 {
		return true
	}
	if strings.TrimSpace(s.env.TelegramProxyActiveURL) != "" {
		return true
	}
	return s.env.TelegramProxyEnabled
}

func (s *TelegramSettingsService) applyTelegramProxyEnv(cfg *model.TelegramSettings) {
	if !s.envProxyConfigured() {
		return
	}
	envURLs := parseTelegramProxyURLs(s.env.TelegramProxyURLs)
	if len(envURLs) > 0 {
		cfg.ProxyURLs = envURLs
		cfg.ProxyEnabled = true
	}
	if s.env.TelegramProxyEnabled {
		cfg.ProxyEnabled = true
	}
	if active := strings.TrimSpace(s.env.TelegramProxyActiveURL); active != "" {
		cfg.ProxyActiveURL = active
	}
	cfg.ProxyAutoFailover = s.env.TelegramProxyAutoFailover
}

// EnsureEnvProxyPersisted writes Telegram proxy settings from .env into the DB on startup.
func (s *TelegramSettingsService) EnsureEnvProxyPersisted(ctx context.Context) error {
	if s.env == nil || !s.envProxyConfigured() {
		return nil
	}

	rec, err := s.GetStored(ctx)
	if err != nil {
		return err
	}
	next := rec.Config
	s.applyTelegramProxyEnv(&next)
	normalizeTelegramSettingsFromStorage(&next)
	if telegramProxyConfigEqual(rec.Config, next) {
		return nil
	}
	if _, err := s.repo.Update(ctx, next); err != nil {
		return err
	}
	s.logger.Info("telegram proxy settings synced from environment")
	return nil
}

func (s *TelegramSettingsService) GetStored(ctx context.Context) (*model.TelegramSettingsRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := s.finalizeSettings(model.DefaultTelegramSettings())
			return &model.TelegramSettingsRecord{Config: def}, nil
		}
		return nil, err
	}
	rec.Config = s.finalizeSettings(rec.Config)
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
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}

	cfg := mergeTelegramSettingsUpdate(rec.Config, req.Settings)
	if strings.TrimSpace(req.BotToken) != "" {
		cfg.BotToken = strings.TrimSpace(req.BotToken)
	} else {
		cfg.BotToken = rec.Config.BotToken
	}
	cfg.StartImageFilename = rec.Config.StartImageFilename
	cfg.ProxyURLs = normalizeProxyURLs(cfg.ProxyURLs)
	cfg.ProxyActiveURL = strings.TrimSpace(cfg.ProxyActiveURL)
	if cfg.ProxyEnabled && cfg.ProxyActiveURL == "" && len(cfg.ProxyURLs) > 0 {
		cfg.ProxyActiveURL = cfg.ProxyURLs[0]
	}
	if !cfg.ProxyEnabled {
		cfg.ProxyActiveURL = ""
	}
	if req.ClearStartImage {
		cfg.StartImageFilename = ""
		if s.assets != nil {
			_ = s.assets.DeleteStartImage()
		}
	}

	if err := validateTelegramSettings(cfg); err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return buildTelegramAdminView(updated, s.currentRuntime()), nil
}

func (s *TelegramSettingsService) SaveStartImage(ctx context.Context, contentType string, r io.Reader) (*model.TelegramAdminView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	if s.assets == nil {
		return nil, fmt.Errorf("%w: assets storage not configured", ErrInvalidTelegramSettings)
	}
	filename, err := s.assets.SaveStartImage(contentType, r, model.TelegramPhotoMaxBytes)
	if err != nil {
		return nil, err
	}
	cfg := rec.Config
	cfg.StartImageFilename = filename
	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return buildTelegramAdminView(updated, s.currentRuntime()), nil
}

func (s *TelegramSettingsService) StartImageReader(filename string) (io.ReadCloser, string, error) {
	if s.assets == nil || !s.assets.StartImageExists(filename) {
		return nil, "", os.ErrNotExist
	}
	f, err := s.assets.OpenStartImage(filename)
	if err != nil {
		return nil, "", err
	}
	ext := strings.ToLower(filepath.Ext(filename))
	ct := "image/jpeg"
	switch ext {
	case ".png":
		ct = "image/png"
	case ".webp":
		ct = "image/webp"
	}
	return f, ct, nil
}

func buildTelegramAdminView(rec *model.TelegramSettingsRecord, runtime model.TelegramBotRuntimeStatus) *model.TelegramAdminView {
	pub := rec.Config
	pub.BotToken = ""
	pub.StartImageFilename = ""
	imageOK := rec.Config.StartImageFilename != ""
	return &model.TelegramAdminView{
		Settings:             pub,
		BotTokenSet:          strings.TrimSpace(rec.Config.BotToken) != "",
		BotTokenHint:         maskSecret(rec.Config.BotToken),
		StartImageConfigured: imageOK,
		StartTextRunes:       runeLen(rec.Config.StartText),
		StartTextLimit:       model.TelegramMessageMaxRunes,
		StartCaptionLimit:    model.TelegramCaptionMaxRunes,
		UpdatedAt:            rec.UpdatedAt,
		Runtime:              runtime,
	}
}

func validateTelegramSettings(cfg model.TelegramSettings) error {
	cfg.ProxyURLs = normalizeProxyURLs(cfg.ProxyURLs)
	cfg.ProxyActiveURL = strings.TrimSpace(cfg.ProxyActiveURL)
	if cfg.ProxyEnabled {
		if len(cfg.ProxyURLs) == 0 {
			return fmt.Errorf("%w: add at least one proxy url", ErrInvalidTelegramSettings)
		}
		for _, raw := range cfg.ProxyURLs {
			u, err := url.Parse(raw)
			if err != nil || u.Scheme == "" || u.Host == "" {
				return fmt.Errorf("%w: invalid proxy url %q", ErrInvalidTelegramSettings, raw)
			}
			if strings.ToLower(u.Scheme) != "http" {
				return fmt.Errorf("%w: proxy url %q: only http:// proxies are supported", ErrInvalidTelegramSettings, raw)
			}
		}
		if cfg.ProxyActiveURL == "" {
			cfg.ProxyActiveURL = cfg.ProxyURLs[0]
		}
		if !containsProxyURL(cfg.ProxyURLs, cfg.ProxyActiveURL) {
			return fmt.Errorf("%w: active proxy must be one of proxy_urls", ErrInvalidTelegramSettings)
		}
	} else {
		cfg.ProxyActiveURL = ""
	}

	if cfg.Enabled {
		if strings.TrimSpace(cfg.ChatID) == "" {
			return fmt.Errorf("%w: chat_id required when notifications enabled", ErrInvalidTelegramSettings)
		}
	}
	if cfg.SupportEnabled {
		if strings.TrimSpace(cfg.BotToken) == "" {
			return fmt.Errorf("%w: bot token required for consultations", ErrInvalidTelegramSettings)
		}
		if strings.TrimSpace(cfg.SupportForumChatID) == "" {
			return fmt.Errorf("%w: support_forum_chat_id required when consultations enabled", ErrInvalidTelegramSettings)
		}
	}
	if cfg.UrgentEnabled {
		if strings.TrimSpace(cfg.BotToken) == "" {
			return fmt.Errorf("%w: bot token required for urgent contact", ErrInvalidTelegramSettings)
		}
		if strings.TrimSpace(cfg.ChatID) == "" {
			return fmt.Errorf("%w: chat_id required for urgent telegram alerts", ErrInvalidTelegramSettings)
		}
		if strings.TrimSpace(cfg.UrgentEmail) == "" {
			return fmt.Errorf("%w: urgent_email required", ErrInvalidTelegramSettings)
		}
	}
	if cfg.StartEnabled {
		if strings.TrimSpace(cfg.BotToken) == "" {
			return fmt.Errorf("%w: bot token required for /start handler", ErrInvalidTelegramSettings)
		}
		hasText := strings.TrimSpace(cfg.StartText) != ""
		hasImage := strings.TrimSpace(cfg.StartImageFilename) != ""
		if !hasText && !hasImage {
			return fmt.Errorf("%w: add welcome text or image for /start", ErrInvalidTelegramSettings)
		}
		if runeLen(cfg.StartText) > model.TelegramMessageMaxRunes {
			return fmt.Errorf("%w: start text exceeds %d characters", ErrInvalidTelegramSettings, model.TelegramMessageMaxRunes)
		}
	}
	return nil
}
