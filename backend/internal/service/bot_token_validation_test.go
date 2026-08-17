package service

import "testing"

func TestValidateTelegramBotToken(t *testing.T) {
	valid := "7581464736:AAHnKnofzZNT9vmor0YDRqHej1ZXVhuLlTY"
	if err := validateTelegramBotToken(valid); err != nil {
		t.Fatalf("expected valid telegram token, got %v", err)
	}

	cases := []string{
		"",
		"UA.xxxxx.app",
		"f9LHodD0cOKMJTYFshjrcM4LHJndaggdE6vyv5xaUZV2LUKp41CDgjRihGLKKy3ycrd9VobzaTLhjuyimwHi",
		"7581464736",
		":AAHnKnofzZNT9vmor0YDRqHej1ZXVhuLlTY",
		"7581464736:AA short",
		"7581464736:AAHnKnofzZNT9vmor0YDRqHej1ZXVhu LlTY",
	}
	for _, token := range cases {
		if token == "" {
			if err := validateTelegramBotToken(token); err != nil {
				t.Fatalf("empty token should be allowed, got %v", err)
			}
			continue
		}
		if err := validateTelegramBotToken(token); err == nil {
			t.Fatalf("expected invalid telegram token %q", token)
		}
	}
}

func TestValidateMaxBotToken(t *testing.T) {
	valid := "f9LHodD0cOKMJTYFshjrcM4LHJndaggdE6vyv5xaUZV2LUKp41CDgjRihGLKKy3ycrd9VobzaTLhjuyimwHi"
	if err := validateMaxBotToken(valid); err != nil {
		t.Fatalf("expected valid max token, got %v", err)
	}

	cases := []string{
		"7581464736:AAHnKnofzZNT9vmor0YDRqHej1ZXVhuLlTY",
		"UA.xxxxx.app",
		"tooshort",
		"f9LHodD0cOKMJTYFshjrcM4LHJndaggdE6vyv5xaUZV2LUKp41CDgjRihGLKKy3ycrd9VobzaTLhjuyimwHi!",
	}
	for _, token := range cases {
		if err := validateMaxBotToken(token); err == nil {
			t.Fatalf("expected invalid max token %q", token)
		}
	}
}
