package service

import (
	"encoding/json"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

const telegramCallbackConsultation = "consultation"

func startReplyMarkup(cfg model.TelegramSettings) ([]byte, error) {
	dashboardURL := strings.TrimSpace(cfg.DashboardURL)
	if dashboardURL == "" {
		dashboardURL = "https://erman.ai/dashboard/"
	}
	markup := map[string]any{
		"inline_keyboard": [][]map[string]string{
			{
				{"text": "Консультация", "callback_data": telegramCallbackConsultation},
				{"text": "Инструменты", "url": dashboardURL},
			},
		},
	}
	return json.Marshal(markup)
}
