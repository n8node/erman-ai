package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var (
	ErrPublicPageSlugTaken   = errors.New("public page slug already taken")
	ErrPublicPageSlugInvalid = errors.New("public page slug invalid")
)

var publicPageSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var reservedPublicPageSlugs = map[string]struct{}{
	"login": {}, "register": {}, "verify-email": {}, "onboarding": {},
	"share": {}, "tools": {}, "admin": {}, "billing": {}, "settings": {},
	"api-keys": {}, "discuss": {}, "api": {},
}

var allowedPublicPageTemplates = map[string]struct{}{
	"calculator-landing": {},
	"strategy-landing":   {},
	"proposal-landing":   {},
}

type PublicPageService struct {
	repo *repository.PublicPageRepository
}

func NewPublicPageService(repo *repository.PublicPageRepository) *PublicPageService {
	return &PublicPageService{repo: repo}
}

type PublicPageInput struct {
	Slug            string
	Title           string
	Template        string
	ContentHTML     string
	MetaDescription string
	IsPublished     bool
	SortOrder       int
}

func (s *PublicPageService) ListPublishedSlugs(ctx context.Context) ([]string, error) {
	return s.repo.ListPublishedSlugs(ctx)
}

func (s *PublicPageService) GetPublishedBySlug(ctx context.Context, slug string) (*model.PublicPage, error) {
	slug = normalizePublicPageSlug(slug)
	if slug == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.GetPublishedBySlug(ctx, slug)
}

func (s *PublicPageService) ListAdmin(ctx context.Context) ([]model.PublicPage, error) {
	return s.repo.ListAll(ctx)
}

func (s *PublicPageService) Create(ctx context.Context, input PublicPageInput) (*model.PublicPage, error) {
	if err := s.validateInput(ctx, input, ""); err != nil {
		return nil, err
	}
	template := normalizePublicPageTemplate(input.Template)
	return s.repo.Create(
		ctx,
		normalizePublicPageSlug(input.Slug),
		strings.TrimSpace(input.Title),
		template,
		input.ContentHTML,
		strings.TrimSpace(input.MetaDescription),
		input.IsPublished,
		input.SortOrder,
	)
}

func (s *PublicPageService) Update(ctx context.Context, id string, input PublicPageInput) (*model.PublicPage, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidInput
	}
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.validateInput(ctx, input, id); err != nil {
		return nil, err
	}

	template := existing.Template
	if template == "" {
		template = normalizePublicPageTemplate(input.Template)
	}

	contentHTML := input.ContentHTML
	metaDescription := strings.TrimSpace(input.MetaDescription)
	if template != "" {
		contentHTML = existing.ContentHTML
		metaDescription = existing.MetaDescription
	}

	return s.repo.Update(
		ctx,
		id,
		normalizePublicPageSlug(input.Slug),
		strings.TrimSpace(input.Title),
		template,
		contentHTML,
		metaDescription,
		input.IsPublished,
		input.SortOrder,
	)
}

func (s *PublicPageService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}

func (s *PublicPageService) validateInput(ctx context.Context, input PublicPageInput, excludeID string) error {
	slug := normalizePublicPageSlug(input.Slug)
	if slug == "" || !publicPageSlugPattern.MatchString(slug) {
		return ErrPublicPageSlugInvalid
	}
	if _, reserved := reservedPublicPageSlugs[slug]; reserved {
		return ErrPublicPageSlugInvalid
	}
	if strings.TrimSpace(input.Title) == "" {
		return ErrInvalidInput
	}
	template := normalizePublicPageTemplate(input.Template)
	if template != "" {
		if _, ok := allowedPublicPageTemplates[template]; !ok {
			return ErrInvalidInput
		}
	}
	exists, err := s.repo.SlugExists(ctx, slug, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return ErrPublicPageSlugTaken
	}
	return nil
}

func normalizePublicPageSlug(slug string) string {
	return strings.Trim(strings.ToLower(strings.TrimSpace(slug)), "/")
}

func normalizePublicPageTemplate(template string) string {
	return strings.TrimSpace(template)
}
