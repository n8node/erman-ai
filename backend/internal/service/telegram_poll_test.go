package service

import "testing"

func TestIsStartCommand(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"/start", true},
		{"/start ", true},
		{"/start ref", true},
		{"/start@ErmanBot", true},
		{"/start@ErmanBot promo", true},
		{"/help", false},
		{"", false},
	}
	for _, tc := range cases {
		msg := &telegramMessage{Text: tc.text, Entities: []telegramEntity{{Type: "bot_command", Offset: 0, Length: len(tc.text)}}}
		if len(tc.text) > 0 && tc.text[0] == '/' {
			// entity length for simple ASCII commands
			msg.Entities[0].Length = len(tc.text)
			if i := indexSpace(tc.text); i > 0 {
				msg.Entities[0].Length = i
			}
		}
		got := isStartCommand(msg)
		if got != tc.want {
			t.Fatalf("isStartCommand(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}

func indexSpace(s string) int {
	for i, r := range s {
		if r == ' ' {
			return i
		}
	}
	return -1
}
