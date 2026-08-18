package service

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/erman-ai/erman-ai/internal/model"
)

func TestIsStartCommand(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"/start", true},
		{"/start ref", true},
		{"/start@ErmanBot", true},
		{"/help", false},
	}
	for _, tc := range cases {
		msg := &telegramMessage{Text: tc.text}
		if got := isStartCommand(msg); got != tc.want {
			t.Fatalf("isStartCommand(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}

func TestParseUpdatesArray(t *testing.T) {
	raw := json.RawMessage(`[{"update_id":100,"message":{"message_id":1,"text":"/start","chat":{"id":639160984}}}]`)
	var updates []telegramUpdate
	if err := json.Unmarshal(raw, &updates); err != nil {
		t.Fatal(err)
	}
	if len(updates) != 1 || updates[0].Message == nil || updates[0].Message.Text != "/start" {
		t.Fatalf("unexpected parse: %+v", updates)
	}
}

func TestIsTelegramUpdatesBlocked(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("telegram api: Conflict: webhook is active"), true},
		{errors.New("telegram api: Conflict: terminated by other getUpdates request"), true},
		{errors.New("telegram api: bot was blocked"), false},
	}
	for _, tc := range cases {
		if got := isTelegramUpdatesBlocked(tc.err); got != tc.want {
			t.Fatalf("isTelegramUpdatesBlocked(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}

func TestShouldRunPolling(t *testing.T) {
	svc := &TelegramService{}
	cfg := model.TelegramSettings{StartEnabled: true, BotToken: "123:abc"}

	if !svc.shouldRunPolling(cfg, model.TelegramBotRuntimeStatus{Status: model.TelegramBotStatusOnline}) {
		t.Fatal("expected polling for online bot")
	}
	if !svc.shouldRunPolling(cfg, model.TelegramBotRuntimeStatus{Status: model.TelegramBotStatusDegraded}) {
		t.Fatal("expected polling for degraded bot when /start is enabled")
	}
	if svc.shouldRunPolling(cfg, model.TelegramBotRuntimeStatus{Status: model.TelegramBotStatusOffline}) {
		t.Fatal("expected no polling when bot is offline")
	}
	if svc.shouldRunPolling(model.TelegramSettings{}, model.TelegramBotRuntimeStatus{Status: model.TelegramBotStatusOnline}) {
		t.Fatal("expected no polling when bot features are disabled")
	}
}
