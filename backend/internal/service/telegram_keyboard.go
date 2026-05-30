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
	rows := [][]map[string]string{
		{
			{"text": "Консультация", "callback_data": telegramCallbackConsultation},
			{"text": "Инструменты", "url": dashboardURL},
		},
	}
	if cfg.UrgentEnabled {
		rows = append(rows, []map[string]string{
			{"text": "Срочно связаться", "callback_data": telegramCallbackUrgent},
		})
	}
	markup := map[string]any{"inline_keyboard": rows}
	return json.Marshal(markup)
}
