package service

import (
	"encoding/json"
	"testing"
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
