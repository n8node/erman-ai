package service

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	telegramBotTokenPattern = regexp.MustCompile(`^\d{8,12}:[A-Za-z0-9_-]{20,}$`)
	maxBotTokenPattern      = regexp.MustCompile(`^[A-Za-z0-9]{32,128}$`)
)

func validateTelegramBotToken(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	if strings.Contains(token, " ") || strings.ContainsRune(token, '\n') {
		return fmt.Errorf("invalid telegram bot token format: must be a single line without spaces")
	}
	if !telegramBotTokenPattern.MatchString(token) {
		return fmt.Errorf("invalid telegram bot token format: expected 123456789:AAH... from @BotFather")
	}
	return nil
}

func validateMaxBotToken(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	if strings.Contains(token, " ") || strings.ContainsRune(token, '\n') {
		return fmt.Errorf("invalid max bot token format: must be a single line without spaces")
	}
	if !maxBotTokenPattern.MatchString(token) {
		return fmt.Errorf("invalid max bot token format: expected alphanumeric token from dev.max.ru")
	}
	return nil
}
