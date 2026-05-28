package service

import (
	"context"
	"os"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

func (s *TelegramService) sendStartReply(ctx context.Context, cfg model.TelegramSettings, userChatID string) error {
	if !cfg.StartEnabled {
		return nil
	}

	token := strings.TrimSpace(cfg.BotToken)
	if token == "" {
		return ErrTelegramNotConfigured
	}

	text := strings.TrimSpace(cfg.StartText)
	imagePath := ""
	if cfg.StartImageFilename != "" && s.assets != nil {
		imagePath = s.assets.StartImagePath(cfg.StartImageFilename)
		if _, err := os.Stat(imagePath); err != nil {
			imagePath = ""
		}
	}

	if imagePath != "" {
		caption, remainder := splitCaptionAndRemainder(text, model.TelegramCaptionMaxRunes)
		if err := s.telegramSendPhotoFile(ctx, token, userChatID, imagePath, caption); err != nil {
			return err
		}
		if remainder != "" {
			return s.telegramSendMessage(ctx, token, userChatID, remainder)
		}
		return nil
	}

	if text == "" {
		return nil
	}
	return s.telegramSendMessage(ctx, token, userChatID, truncateRunes(text, model.TelegramMessageMaxRunes))
}

func splitCaptionAndRemainder(text string, captionMax int) (caption, remainder string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ""
	}
	if runeLen(text) <= captionMax {
		return text, ""
	}
	runes := []rune(text)
	return string(runes[:captionMax]), strings.TrimSpace(string(runes[captionMax:]))
}
