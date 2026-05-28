package service

import (
	"context"
	"errors"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var (
	ErrCannotDeleteSelf       = errors.New("cannot delete your own account")
	ErrCannotModifySuperadmin = errors.New("cannot modify superadmin account")
	ErrCannotImpersonateAdmin = errors.New("cannot impersonate superadmin")
)

type AdminUserService struct {
	users    *repository.UserRepository
	plans    *repository.PlanRepository
	auth     *middleware.Auth
	telegram *TelegramService
}

func NewAdminUserService(users *repository.UserRepository, plans *repository.PlanRepository, auth *middleware.Auth, telegram *TelegramService) *AdminUserService {
	return &AdminUserService{users: users, plans: plans, auth: auth, telegram: telegram}
}

type AdminUserList struct {
	Items  []repository.AdminUserRow `json:"items"`
	Total  int                       `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
}

func (s *AdminUserService) List(ctx context.Context, search string, limit, offset int) (*AdminUserList, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.users.ListAdmin(ctx, search, limit, offset)
	if err != nil {
		return nil, err
	}
	total, err := s.users.CountAdmin(ctx, search)
	if err != nil {
		return nil, err
	}
	return &AdminUserList{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *AdminUserService) UpdatePlan(ctx context.Context, actorID, targetID, planID string) (*repository.AdminUserRow, error) {
	target, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if target.Role == "superadmin" && target.ID != actorID {
		return nil, ErrCannotModifySuperadmin
	}
	newPlan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	oldPlanID := target.PlanID
	if err := s.users.UpdatePlanID(ctx, targetID, planID); err != nil {
		return nil, err
	}
	if s.telegram != nil && newPlan.PriceMonthlyRUB > 0 && planChanged(oldPlanID, planID) {
		s.telegram.NotifyPayment(ctx, target, newPlan, newPlan.PriceMonthlyRUB)
	}
	return s.users.GetAdminRow(ctx, targetID)
}

func planChanged(oldPlanID *string, newPlanID string) bool {
	if oldPlanID == nil {
		return true
	}
	return *oldPlanID != newPlanID
}

func (s *AdminUserService) Delete(ctx context.Context, actorID, targetID string) error {
	target, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		return err
	}
	if target.ID == actorID {
		return ErrCannotDeleteSelf
	}
	if target.Role == "superadmin" {
		return ErrCannotModifySuperadmin
	}
	return s.users.Delete(ctx, targetID)
}

func (s *AdminUserService) Impersonate(ctx context.Context, actorID, targetID string) (*AuthResult, error) {
	if actorID == targetID {
		return nil, ErrInvalidInput
	}
	target, err := s.users.GetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if target.Role == "superadmin" {
		return nil, ErrCannotImpersonateAdmin
	}
	if target.IsBlocked {
		return nil, ErrUserBlocked
	}

	token, err := s.auth.IssueToken(target.ID, target.Role, tokenTTL)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: target}, nil
}
