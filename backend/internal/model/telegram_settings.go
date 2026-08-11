package model

import "time"

const (
	TelegramCaptionMaxRunes = 1024
	TelegramMessageMaxRunes = 4096
	TelegramPhotoMaxBytes   = 10 * 1024 * 1024
)

type TelegramSettings struct {
	Enabled               bool     `json:"enabled"`
	ChatID                string   `json:"chat_id"`
	BotToken              string   `json:"bot_token"`
	ProxyEnabled          bool     `json:"proxy_enabled"`
	ProxyActiveURL        string   `json:"proxy_active_url"`
	ProxyAutoFailover     bool     `json:"proxy_auto_failover"`
	ProxyURLs             []string `json:"proxy_urls"`
	StartEnabled          bool     `json:"start_enabled"`
	StartText             string   `json:"start_text"`
	StartImageFilename    string   `json:"start_image_filename"`
	SupportEnabled        bool     `json:"support_enabled"`
	SupportForumChatID    string   `json:"support_forum_chat_id"`
	DashboardURL          string   `json:"dashboard_url"`
	UrgentEnabled         bool     `json:"urgent_enabled"`
	UrgentEmail           string   `json:"urgent_email"`
	UrgentInstruction     string   `json:"urgent_instruction"`
	NotifyRegistration    bool     `json:"notify_registration"`
	RegistrationTemplate  string   `json:"registration_template"`
	NotifyEmailVerified   bool     `json:"notify_email_verified"`
	EmailVerifiedTemplate string   `json:"email_verified_template"`
	NotifyPayment         bool     `json:"notify_payment"`
	PaymentTemplate       string   `json:"payment_template"`
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
	TelegramBotStatusDegraded      TelegramBotStatus = "degraded"
	TelegramBotStatusOffline       TelegramBotStatus = "offline"
)

type TelegramBotRuntimeStatus struct {
	Status            TelegramBotStatus `json:"status"`
	Message           string            `json:"message"`
	BotUsername       string            `json:"bot_username,omitempty"`
	LastError         string            `json:"last_error,omitempty"`
	LastCheckAt       time.Time         `json:"last_check_at,omitempty"`
	SupervisorRunning bool              `json:"supervisor_running"`
	PollingRunning    bool              `json:"polling_running"`
}

type TelegramAdminView struct {
	Settings             TelegramSettings         `json:"settings"`
	BotTokenSet          bool                     `json:"bot_token_set"`
	BotTokenHint         string                   `json:"bot_token_hint,omitempty"`
	StartImageConfigured bool                     `json:"start_image_configured"`
	StartTextRunes       int                      `json:"start_text_runes"`
	StartTextLimit       int                      `json:"start_text_limit"`
	StartCaptionLimit    int                      `json:"start_caption_limit"`
	UpdatedAt            time.Time                `json:"updated_at"`
	Runtime              TelegramBotRuntimeStatus `json:"runtime"`
}

type TelegramAdminUpdateRequest struct {
	Settings        TelegramSettings `json:"settings"`
	BotToken        string           `json:"bot_token,omitempty"`
	ClearStartImage bool             `json:"clear_start_image,omitempty"`
}

type TelegramTestResult struct {
	OK      bool                      `json:"ok"`
	Message string                    `json:"message"`
	Runtime *TelegramBotRuntimeStatus `json:"runtime,omitempty"`
}

func DefaultUrgentInstruction() string {
	return "Если вы хотите срочно связаться со мной, напишите ниже ваше сообщение — оно будет разослано по моим контактам: email, Telegram, MAX, VK, Instagram.\n\n" +
		"В ответ я смогу написать вам в Telegram или по контактам, которые вы укажете в сообщении."
}

func DefaultTelegramSettings() TelegramSettings {
	return TelegramSettings{
		StartText:             "Добро пожаловать в Erman AI!\n\nВыберите действие:",
		DashboardURL:          "https://erman.ai/dashboard/",
		ProxyEnabled:          false,
		ProxyAutoFailover:     true,
		ProxyURLs:             []string{},
		UrgentEnabled:         true,
		UrgentEmail:           "erman.ai@yandex.ru",
		UrgentInstruction:     DefaultUrgentInstruction(),
		NotifyRegistration:    true,
		RegistrationTemplate:  "🆕 Новый пользователь\nEmail: {email}\nСегмент: {accountSegment}\nРеферал: {referral}",
		NotifyEmailVerified:   true,
		EmailVerifiedTemplate: "✅ Email подтверждён\nEmail: {email}",
		NotifyPayment:         true,
		PaymentTemplate:       "💰 Оплата тарифа\nПользователь: {userEmail}\nТариф: {planName}\nСумма: {amount} {currency}",
	}
}
