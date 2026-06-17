package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const legalScanFetchTimeout = 25 * time.Second
const legalScanMaxBodyBytes = 2 << 20 // 2 MB

var errLegalScanFetchBlocked = errors.New("url not allowed")

type fetchedPage struct {
	URL        string
	HTML       string
	HTTPS      bool
	StatusCode int
}

func fetchLegalScanSite(ctx context.Context, rawURL string) (*fetchedPage, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%w: invalid url", ErrInvalidInput)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("%w: only http/https allowed", ErrInvalidInput)
	}
	if err := validateLegalScanHost(parsed.Hostname()); err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: legalScanFetchTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			if err := validateLegalScanHost(req.URL.Hostname()); err != nil {
				return err
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ErmanAI-LegalScan/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, legalScanMaxBodyBytes))
	if err != nil {
		return nil, err
	}

	finalURL := resp.Request.URL.String()
	return &fetchedPage{
		URL:        finalURL,
		HTML:       string(body),
		HTTPS:      strings.HasPrefix(strings.ToLower(finalURL), "https://"),
		StatusCode: resp.StatusCode,
	}, nil
}

func validateLegalScanHost(host string) error {
	host = strings.Trim(host, "[]")
	if host == "" || host == "localhost" {
		return errLegalScanFetchBlocked
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return errLegalScanFetchBlocked
		}
	}
	lower := strings.ToLower(host)
	if lower == "127.0.0.1" || strings.HasSuffix(lower, ".local") {
		return errLegalScanFetchBlocked
	}
	return nil
}
