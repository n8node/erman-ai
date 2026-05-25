package service

import (
	"context"
	"strings"

	"github.com/erman-ai/erman-ai/internal/i18n"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

type TooltipService struct {
	repo *repository.TooltipRepository
}

func NewTooltipService(repo *repository.TooltipRepository) *TooltipService {
	return &TooltipService{repo: repo}
}

func (s *TooltipService) ListForLocale(ctx context.Context, prefix, locale string) (map[string]string, error) {
	items, err := s.repo.ListByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(items))
	for _, t := range items {
		out[t.Key] = pickTooltipText(t, locale)
	}
	return out, nil
}

func (s *TooltipService) ListAll(ctx context.Context) ([]model.UITooltip, error) {
	return s.repo.ListAll(ctx)
}

func (s *TooltipService) Update(ctx context.Context, key, textRU, textEN string) (*model.UITooltip, error) {
	return s.repo.Update(ctx, key, textRU, textEN)
}

func pickTooltipText(t model.UITooltip, locale string) string {
	locale = i18n.NormalizeLocale(locale)
	if t.LocaleTexts != nil {
		if v := strings.TrimSpace(t.LocaleTexts[locale]); v != "" {
			return v
		}
		for _, fb := range []string{"en", "ru"} {
			if v := strings.TrimSpace(t.LocaleTexts[fb]); v != "" {
				return v
			}
		}
	}
	if locale == "en" && t.TextEN != "" {
		return t.TextEN
	}
	if locale != "ru" && t.TextEN != "" {
		return t.TextEN
	}
	if t.TextRU != "" {
		return t.TextRU
	}
	return t.TextEN
}
