package service

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

const telegramHealthInterval = 30 * time.Second

// TelegramService sends notifications and runs a background health supervisor
// that starts with the backend process (container restart / server reboot).
type TelegramService struct {
	settings *TelegramSettingsService
	client   *http.Client
	logger   *slog.Logger

	mu                 sync.RWMutex
	runtime            model.TelegramBotRuntimeStatus
	supervisorRunning  bool
	started            bool
	stopCh             chan struct{}
	triggerCh          chan struct{}
}

func NewTelegramService(settings *TelegramSettingsService, logger *slog.Logger) *TelegramService {
	if logger == nil {
		logger = slog.Default()
	}
	return &TelegramService{
		settings: settings,
		client:   &http.Client{Timeout: 15 * time.Second},
		logger:   logger,
		runtime: model.TelegramBotRuntimeStatus{
			Status:  model.TelegramBotStatusDisabled,
			Message: "Супервизор не запущен",
		},
		triggerCh: make(chan struct{}, 1),
	}
}

// Start launches the background supervisor (idempotent).
func (s *TelegramService) Start() {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	s.stopCh = make(chan struct{})
	s.supervisorRunning = true
	s.mu.Unlock()

	s.logger.Info("telegram bot supervisor starting")
	go s.supervisorLoop()
}

// Stop shuts down the supervisor gracefully.
func (s *TelegramService) Stop() {
	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return
	}
	close(s.stopCh)
	s.started = false
	s.supervisorRunning = false
	s.mu.Unlock()
	s.logger.Info("telegram bot supervisor stopped")
}

// Restart triggers an immediate health check and ensures the supervisor is running.
func (s *TelegramService) Restart(ctx context.Context) model.TelegramBotRuntimeStatus {
	s.Start()
	s.triggerHealthCheck()
	if ctx != nil {
		select {
		case <-ctx.Done():
		case <-time.After(12 * time.Second):
		}
	}
	return s.GetRuntimeStatus()
}

func (s *TelegramService) GetRuntimeStatus() model.TelegramBotRuntimeStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.runtime
}

func (s *TelegramService) setRuntime(st model.TelegramBotRuntimeStatus) {
	s.mu.Lock()
	st.SupervisorRunning = s.supervisorRunning
	s.runtime = st
	s.mu.Unlock()
}

func (s *TelegramService) triggerHealthCheck() {
	select {
	case s.triggerCh <- struct{}{}:
	default:
	}
}

func (s *TelegramService) supervisorLoop() {
	ticker := time.NewTicker(telegramHealthInterval)
	defer ticker.Stop()

	s.runHealthCheck()

	for {
		select {
		case <-s.stopCh:
			s.mu.Lock()
			s.supervisorRunning = false
			s.mu.Unlock()
			return
		case <-s.triggerCh:
			s.runHealthCheck()
		case <-ticker.C:
			s.runHealthCheck()
		}
	}
}

func (s *TelegramService) runHealthCheck() {
	s.setRuntime(model.TelegramBotRuntimeStatus{
		Status:  model.TelegramBotStatusStarting,
		Message: "Проверка подключения…",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	st := s.checkHealth(ctx)
	s.setRuntime(st)

	if st.Status == model.TelegramBotStatusOnline {
		s.logger.Info("telegram bot online", "username", st.BotUsername)
	} else if st.Status == model.TelegramBotStatusOffline || st.Status == model.TelegramBotStatusMisconfigured {
		s.logger.Warn("telegram bot unhealthy", "status", st.Status, "error", st.LastError)
	}
}

func (s *TelegramService) checkHealth(ctx context.Context) model.TelegramBotRuntimeStatus {
	now := time.Now()
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return model.TelegramBotRuntimeStatus{
			Status:      model.TelegramBotStatusOffline,
			Message:     "Ошибка загрузки настроек",
			LastError:   err.Error(),
			LastCheckAt: now,
		}
	}

	if !cfg.Enabled {
		return model.TelegramBotRuntimeStatus{
			Status:      model.TelegramBotStatusDisabled,
			Message:     "Уведомления отключены",
			LastCheckAt: now,
		}
	}

	token := strings.TrimSpace(cfg.BotToken)
	chatID := strings.TrimSpace(cfg.ChatID)
	if token == "" || chatID == "" {
		msg := "Укажите токен бота и ID чата"
		if token == "" && chatID == "" {
			msg = "Не заданы токен бота и ID чата"
		} else if token == "" {
			msg = "Не задан токен бота"
		} else {
			msg = "Не задан ID чата"
		}
		return model.TelegramBotRuntimeStatus{
			Status:      model.TelegramBotStatusMisconfigured,
			Message:     msg,
			LastCheckAt: now,
		}
	}

	_, _ = s.telegramAPI(ctx, token, "deleteWebhook", map[string]any{"drop_pending_updates": true})

	me, err := s.telegramGetMe(ctx, token)
	if err != nil {
		return model.TelegramBotRuntimeStatus{
			Status:      model.TelegramBotStatusOffline,
			Message:     "Бот недоступен",
			LastError:   err.Error(),
			LastCheckAt: now,
		}
	}

	if err := s.telegramGetChat(ctx, token, chatID); err != nil {
		return model.TelegramBotRuntimeStatus{
			Status:      model.TelegramBotStatusOffline,
			Message:     "Бот не видит указанный чат",
			BotUsername: me.Username,
			LastError:   err.Error(),
			LastCheckAt: now,
		}
	}

	msg := "Бот работает"
	if me.Username != "" {
		msg = "Бот работает (@" + me.Username + ")"
	}
	return model.TelegramBotRuntimeStatus{
		Status:      model.TelegramBotStatusOnline,
		Message:     msg,
		BotUsername: me.Username,
		LastCheckAt: now,
	}
}
