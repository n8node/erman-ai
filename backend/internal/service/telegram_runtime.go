package service

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

const telegramHealthInterval = 30 * time.Second
const telegramQueueInterval = 5 * time.Second

// TelegramService sends notifications and runs a background health supervisor
// that starts with the backend process (container restart / server reboot).
type TelegramService struct {
	settings    *TelegramSettingsService
	threads     *repository.TelegramSupportThreadRepository
	userState   *repository.TelegramUserStateRepository
	urgentSends *repository.TelegramUrgentSendRepository
	queue       *repository.TelegramNotificationQueueRepository
	mail        *MailService
	max         *MaxService
	assets      *TelegramAssets
	client      *http.Client
	logger      *slog.Logger

	mu                sync.RWMutex
	runtime           model.TelegramBotRuntimeStatus
	supervisorRunning bool
	started           bool
	pollRunning       bool
	updateOffset      int64
	stopCh            chan struct{}
	pollStopCh        chan struct{}
	triggerCh         chan struct{}
	queueTriggerCh    chan struct{}
}

func NewTelegramService(
	settings *TelegramSettingsService,
	threads *repository.TelegramSupportThreadRepository,
	userState *repository.TelegramUserStateRepository,
	urgentSends *repository.TelegramUrgentSendRepository,
	queue *repository.TelegramNotificationQueueRepository,
	mail *MailService,
	max *MaxService,
	assets *TelegramAssets,
	logger *slog.Logger,
) *TelegramService {
	if logger == nil {
		logger = slog.Default()
	}
	return &TelegramService{
		settings:    settings,
		threads:     threads,
		userState:   userState,
		urgentSends: urgentSends,
		queue:       queue,
		mail:        mail,
		max:         max,
		assets:      assets,
		client:      &http.Client{Timeout: 60 * time.Second},
		logger:      logger,
		runtime: model.TelegramBotRuntimeStatus{
			Status:  model.TelegramBotStatusDisabled,
			Message: "Супервизор не запущен",
		},
		triggerCh:      make(chan struct{}, 1),
		queueTriggerCh: make(chan struct{}, 1),
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
	go s.queueLoop()
	s.triggerQueueDelivery()
	s.ensurePollingFromSettings()
}

func (s *TelegramService) ensurePollingFromSettings() {
	cfg, err := s.settings.GetEffective(context.Background())
	if err != nil {
		return
	}
	if !s.botNeedsPolling(cfg) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	st := s.checkHealth(ctx)
	s.setRuntime(st)
	if s.shouldRunPolling(cfg, st) {
		s.syncPolling(true, cfg)
	}
}

// Stop shuts down the supervisor gracefully.
func (s *TelegramService) Stop() {
	s.stopPolling()
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

// RunHealthCheck performs a fresh Telegram bot health check (used by ops monitors).
func (s *TelegramService) RunHealthCheck(ctx context.Context) model.TelegramBotRuntimeStatus {
	return s.checkHealth(ctx)
}

// SendDirectAdminMessage sends an ops alert directly to the admin chat, bypassing the queue.
func (s *TelegramService) SendDirectAdminMessage(ctx context.Context, text string) error {
	cfg, err := s.settings.GetEffective(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled {
		return ErrTelegramDisabled
	}
	return s.send(ctx, cfg, text)
}

func (s *TelegramService) setRuntime(st model.TelegramBotRuntimeStatus) {
	s.mu.Lock()
	st.SupervisorRunning = s.supervisorRunning
	st.PollingRunning = s.pollRunning
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

	cfg, _ := s.settings.GetEffective(context.Background())
	if s.botNeedsPolling(cfg) {
		if s.shouldRunPolling(cfg, st) {
			s.syncPolling(true, cfg)
		}
	} else {
		s.stopPolling()
	}

	if st.Status == model.TelegramBotStatusOnline {
		s.logger.Info("telegram bot online", "username", st.BotUsername, "polling", s.pollRunning)
	} else if st.Status == model.TelegramBotStatusOffline || st.Status == model.TelegramBotStatusMisconfigured {
		s.logger.Warn("telegram bot unhealthy", "status", st.Status, "error", st.LastError)
	}
}

func (s *TelegramService) shouldRunPolling(cfg model.TelegramSettings, st model.TelegramBotRuntimeStatus) bool {
	if !s.botNeedsPolling(cfg) {
		return false
	}
	return st.Status == model.TelegramBotStatusOnline
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

	if !cfg.Enabled && !cfg.StartEnabled && !cfg.SupportEnabled {
		return model.TelegramBotRuntimeStatus{
			Status:      model.TelegramBotStatusDisabled,
			Message:     "Бот отключён",
			LastCheckAt: now,
		}
	}

	token := strings.TrimSpace(cfg.BotToken)
	chatID := strings.TrimSpace(cfg.ChatID)
	if token == "" {
		return model.TelegramBotRuntimeStatus{
			Status:      model.TelegramBotStatusMisconfigured,
			Message:     "Не задан токен бота",
			LastCheckAt: now,
		}
	}
	me, err := s.telegramGetMe(ctx, token)
	if err != nil {
		return model.TelegramBotRuntimeStatus{
			Status:      model.TelegramBotStatusOffline,
			Message:     "Бот недоступен",
			LastError:   err.Error(),
			LastCheckAt: now,
		}
	}

	var notifyWarn string
	if cfg.SupportEnabled {
		forumID := strings.TrimSpace(cfg.SupportForumChatID)
		if forumID == "" {
			return model.TelegramBotRuntimeStatus{
				Status:      model.TelegramBotStatusMisconfigured,
				Message:     "Не задан ID forum-группы для консультаций",
				LastCheckAt: now,
			}
		}
		if err := s.telegramGetChat(ctx, token, forumID); err != nil {
			return model.TelegramBotRuntimeStatus{
				Status:      model.TelegramBotStatusOffline,
				Message:     "Бот не видит forum-группу",
				BotUsername: me.Username,
				LastError:   err.Error(),
				LastCheckAt: now,
			}
		}
	}

	if cfg.Enabled {
		if chatID == "" {
			if !cfg.StartEnabled && !cfg.SupportEnabled {
				return model.TelegramBotRuntimeStatus{
					Status:      model.TelegramBotStatusMisconfigured,
					Message:     "Не задан ID чата для уведомлений",
					LastCheckAt: now,
				}
			}
			notifyWarn = "ID чата для уведомлений не задан"
		} else if err := s.telegramGetChat(ctx, token, chatID); err != nil {
			if !cfg.StartEnabled && !cfg.SupportEnabled {
				return model.TelegramBotRuntimeStatus{
					Status:      model.TelegramBotStatusOffline,
					Message:     "Бот не видит указанный чат",
					BotUsername: me.Username,
					LastError:   err.Error(),
					LastCheckAt: now,
				}
			}
			notifyWarn = "уведомления в чат недоступны: " + err.Error()
		}
	}

	msg := "Бот работает"
	if me.Username != "" {
		msg = "Бот работает (@" + me.Username + ")"
	}
	if notifyWarn != "" {
		msg += " — " + notifyWarn
	}
	return model.TelegramBotRuntimeStatus{
		Status:      model.TelegramBotStatusOnline,
		Message:     msg,
		BotUsername: me.Username,
		LastError:   notifyWarn,
		LastCheckAt: now,
	}
}
