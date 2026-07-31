package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

const (
	notificationHealthCheckInterval = time.Hour
	notificationHealthStartupDelay  = 2 * time.Minute
)

type notificationHealthAlertPlan struct {
	SendEmail    bool
	SendTelegram bool
	SendMax      bool
}

// NotificationHealthMonitor runs hourly self-diagnostics for Telegram and MAX bots.
// When a channel fails, it alerts through the other channels and email using a failover matrix.
type NotificationHealthMonitor struct {
	telegram *TelegramService
	max      *MaxService
	mail     *MailService
	logger   *slog.Logger

	mu      sync.Mutex
	started bool
	stopCh  chan struct{}
}

func NewNotificationHealthMonitor(
	telegram *TelegramService,
	max *MaxService,
	mail *MailService,
	logger *slog.Logger,
) *NotificationHealthMonitor {
	if logger == nil {
		logger = slog.Default()
	}
	return &NotificationHealthMonitor{
		telegram: telegram,
		max:      max,
		mail:     mail,
		logger:   logger,
	}
}

func (m *NotificationHealthMonitor) Start() {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return
	}
	m.started = true
	m.stopCh = make(chan struct{})
	m.mu.Unlock()

	m.logger.Info("notification health monitor starting")
	go m.loop()
}

func (m *NotificationHealthMonitor) Stop() {
	m.mu.Lock()
	if !m.started {
		m.mu.Unlock()
		return
	}
	close(m.stopCh)
	m.started = false
	m.mu.Unlock()
	m.logger.Info("notification health monitor stopped")
}

func (m *NotificationHealthMonitor) loop() {
	timer := time.NewTimer(notificationHealthStartupDelay)
	defer timer.Stop()

	select {
	case <-m.stopCh:
		return
	case <-timer.C:
		m.runCheck()
	}

	ticker := time.NewTicker(notificationHealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.runCheck()
		}
	}
}

func (m *NotificationHealthMonitor) runCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	tgStatus, tgMonitored := m.checkTelegram(ctx)
	maxStatus, maxMonitored := m.checkMax(ctx)

	tgDown := tgMonitored && isTelegramUnhealthy(tgStatus)
	maxDown := maxMonitored && isMaxUnhealthy(maxStatus)
	plan := planNotificationHealthAlerts(tgMonitored, tgDown, maxMonitored, maxDown)
	if !plan.SendEmail && !plan.SendTelegram && !plan.SendMax {
		return
	}

	m.logger.Warn(
		"notification channel unhealthy",
		"telegram_down", tgDown,
		"max_down", maxDown,
		"telegram_status", tgStatus.Status,
		"max_status", maxStatus.Status,
	)

	email := m.alertEmail(ctx)
	subject, bodyText, bodyHTML := formatNotificationHealthAlert(tgMonitored, tgDown, tgStatus, maxMonitored, maxDown, maxStatus)

	if plan.SendEmail && email != "" && m.mail != nil {
		if err := m.mail.Send(ctx, email, subject, bodyHTML); err != nil {
			m.logger.Warn("notification health email failed", "err", err)
		}
	}

	if plan.SendTelegram && m.telegram != nil {
		if err := m.telegram.SendDirectAdminMessage(ctx, bodyText); err != nil {
			m.logger.Warn("notification health telegram alert failed", "err", err)
		}
	}

	if plan.SendMax && m.max != nil {
		if err := m.max.SendNotification(ctx, bodyText); err != nil {
			if !errors.Is(err, ErrMaxDisabled) && !errors.Is(err, ErrMaxNotConfigured) {
				m.logger.Warn("notification health max alert failed", "err", err)
			}
		}
	}
}

func (m *NotificationHealthMonitor) checkTelegram(ctx context.Context) (model.TelegramBotRuntimeStatus, bool) {
	if m.telegram == nil {
		return model.TelegramBotRuntimeStatus{}, false
	}
	st := m.telegram.RunHealthCheck(ctx)
	monitored := isTelegramMonitored(ctx, m.telegram)
	return st, monitored
}

func (m *NotificationHealthMonitor) checkMax(ctx context.Context) (model.MaxBotRuntimeStatus, bool) {
	if m.max == nil {
		return model.MaxBotRuntimeStatus{}, false
	}
	st := m.max.CheckHealth(ctx)
	monitored := isMaxMonitored(ctx, m.max)
	return st, monitored
}

