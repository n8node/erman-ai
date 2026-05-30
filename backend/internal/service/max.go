package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

var (
	ErrMaxNotConfigured = errors.New("max bot not configured")
	ErrMaxDisabled      = errors.New("max integration disabled")
)

type MaxService struct {
	settings *MaxSettingsService
	client   *http.Client
	logger   *slog.Logger
}

func NewMaxService(settings *MaxSettingsService, logger *slog.Logger) *MaxService {
	if logger == nil {
		logger = slog.Default()
	}
	return &MaxService{
		settings: settings,
		client:   &http.Client{Timeout: 20 * time.Second},
		logger:   logger,
	}
}

func (s *MaxService) CheckHealth(ctx context.Context) model.MaxBotRuntimeStatus {
	now := time.Now()
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return model.MaxBotRuntimeStatus{
			Status:      model.MaxBotStatusOffline,
			Message:     "Не удалось загрузить настройки",
			LastError:   err.Error(),
			LastCheckAt: now,
		}
	}
	if !cfg.Enabled {
		return model.MaxBotRuntimeStatus{
			Status:      model.MaxBotStatusDisabled,
			Message:     "Интеграция MAX отключена",
			LastCheckAt: now,
		}
	}
	token := strings.TrimSpace(cfg.BotToken)
	if token == "" {
		return model.MaxBotRuntimeStatus{
			Status:      model.MaxBotStatusMisconfigured,
			Message:     "Укажите токен бота MAX",
			LastCheckAt: now,
		}
	}

	me, err := s.maxGetMe(ctx, token)
	if err != nil {
		return model.MaxBotRuntimeStatus{
			Status:      model.MaxBotStatusOffline,
			Message:     "Не удалось подключиться к MAX API",
			LastError:   err.Error(),
			LastCheckAt: now,
		}
	}

	msg := "Бот MAX подключён"
	if !cfg.UrgentAlertsEnabled {
		msg += " (срочные алерты выключены)"
	}
	sendCfg, err := s.buildSendConfig(cfg)
	if err != nil {
		msg = "Бот подключён, но получатель срочных уведомлений не задан"
		return model.MaxBotRuntimeStatus{
			Status:      model.MaxBotStatusMisconfigured,
			Message:     msg,
			BotUsername: pickMaxUsername(me, cfg),
			BotUserID:   me.UserID,
			LastError:   err.Error(),
			LastCheckAt: now,
		}
	}
	_ = sendCfg

	return model.MaxBotRuntimeStatus{
		Status:      model.MaxBotStatusOnline,
		Message:     msg,
		BotUsername: pickMaxUsername(me, cfg),
		BotUserID:   me.UserID,
		LastCheckAt: now,
	}
}

func pickMaxUsername(me *maxAPIUser, cfg model.MaxSettings) string {
	if me != nil && me.Username != "" {
		return me.Username
	}
	return strings.TrimSpace(cfg.BotUsername)
}

func (s *MaxService) buildSendConfig(cfg model.MaxSettings) (maxSendConfig, error) {
	token := strings.TrimSpace(cfg.BotToken)
	if token == "" {
		return maxSendConfig{}, ErrMaxNotConfigured
	}
	out := maxSendConfig{Token: token}
	if uid := strings.TrimSpace(cfg.NotifyUserID); uid != "" {
		parsed, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			return maxSendConfig{}, fmt.Errorf("invalid notify_user_id")
		}
		out.NotifyUserID = parsed
	}
	if cid := strings.TrimSpace(cfg.NotifyChatID); cid != "" {
		parsed, err := strconv.ParseInt(cid, 10, 64)
		if err != nil {
			return maxSendConfig{}, fmt.Errorf("invalid notify_chat_id")
		}
		out.NotifyChatID = parsed
	}
	if out.NotifyUserID == 0 && out.NotifyChatID == 0 {
		return maxSendConfig{}, fmt.Errorf("notify_user_id or notify_chat_id required")
	}
	return out, nil
}

func (s *MaxService) SendNotification(ctx context.Context, text string) error {
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return ErrMaxDisabled
	}
	sendCfg, err := s.buildSendConfig(cfg)
	if err != nil {
		return err
	}
	return s.maxSendMessage(ctx, sendCfg, text)
}

func (s *MaxService) SendUrgentAlert(ctx context.Context, text string) error {
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled || !cfg.UrgentAlertsEnabled {
		return ErrMaxDisabled
	}
	sendCfg, err := s.buildSendConfig(cfg)
	if err != nil {
		return err
	}
	return s.maxSendMessage(ctx, sendCfg, text)
}

func (s *MaxService) SendTest(ctx context.Context) (bool, string) {
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return false, err.Error()
	}
	if !cfg.Enabled {
		return false, "Включите интеграцию MAX в настройках"
	}
	sendCfg, err := s.buildSendConfig(cfg)
	if err != nil {
		return false, err.Error()
	}
	text := "✅ Тестовое сообщение Erman AI\nБот MAX настроен и может отправлять срочные уведомления."
	if err := s.maxSendMessage(ctx, sendCfg, text); err != nil {
		return false, err.Error()
	}
	return true, "Тестовое сообщение отправлено в MAX"
}
