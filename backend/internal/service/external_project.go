package service

import (
	"context"
	"net/url"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

type ExternalProjectService struct {
	repo *repository.ExternalProjectRepository
}

func NewExternalProjectService(repo *repository.ExternalProjectRepository) *ExternalProjectService {
	return &ExternalProjectService{repo: repo}
}

type ExternalProjectInput struct {
	Title     string
	URL       string
	SortOrder int
	IsEnabled bool
}

func (s *ExternalProjectService) ListPublic(ctx context.Context) ([]model.ExternalProject, error) {
	return s.repo.ListEnabled(ctx)
}

func (s *ExternalProjectService) ListAdmin(ctx context.Context) ([]model.ExternalProject, error) {
	return s.repo.ListAll(ctx)
}

func (s *ExternalProjectService) Create(ctx context.Context, input ExternalProjectInput) (*model.ExternalProject, error) {
	if err := validateExternalProjectInput(input); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, input.Title, input.URL, input.SortOrder, input.IsEnabled)
}

func (s *ExternalProjectService) Update(ctx context.Context, id string, input ExternalProjectInput) (*model.ExternalProject, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidInput
	}
	if err := validateExternalProjectInput(input); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input.Title, input.URL, input.SortOrder, input.IsEnabled)
}

func (s *ExternalProjectService) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}

func validateExternalProjectInput(input ExternalProjectInput) error {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return ErrInvalidInput
	}
	rawURL := strings.TrimSpace(input.URL)
	if rawURL == "" {
		return ErrInvalidInput
	}
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return ErrInvalidInput
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return ErrInvalidInput
	}
	return nil
}
