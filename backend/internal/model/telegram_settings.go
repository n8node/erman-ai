package model

import "time"

type TelegramSettings struct {
	Enabled                 bool   `json:"enabled"`
	ChatID                  string `json:"chat_id"`
	BotToken                string `json:"bot_token"`
	NotifyRegistration      bool   `json:"notify_registration"`
	RegistrationTemplate    string `json:"registration_template"`
	NotifyEmailVerified     bool   `json:"notify_email_verified"`
	EmailVerifiedTemplate   string `json:"email_verified_template"`
	NotifyPayment           bool   `json:"notify_payment"`
	PaymentTemplate         string `json:"payment_template"`
}

type TelegramSettingsRecord struct {
	Config    TelegramSettings `json:"config"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type TelegramBotStatus string

const (
	TelegramBotStatusDisabled      TelegramBotStatus = "disabled"
	TelegramBotStatusMisconfigured TelegramBotStatus = "misconfigured"
	TelegramBotStatusStarting      TelegramBotStatus = "starting"
	TelegramBotStatusOnline        TelegramBotStatus = "online"
	TelegramBotStatusOffline       TelegramBotStatus = "offline"
)

type TelegramBotRuntimeStatus struct {
	Status            TelegramBotStatus `json:"status"`
	Message           string            `json:"message"`
	BotUsername       string            `json:"bot_username,omitempty"`
	LastError         string            `json:"last_error,omitempty"`
	LastCheckAt       time.Time         `json:"last_check_at,omitempty"`
	SupervisorRunning bool              `json:"supervisor_running"`
}

type TelegramAdminView struct {
	Settings     TelegramSettings         `json:"settings"`
	BotTokenSet  bool                     `json:"bot_token_set"`
	BotTokenHint string                   `json:"bot_token_hint,omitempty"`
	UpdatedAt    time.Time                `json:"updated_at"`
	Runtime      TelegramBotRuntimeStatus `json:"runtime"`
}

type TelegramAdminUpdateRequest struct {
	Settings  TelegramSettings `json:"settings"`
	BotToken  string           `json:"bot_token,omitempty"`
}

type TelegramTestResult struct {
	OK      bool                     `json:"ok"`
	Message string                   `json:"message"`
	Runtime *TelegramBotRuntimeStatus `json:"runtime,omitempty"`
}

func DefaultTelegramSettings() TelegramSettings {
	return TelegramSettings{
		NotifyRegistration:    true,
		RegistrationTemplate:  "🆕 Новый пользователь\nEmail: {email}\nСегмент: {accountSegment}\nРеферал: {referral}",
		NotifyEmailVerified:   true,
		EmailVerifiedTemplate: "✅ Email подтверждён\nEmail: {email}",
		NotifyPayment:         true,
		PaymentTemplate:       "💰 Оплата тарифа\nПользователь: {userEmail}\nТариф: {planName}\nСумма: {amount} {currency}",
	}
}
