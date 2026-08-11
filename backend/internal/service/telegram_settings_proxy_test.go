package service

import (
	"testing"

	"github.com/erman-ai/erman-ai/internal/model"
)

func TestNormalizeTelegramSettingsFromStorageEnablesProxyWhenURLsPresent(t *testing.T) {
	cfg := model.TelegramSettings{
		ProxyEnabled: false,
		ProxyURLs:    []string{"http://127.0.0.1:3128"},
	}
	normalizeTelegramSettingsFromStorage(&cfg)
	if !cfg.ProxyEnabled {
		t.Fatal("expected proxy_enabled=true when proxy urls exist")
	}
	if cfg.ProxyActiveURL != "http://127.0.0.1:3128" {
		t.Fatalf("expected active proxy auto-selected, got %q", cfg.ProxyActiveURL)
	}
}

func TestMergeTelegramSettingsUpdatePreservesStoredProxy(t *testing.T) {
	stored := model.TelegramSettings{
		ProxyEnabled:   true,
		ProxyURLs:      []string{"http://127.0.0.1:3128"},
		ProxyActiveURL: "http://127.0.0.1:3128",
	}
	incoming := model.TelegramSettings{
		Enabled:      true,
		ChatID:       "-100123",
		ProxyEnabled: false,
		ProxyURLs:    nil,
	}
	merged := mergeTelegramSettingsUpdate(stored, incoming)
	if !merged.ProxyEnabled {
		t.Fatal("expected stored proxy_enabled to be preserved")
	}
	if len(merged.ProxyURLs) != 1 || merged.ProxyURLs[0] != "http://127.0.0.1:3128" {
		t.Fatalf("expected stored proxy urls to be preserved, got %#v", merged.ProxyURLs)
	}
}

func TestParseTelegramProxyURLs(t *testing.T) {
	urls := parseTelegramProxyURLs("http://a:1\nhttp://b:2,http://c:3")
	if len(urls) != 3 {
		t.Fatalf("expected 3 urls, got %d", len(urls))
	}
}
