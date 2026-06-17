package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
)

const (
	legalScanDefaultMaxPages  = 10
	legalScanDefaultMaxDepth  = 2
	legalScanCrawlMaxPagesCap = 50
	legalScanMaxTotalBytes    = 10 << 20 // 10 MB
	legalScanCrawlDelay       = 250 * time.Millisecond
)

var legalScanPlanPageLimits = map[string]int{
	"free":     5,
	"pro":      20,
	"business": 50,
}

type legalScanCrawlResult struct {
	StartURL   string
	Pages      []fetchedPage
	Meta       model.LegalScanCrawlMeta
	HTTPS      bool
	TotalBytes int64
}

type legalScanCrawlOptions struct {
	MaxPages     int
	MaxDepth     int
	SameHostOnly bool
	OnProgress   func(meta model.LegalScanCrawlMeta, layer1 model.LegalScanLayer1)
}

func normalizeLegalScanCrawlConfig(c *model.LegalScanCrawlConfig, planSlug string) {
	if c.MaxPages <= 0 {
		c.MaxPages = legalScanDefaultMaxPages
	}
	if c.MaxDepth < 0 {
		c.MaxDepth = legalScanDefaultMaxDepth
	}
	planCap := legalScanPlanPageLimit(planSlug)
	if c.MaxPages > planCap {
		c.MaxPages = planCap
	}
	if c.MaxPages > legalScanCrawlMaxPagesCap {
		c.MaxPages = legalScanCrawlMaxPagesCap
	}
}

func legalScanPlanPageLimit(planSlug string) int {
	if lim, ok := legalScanPlanPageLimits[planSlug]; ok {
		return lim
	}
	return legalScanPlanPageLimits["free"]
}

func legalScanRunTimeoutFor(pages int) time.Duration {
	sec := 30 + pages*4
	if sec > 180 {
		sec = 180
	}
	return time.Duration(sec) * time.Second
}

func crawlLegalScanSite(
	ctx context.Context,
	startURL string,
	features model.LegalScanSiteFeatures,
	opts legalScanCrawlOptions,
) (*legalScanCrawlResult, error) {
	start := time.Now()
	parsed, err := url.Parse(startURL)
	if err != nil || parsed.Host == "" {
		return nil, fmt.Errorf("%w: invalid url", ErrInvalidInput)
	}
	baseHost := parsed.Hostname()

	type queueItem struct {
		url   string
		depth int
		pri   int
	}

	visited := make(map[string]struct{})
	var fetched []fetchedPage
	var fetchedURLs []string
	skipped := 0
	var totalBytes int64
	httpsOK := false
	var lastFetchErr error

	var queue []queueItem
	startNorm := normalizeLegalScanURL(startURL)
	if startNorm != "" {
		visited[startNorm] = struct{}{}
	}
	queue = append(queue, queueItem{url: startURL, depth: 0, pri: 1000})

crawlLoop:
	for len(queue) > 0 && len(fetched) < opts.MaxPages {
		if ctx.Err() != nil {
			break
		}
		if totalBytes >= legalScanMaxTotalBytes {
			break
		}

		best := 0
		for i := 1; i < len(queue); i++ {
			if queue[i].pri > queue[best].pri {
				best = i
			}
		}
		item := queue[best]
		queue = append(queue[:best], queue[best+1:]...)

		page, err := fetchLegalScanSite(ctx, item.url)
		if err != nil {
			lastFetchErr = err
			skipped++
			continue
		}

		fetched = append(fetched, *page)
		fetchedURLs = append(fetchedURLs, page.URL)
		totalBytes += int64(len(page.HTML))
		if page.HTTPS {
			httpsOK = true
		}

		meta := model.LegalScanCrawlMeta{
			StartURL:       startURL,
			PagesRequested: opts.MaxPages,
			PagesFetched:   len(fetched),
			MaxDepth:       opts.MaxDepth,
			FetchedURLs:    append([]string(nil), fetchedURLs...),
			SkippedCount:   skipped,
			DurationMs:     int(time.Since(start).Milliseconds()),
		}

		if opts.OnProgress != nil {
			layer1 := runLegalScanLayer1FromPages(fetched, startURL, httpsOK, features, &meta)
			opts.OnProgress(meta, layer1)
		}

		if item.depth >= opts.MaxDepth || len(fetched) >= opts.MaxPages {
			select {
			case <-ctx.Done():
				break crawlLoop
			case <-time.After(legalScanCrawlDelay):
			}
			continue
		}

		links := extractLegalScanLinks(page.HTML, page.URL, item.depth+1, baseHost, opts.SameHostOnly)
		sortScoredLinks(links)
		for _, link := range links {
			norm := normalizeLegalScanURL(link.URL)
			if norm == "" {
				continue
			}
			if _, ok := visited[norm]; ok {
				continue
			}
			visited[norm] = struct{}{}
			queue = append(queue, queueItem{url: link.URL, depth: link.Depth, pri: link.Priority})
		}

		select {
		case <-ctx.Done():
			break crawlLoop
		case <-time.After(legalScanCrawlDelay):
		}
	}

	meta := model.LegalScanCrawlMeta{
		StartURL:       startURL,
		PagesRequested: opts.MaxPages,
		PagesFetched:   len(fetched),
		MaxDepth:       opts.MaxDepth,
		FetchedURLs:    fetchedURLs,
		SkippedCount:   skipped,
		DurationMs:     int(time.Since(start).Milliseconds()),
	}

	if len(fetched) == 0 {
		if lastFetchErr != nil {
			return nil, lastFetchErr
		}
		return nil, errors.New("no pages fetched")
	}

	return &legalScanCrawlResult{
		StartURL:   startURL,
		Pages:      fetched,
		Meta:       meta,
		HTTPS:      httpsOK,
		TotalBytes: totalBytes,
	}, nil
}
