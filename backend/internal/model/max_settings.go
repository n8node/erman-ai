package model

import "time"

type MaxSettings struct {
	Enabled              bool   `json:"enabled"`
	BotToken             string `json:"bot_token"`
	BotUsername          string `json:"bot_username"`
	NotifyUserID         string `json:"notify_user_id"`
	NotifyChatID         string `json:"notify_chat_id"`
	UrgentAlertsEnabled  bool   `json:"urgent_alerts_enabled"`
}

type MaxSettingsRecord struct {
	Config    MaxSettings `json:"config"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type MaxBotStatus string

const (
	MaxBotStatusDisabled      MaxBotStatus = "disabled"
	MaxBotStatusMisconfigured MaxBotStatus = "misconfigured"
	MaxBotStatusOnline        MaxBotStatus = "online"
	MaxBotStatusOffline       MaxBotStatus = "offline"
)

type MaxBotRuntimeStatus struct {
	Status      MaxBotStatus `json:"status"`
	Message     string       `json:"message"`
	BotUsername string       `json:"bot_username,omitempty"`
	BotUserID   int64        `json:"bot_user_id,omitempty"`
	LastError   string       `json:"last_error,omitempty"`
	LastCheckAt time.Time    `json:"last_check_at,omitempty"`
}

type MaxAdminView struct {
	Settings     MaxSettings         `json:"settings"`
	BotTokenSet  bool                `json:"bot_token_set"`
	BotTokenHint string              `json:"bot_token_hint,omitempty"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Runtime      MaxBotRuntimeStatus `json:"runtime"`
}

type MaxAdminUpdateRequest struct {
	Settings MaxSettings `json:"settings"`
	BotToken string      `json:"bot_token,omitempty"`
}

type MaxTestResult struct {
	OK      bool                 `json:"ok"`
	Message string               `json:"message"`
	Runtime *MaxBotRuntimeStatus `json:"runtime,omitempty"`
}

func DefaultMaxSettings() MaxSettings {
	return MaxSettings{
		BotUsername:         "id504228241678_bot",
		UrgentAlertsEnabled: true,
	}
}
