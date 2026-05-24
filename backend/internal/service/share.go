package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

type ShareService struct {
	shared *repository.SharedReportRepository
	runs   *repository.ToolRunRepository
	billing *BillingService
}

func NewShareService(shared *repository.SharedReportRepository, runs *repository.ToolRunRepository, billing *BillingService) *ShareService {
	return &ShareService{shared: shared, runs: runs, billing: billing}
}

type ShareResult struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	URL       string    `json:"url"`
}

func (s *ShareService) CreateShare(ctx context.Context, userID, runID, baseURL string, user *model.User) (*ShareResult, error) {
	if user.AccountSegment != model.AccountSegmentPartner {
		return nil, ErrPartnerOnly
	}
	if err := s.billing.CanShareReport(ctx, userID); err != nil {
		return nil, err
	}

	run, err := s.runs.GetByIDForUser(ctx, runID, userID)
	if err != nil {
		return nil, err
	}
	if run.ToolSlug != "calculator" || run.Status != model.RunStatusDone {
		return nil, ErrInvalidInput
	}

	token, err := s.generateShareToken(ctx)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour)
	sr, err := s.shared.Create(ctx, runID, userID, token, expiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.runs.IncrementUsageCounter(ctx, userID, "share_report"); err != nil {
		return nil, err
	}

	url := baseURL + "/dashboard/share/" + sr.Token
	return &ShareResult{Token: sr.Token, ExpiresAt: sr.ExpiresAt, URL: url}, nil
}

type PublicSharePayload struct {
	Run             *model.ToolRun
	ShowPlatformCta bool
}

func (s *ShareService) GetPublicReport(ctx context.Context, token string) (*PublicSharePayload, error) {
	sr, err := s.shared.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if sr.RevokedAt != nil {
		return nil, repository.ErrNotFound
	}
	if time.Now().After(sr.ExpiresAt) {
		return nil, repository.ErrNotFound
	}

	run, err := s.runs.GetByID(ctx, sr.RunID)
	if err != nil {
		return nil, err
	}

	showPlatformCta := true
	if up, err := s.billing.GetUserPlan(ctx, sr.UserID); err == nil {
		showPlatformCta = !s.billing.HasFeature(&up.Plan, "white_label")
	}

	_ = s.shared.IncrementView(ctx, sr.ID)
	return &PublicSharePayload{Run: run, ShowPlatformCta: showPlatformCta}, nil
}

func (s *ShareService) generateShareToken(ctx context.Context) (string, error) {
	for range 8 {
		token, err := randomShareID()
		if err != nil {
			return "", err
		}
		exists, err := s.shared.TokenExists(ctx, token)
		if err != nil {
			return "", err
		}
		if !exists {
			return token, nil
		}
	}
	return "", errors.New("failed to generate unique share token")
}

func randomShareID() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10_000_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("id%010d", n.Int64()), nil
}

type LeadService struct {
	leads *repository.LeadRepository
	runs  *repository.ToolRunRepository
}

func NewLeadService(leads *repository.LeadRepository, runs *repository.ToolRunRepository) *LeadService {
	return &LeadService{leads: leads, runs: runs}
}

func (s *LeadService) Create(ctx context.Context, user *model.User, runID, name, email string, company, phone, message *string) (*model.Lead, error) {
	if user.AccountSegment != model.AccountSegmentDirectLead {
		return nil, errors.New("lead capture for direct clients only")
	}
	if name == "" || email == "" {
		return nil, ErrInvalidInput
	}
	if _, err := s.runs.GetByIDForUser(ctx, runID, user.ID); err != nil {
		return nil, err
	}
	return s.leads.Create(ctx, user.ID, runID, name, email, company, phone, message)
}
