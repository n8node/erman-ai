package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
)

func normalizeProxyURLs(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, raw := range in {
		clean := strings.TrimSpace(raw)
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}

func containsProxyURL(items []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, item := range items {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}

func proxyOrder(activeURL string, urls []string) []string {
	if len(urls) == 0 {
		return nil
	}
	active := strings.TrimSpace(activeURL)
	if active == "" || !slices.Contains(urls, active) {
		return urls
	}
	out := make([]string, 0, len(urls))
	out = append(out, active)
	for _, raw := range urls {
		if raw == active {
			continue
		}
		out = append(out, raw)
	}
	return out
}

func httpClientForProxy(base *http.Client, proxyURL string) (*http.Client, error) {
	parsed, err := url.Parse(strings.TrimSpace(proxyURL))
	if err != nil {
		return nil, err
	}
	if base == nil {
		base = &http.Client{}
	}
	var baseTransport *http.Transport
	if base.Transport != nil {
		if t, ok := base.Transport.(*http.Transport); ok {
			baseTransport = t.Clone()
		}
	}
	if baseTransport == nil {
		baseTransport = &http.Transport{}
	}
	baseTransport.Proxy = http.ProxyURL(parsed)
	return &http.Client{
		Timeout:   base.Timeout,
		Transport: baseTransport,
	}, nil
}

func normalizeLLMHTTPProxySettings(cfg *model.LLMHTTPProxySettings) {
	if cfg == nil {
		return
	}
	cfg.ProxyURLs = normalizeProxyURLs(cfg.ProxyURLs)
	cfg.ProxyActiveURL = strings.TrimSpace(cfg.ProxyActiveURL)
	if cfg.Enabled && cfg.ProxyActiveURL == "" && len(cfg.ProxyURLs) > 0 {
		cfg.ProxyActiveURL = cfg.ProxyURLs[0]
	}
	if !cfg.Enabled {
		cfg.ProxyActiveURL = ""
	}
}

func validateLLMHTTPProxySettings(cfg model.LLMHTTPProxySettings) error {
	cfg.ProxyURLs = normalizeProxyURLs(cfg.ProxyURLs)
	cfg.ProxyActiveURL = strings.TrimSpace(cfg.ProxyActiveURL)
	if !cfg.Enabled {
		return nil
	}
	if len(cfg.ProxyURLs) == 0 {
		return fmt.Errorf("add at least one proxy url")
	}
	for _, raw := range cfg.ProxyURLs {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("invalid proxy url %q", raw)
		}
	}
	if cfg.ProxyActiveURL == "" {
		cfg.ProxyActiveURL = cfg.ProxyURLs[0]
	}
	if !containsProxyURL(cfg.ProxyURLs, cfg.ProxyActiveURL) {
		return fmt.Errorf("active proxy must be one of proxy_urls")
	}
	return nil
}

func doHTTPWithProxy(
	ctx context.Context,
	baseClient *http.Client,
	proxy *model.LLMHTTPProxySettings,
	makeRequest func(*http.Client) (*http.Response, error),
) (*http.Response, error) {
	if baseClient == nil {
		baseClient = &http.Client{}
	}
	if proxy == nil || !proxy.Enabled || len(proxy.ProxyURLs) == 0 {
		return makeRequest(baseClient)
	}

	proxies := proxyOrder(proxy.ProxyActiveURL, normalizeProxyURLs(proxy.ProxyURLs))
	var lastErr error
	for idx, proxyURL := range proxies {
		proxyClient, err := httpClientForProxy(baseClient, proxyURL)
		if err != nil {
			lastErr = fmt.Errorf("proxy %q: %w", proxyURL, err)
			if !proxy.ProxyAutoFailover {
				return nil, lastErr
			}
			continue
		}
		resp, reqErr := makeRequest(proxyClient)
		if reqErr == nil {
			return resp, nil
		}
		lastErr = fmt.Errorf("proxy %q: %w", proxyURL, reqErr)
		if !proxy.ProxyAutoFailover || idx == len(proxies)-1 {
			return nil, lastErr
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("proxy request failed")
}
