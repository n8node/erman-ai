package model

import "time"

const (
	TelegramUserModeIdle       = "idle"
	TelegramUserModeUrgentWait = "urgent_wait"
)

type TelegramBotUserState struct {
	UserChatID string    `json:"user_chat_id"`
	Mode       string    `json:"mode"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type TelegramUrgentSend struct {
	ID             string    `json:"id"`
	UserChatID     string    `json:"user_chat_id"`
	MessagePreview string    `json:"message_preview"`
	SentAt         time.Time `json:"sent_at"`
}

const TelegramUrgentDailyLimit = 2
