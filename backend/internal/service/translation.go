package service

import (
	"context"

	"github.com/erman-ai/erman-ai/internal/i18n"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

type TranslationService struct {
	repo *repository.TranslationRepository
}

func NewTranslationService(repo *repository.TranslationRepository) *TranslationService {
	return &TranslationService{repo: repo}
}

func (s *TranslationService) ListPublic(ctx context.Context, locale string) (map[string]string, error) {
	locale = i18n.NormalizeLocale(locale)
	return s.repo.ListByLocale(ctx, locale)
}

func (s *TranslationService) SearchAdmin(ctx context.Context, locale, search string, limit, offset int) ([]model.UITranslation, error) {
	if locale != "" {
		locale = i18n.NormalizeLocale(locale)
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 10000 {
		limit = 10000
	}
	return s.repo.Search(ctx, locale, search, limit, offset)
}

func (s *TranslationService) Upsert(ctx context.Context, key, locale, value string) (*model.UITranslation, error) {
	locale = i18n.NormalizeLocale(locale)
	if key == "" || value == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.Upsert(ctx, key, locale, value)
}

func (s *TranslationService) BulkUpsert(ctx context.Context, items []model.UITranslation) ([]model.UITranslation, error) {
	out := make([]model.UITranslation, 0, len(items))
	for _, item := range items {
		t, err := s.Upsert(ctx, item.Key, item.Locale, item.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, nil
}

func (s *TranslationService) Delete(ctx context.Context, key, locale string) error {
	return s.repo.Delete(ctx, key, i18n.NormalizeLocale(locale))
}
