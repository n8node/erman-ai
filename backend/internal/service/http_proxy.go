package service

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
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

func transportViaHTTPConnectProxy(proxyURL *url.URL) *http.Transport {
	proxy := cloneProxyURL(proxyURL)
	return &http.Transport{
		Proxy:                 nil,
		ForceAttemptHTTP2:     false,
		TLSNextProto:          map[string]func(string, *tls.Conn) http.RoundTripper{},
		DialContext:           directDialContext,
		DialTLSContext:        dialTLSViaHTTPConnectProxy(proxy),
	}
}

func cloneProxyURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	c := *u
	return &c
}

func directDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	var dialer net.Dialer
	return dialer.DialContext(ctx, network, addr)
}

func dialTLSViaHTTPConnectProxy(proxyURL *url.URL) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if network != "tcp" && network != "tcp4" && network != "tcp6" {
			return nil, fmt.Errorf("unsupported network %q", network)
		}
		conn, err := connectViaHTTPProxy(ctx, proxyURL, addr)
		if err != nil {
			return nil, err
		}
		host, _, splitErr := net.SplitHostPort(addr)
		if splitErr != nil {
			conn.Close()
			return nil, splitErr
		}
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"http/1.1"},
		})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			conn.Close()
			return nil, err
		}
		return tlsConn, nil
	}
}

func connectViaHTTPProxy(ctx context.Context, proxyURL *url.URL, targetAddr string) (net.Conn, error) {
	if proxyURL == nil {
		return nil, errors.New("proxy url is required")
	}
	scheme := strings.ToLower(strings.TrimSpace(proxyURL.Scheme))
	if scheme != "http" {
		return nil, fmt.Errorf("unsupported proxy scheme %q (use http://)", proxyURL.Scheme)
	}

	proxyAddr := proxyURL.Host
	if _, _, err := net.SplitHostPort(proxyAddr); err != nil {
		proxyAddr = net.JoinHostPort(proxyAddr, "80")
	}

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("proxy dial: %w", err)
	}

	var req strings.Builder
	req.WriteString("CONNECT ")
	req.WriteString(targetAddr)
	req.WriteString(" HTTP/1.1\r\nHost: ")
	req.WriteString(targetAddr)
	req.WriteString("\r\n")
	if proxyURL.User != nil {
		user := proxyURL.User.Username()
		pass, _ := proxyURL.User.Password()
		token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
		req.WriteString("Proxy-Authorization: Basic ")
		req.WriteString(token)
		req.WriteString("\r\n")
	}
	req.WriteString("User-Agent: ErmanAI-ProxyConnect/1.0\r\n")
	req.WriteString("Proxy-Connection: Keep-Alive\r\n")
	req.WriteString("\r\n")

	if _, err := conn.Write([]byte(req.String())); err != nil {
		conn.Close()
		return nil, fmt.Errorf("proxy connect write: %w", err)
	}

	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, &http.Request{Method: http.MethodConnect})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("proxy connect response: %w", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		conn.Close()
		return nil, fmt.Errorf("proxy connect status %s", resp.Status)
	}
	return conn, nil
}

func httpClientForProxy(base *http.Client, proxyURL string) (*http.Client, error) {
	parsed, err := url.Parse(strings.TrimSpace(proxyURL))
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid proxy url %q", proxyURL)
	}
	if base == nil {
		base = &http.Client{}
	}
	return &http.Client{
		Timeout:   base.Timeout,
		Transport: transportViaHTTPConnectProxy(parsed),
	}, nil
}

func normalizeLLMHTTPProxySettings(cfg *model.LLMHTTPProxySettings) {
	if cfg == nil {
		return
	}
	cfg.ProxyURLs = normalizeProxyURLs(cfg.ProxyURLs)
	cfg.ProxyActiveURL = strings.TrimSpace(cfg.ProxyActiveURL)
	if len(cfg.ProxyURLs) > 0 {
		cfg.Enabled = true
	}
	if cfg.Enabled && cfg.ProxyActiveURL == "" && len(cfg.ProxyURLs) > 0 {
		cfg.ProxyActiveURL = cfg.ProxyURLs[0]
	}
	if !cfg.Enabled {
		cfg.ProxyActiveURL = ""
	}
}

func llmProxyRouteLabel(proxy *model.LLMHTTPProxySettings) string {
	if proxy == nil || !proxy.Enabled || len(proxy.ProxyURLs) == 0 {
		return "route: direct (proxy disabled — required for OpenRouter from RU server)"
	}
	active := strings.TrimSpace(proxy.ProxyActiveURL)
	if active == "" {
		active = proxy.ProxyURLs[0]
	}
	return "route: " + maskProxyURL(active)
}

func maskProxyURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" {
		return "proxy configured"
	}
	host := parsed.Hostname()
	port := parsed.Scheme + "://" + host
	if p := parsed.Port(); p != "" {
		port += ":" + p
	}
	if parsed.User != nil {
		return port + " (authenticated)"
	}
	return port
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
		if strings.ToLower(u.Scheme) != "http" {
			return fmt.Errorf("proxy url %q: only http:// proxies are supported", raw)
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
