package model

import "time"

type TelegramSupportThread struct {
	UserChatID   string    `json:"user_chat_id"`
	ForumChatID  string    `json:"forum_chat_id"`
	TopicID      int       `json:"topic_id"`
	DisplayName  string    `json:"display_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
