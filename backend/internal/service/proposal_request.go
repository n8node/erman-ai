package service

import (
	"context"
	"errors"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrProposalRequestPartnerOnly = errors.New("proposal requests for partners only")

type ProposalRequestList struct {
	Items  []repository.AdminProposalRequestRow `json:"items"`
	Total  int                                  `json:"total"`
	Limit  int                                  `json:"limit"`
	Offset int                                  `json:"offset"`
}

type ProposalRequestService struct {
	requests *repository.ProposalRequestRepository
	runs     *repository.ToolRunRepository
	users    *repository.UserRepository
}

func NewProposalRequestService(
	requests *repository.ProposalRequestRepository,
	runs *repository.ToolRunRepository,
	users *repository.UserRepository,
) *ProposalRequestService {
	return &ProposalRequestService{requests: requests, runs: runs, users: users}
}

func (s *ProposalRequestService) Create(
	ctx context.Context,
	userID, runID, requesterName, telegram, businessNote string,
) (*model.ProposalRequest, error) {
	requesterName = strings.TrimSpace(requesterName)
	telegram = strings.TrimSpace(telegram)
	businessNote = strings.TrimSpace(businessNote)
	if requesterName == "" || telegram == "" || businessNote == "" {
		return nil, ErrInvalidInput
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.AccountSegment != model.AccountSegmentPartner {
		return nil, ErrProposalRequestPartnerOnly
	}

	run, err := s.runs.GetByIDForUser(ctx, runID, userID)
	if err != nil {
		return nil, err
	}
	if run.ToolSlug != "calculator" {
		return nil, ErrInvalidInput
	}

	return s.requests.Create(ctx, userID, runID, requesterName, telegram, businessNote)
}

func (s *ProposalRequestService) ListAdmin(ctx context.Context, limit, offset int) (*ProposalRequestList, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.requests.ListAdmin(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.requests.CountAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &ProposalRequestList{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *ProposalRequestService) GetAdminDetail(ctx context.Context, id string) (*repository.AdminProposalRequestDetail, error) {
	return s.requests.GetAdminDetail(ctx, id)
}

func (s *ProposalRequestService) UpdateStatus(ctx context.Context, id, status string) (*model.ProposalRequest, error) {
	switch model.ProposalRequestStatus(status) {
	case model.ProposalRequestStatusNew,
		model.ProposalRequestStatusInProgress,
		model.ProposalRequestStatusDone:
	default:
		return nil, ErrInvalidInput
	}
	return s.requests.UpdateStatus(ctx, id, status)
}