func (m *NotificationHealthMonitor) alertEmail(ctx context.Context) string {
	if m.telegram == nil {
		return "erman.ai@yandex.ru"
	}
	cfg, err := m.telegram.settings.GetEffective(ctx)
	if err != nil {
		return "erman.ai@yandex.ru"
	}
	email := strings.TrimSpace(cfg.UrgentEmail)
	if email == "" {
		return "erman.ai@yandex.ru"
	}
	return email
}

func isTelegramMonitored(ctx context.Context, telegram *TelegramService) bool {
	cfg, err := telegram.settings.GetEffective(ctx)
	if err != nil {
		return false
	}
	return cfg.Enabled || cfg.StartEnabled || cfg.SupportEnabled
}

func isMaxMonitored(ctx context.Context, max *MaxService) bool {
	cfg, err := max.settings.GetEffective(ctx)
	if err != nil {
		return false
	}
	return cfg.Enabled
}

func isTelegramUnhealthy(st model.TelegramBotRuntimeStatus) bool {
	switch st.Status {
	case model.TelegramBotStatusOffline, model.TelegramBotStatusMisconfigured:
		return true
	default:
		return false
	}
}

func isMaxUnhealthy(st model.MaxBotRuntimeStatus) bool {
	switch st.Status {
	case model.MaxBotStatusOffline, model.MaxBotStatusMisconfigured:
		return true
	default:
		return false
	}
}

func planNotificationHealthAlerts(tgMonitored, tgDown, maxMonitored, maxDown bool) notificationHealthAlertPlan {
	if (!tgMonitored || !tgDown) && (!maxMonitored || !maxDown) {
		return notificationHealthAlertPlan{}
	}

	plan := notificationHealthAlertPlan{SendEmail: true}
	tgCanSend := tgMonitored && !tgDown
	maxCanSend := maxMonitored && !maxDown

	if tgDown && maxDown {
		return plan
	}
	if tgDown && maxCanSend {
		plan.SendMax = true
		return plan
	}
	if maxDown && tgCanSend {
		plan.SendTelegram = true
		return plan
	}
	return plan
}

func formatNotificationHealthAlert(
	tgMonitored, tgDown bool,
	tgStatus model.TelegramBotRuntimeStatus,
	maxMonitored, maxDown bool,
	maxStatus model.MaxBotRuntimeStatus,
) (subject, text, html string) {
	now := time.Now().UTC().Format("02.01.2006 15:04 UTC")
	lines := []string{
		"⚠️ Erman AI — сбой канала уведомлений",
		"",
		fmt.Sprintf("Проверка: %s", now),
	}

	if tgMonitored {
		line := "Telegram: " + telegramHealthLabel(tgStatus)
		if tgDown && tgStatus.LastError != "" {
			line += "\n  " + tgStatus.LastError
		}
		lines = append(lines, line)
	}
	if maxMonitored {
		line := "MAX: " + maxHealthLabel(maxStatus)
		if maxDown && maxStatus.LastError != "" {
			line += "\n  " + maxStatus.LastError
		}
		lines = append(lines, line)
	}

	lines = append(lines, "", "Админка: https://erman.ai/dashboard/admin")
	text = strings.Join(lines, "\n")

	subject = "⚠️ Erman AI — сбой уведомлений"
	if tgDown && !maxDown {
		subject = "⚠️ Erman AI — Telegram бот недоступен"
	} else if maxDown && !tgDown {
		subject = "⚠️ Erman AI — MAX бот недоступен"
	} else if tgDown && maxDown {
		subject = "⚠️ Erman AI — Telegram и MAX недоступны"
	}

	html = fmt.Sprintf("<p>%s</p>", strings.ReplaceAll(text, "\n", "<br>"))
	return subject, text, html
}

func telegramHealthLabel(st model.TelegramBotRuntimeStatus) string {
	if st.Message != "" {
		return st.Message
	}
	return string(st.Status)
}

func maxHealthLabel(st model.MaxBotRuntimeStatus) string {
	if st.Message != "" {
		return st.Message
	}
	return string(st.Status)
}
