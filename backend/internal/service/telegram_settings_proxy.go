package service

import (
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

func parseTelegramProxyURLs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ','
	})
	return normalizeProxyURLs(parts)
}

func normalizeTelegramSettingsFromStorage(cfg *model.TelegramSettings) {
	cfg.ProxyURLs = normalizeProxyURLs(cfg.ProxyURLs)
	cfg.ProxyActiveURL = strings.TrimSpace(cfg.ProxyActiveURL)
	if cfg.ProxyEnabled && cfg.ProxyActiveURL == "" && len(cfg.ProxyURLs) > 0 {
		cfg.ProxyActiveURL = cfg.ProxyURLs[0]
	}
}

func mergeTelegramSettingsUpdate(stored, incoming model.TelegramSettings) model.TelegramSettings {
	out := incoming
	out.ProxyURLs = normalizeProxyURLs(out.ProxyURLs)
	out.ProxyActiveURL = strings.TrimSpace(out.ProxyActiveURL)

	if len(out.ProxyURLs) == 0 && len(stored.ProxyURLs) > 0 {
		out.ProxyURLs = append([]string(nil), stored.ProxyURLs...)
	}
	if out.ProxyActiveURL == "" && stored.ProxyActiveURL != "" && containsProxyURL(out.ProxyURLs, stored.ProxyActiveURL) {
		out.ProxyActiveURL = stored.ProxyActiveURL
	}
	if !out.ProxyEnabled && stored.ProxyEnabled && len(out.ProxyURLs) > 0 && len(incoming.ProxyURLs) == 0 {
		out.ProxyEnabled = true
	}

	if out.ProxyEnabled && out.ProxyActiveURL == "" && len(out.ProxyURLs) > 0 {
		out.ProxyActiveURL = out.ProxyURLs[0]
	}
	if !out.ProxyEnabled {
		out.ProxyActiveURL = ""
	}
	return out
}

func telegramProxyConfigEqual(a, b model.TelegramSettings) bool {
	if a.ProxyEnabled != b.ProxyEnabled ||
		a.ProxyActiveURL != b.ProxyActiveURL ||
		a.ProxyAutoFailover != b.ProxyAutoFailover {
		return false
	}
	if len(a.ProxyURLs) != len(b.ProxyURLs) {
		return false
	}
	for i := range a.ProxyURLs {
		if a.ProxyURLs[i] != b.ProxyURLs[i] {
			return false
		}
	}
	return true
}
