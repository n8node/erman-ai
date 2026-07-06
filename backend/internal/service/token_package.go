package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidTokenPackage = errors.New("invalid token package")

type TokenPackageService struct {
	packages  *repository.TokenPackageRepository
	checkouts *repository.TokenPackageCheckoutRepository
	balances  *repository.UserTokenBalanceRepository
	users     *repository.UserRepository
	payments  *PaymentSettingsService
	checkout  *CheckoutService
}

func NewTokenPackageService(
	packages *repository.TokenPackageRepository,
	checkouts *repository.TokenPackageCheckoutRepository,
	balances *repository.UserTokenBalanceRepository,
	users *repository.UserRepository,
	payments *PaymentSettingsService,
) *TokenPackageService {
	return &TokenPackageService{
		packages:  packages,
		checkouts: checkouts,
		balances:  balances,
		users:     users,
		payments:  payments,
	}
}

func (s *TokenPackageService) SetCheckoutService(checkout *CheckoutService) {
	s.checkout = checkout
}

func (s *TokenPackageService) ListPublic(ctx context.Context) ([]model.TokenPackage, error) {
	return s.packages.ListPublic(ctx)
}

func (s *TokenPackageService) ListAdmin(ctx context.Context) ([]model.TokenPackage, error) {
	return s.packages.ListAdmin(ctx)
}

func (s *TokenPackageService) CreateAdmin(ctx context.Context, in model.TokenPackageUpsertInput) (*model.TokenPackage, error) {
	if err := validateTokenPackageInput(in, true); err != nil {
		return nil, err
	}
	return s.packages.Create(ctx, in)
}

func (s *TokenPackageService) UpdateAdmin(ctx context.Context, id string, in model.TokenPackageUpsertInput) (*model.TokenPackage, error) {
	if err := validateTokenPackageInput(in, false); err != nil {
		return nil, err
	}
	return s.packages.Update(ctx, id, in)
}

func (s *TokenPackageService) GetUserBalance(ctx context.Context, userID string) (int64, error) {
	return s.balances.GetBalance(ctx, userID)
}

func (s *TokenPackageService) CreateCheckout(ctx context.Context, userID, packageID string) (*model.CheckoutResult, error) {
	if s.checkout == nil {
		return nil, ErrCheckoutUnavailable
	}
	enabled, provider := s.checkout.PaymentsEnabled(ctx)
	if !enabled {
		return nil, ErrCheckoutUnavailable
	}

	pkg, err := s.packages.GetByID(ctx, packageID)
	if err != nil {
		return nil, err
	}
	if !pkg.IsPublic || pkg.IsArchived {
		return nil, ErrInvalidTokenPackage
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	checkout, err := s.checkouts.Create(ctx, userID, packageID, string(provider), pkg.PriceRUB, pkg.Tokens)
	if err != nil {
		return nil, err
	}

	cfg, err := s.payments.GetEffective(ctx)
	if err != nil {
		return nil, err
	}

	switch provider {
	case model.PaymentProviderYookassa:
		return s.checkout.createYookassaTokenPackage(ctx, cfg, checkout, pkg, user)
	case model.PaymentProviderRobokassa:
		return s.checkout.createRobokassaTokenPackage(ctx, cfg, checkout, pkg)
	default:
		return nil, ErrCheckoutUnavailable
	}
}

func (s *TokenPackageService) HandleYookassaWebhook(ctx context.Context, event, paymentID, status string, metadata map[string]string) error {
	if event != "payment.succeeded" || status != "succeeded" {
		return nil
	}
	checkoutID := strings.TrimSpace(metadata["token_checkout_id"])
	if checkoutID == "" && paymentID != "" {
		c, err := s.checkouts.GetByExternal(ctx, string(model.PaymentProviderYookassa), paymentID)
		if err == nil {
			checkoutID = c.ID
		}
	}
	if checkoutID == "" {
		return repository.ErrNotFound
	}
	return s.Fulfill(ctx, checkoutID)
}

func (s *TokenPackageService) HandleRobokassaResult(ctx context.Context, invIDStr, outSumStr string) error {
	invID, err := strconv.ParseInt(strings.TrimSpace(invIDStr), 10, 64)
	if err != nil {
		return ErrInvalidInput
	}
	checkout, err := s.checkouts.GetByInvID(ctx, invID)
	if err != nil {
		return err
	}
	if err := verifyRobokassaAmount(checkout.AmountRUB, outSumStr); err != nil {
		return err
	}
	return s.Fulfill(ctx, checkout.ID)
}

func (s *TokenPackageService) Fulfill(ctx context.Context, checkoutID string) error {
	checkout, err := s.checkouts.GetByID(ctx, checkoutID)
	if err != nil {
		return err
	}
	if checkout.Status == model.CheckoutStatusPaid {
		return nil
	}
	if checkout.Status != model.CheckoutStatusPending {
		return ErrInvalidInput
	}

	pkg, err := s.packages.GetByID(ctx, checkout.TokenPackageID)
	if err != nil {
		return err
	}
	if pkg.PriceRUB != checkout.AmountRUB || pkg.Tokens != checkout.TokensAmount {
		return ErrInvalidInput
	}

	paid, err := s.checkouts.MarkPaid(ctx, checkoutID)
	if err != nil {
		return err
	}
	if paid.Status != model.CheckoutStatusPaid {
		return nil
	}

	_, err = s.balances.AddTokens(ctx, checkout.UserID, checkout.TokensAmount)
	return err
}

func validateTokenPackageInput(in model.TokenPackageUpsertInput, creating bool) error {
	if creating && strings.TrimSpace(in.Slug) == "" {
		return fmt.Errorf("%w: slug required", ErrInvalidTokenPackage)
	}
	if strings.TrimSpace(in.Name) == "" {
		return fmt.Errorf("%w: name required", ErrInvalidTokenPackage)
	}
	if in.Tokens <= 0 {
		return fmt.Errorf("%w: tokens must be positive", ErrInvalidTokenPackage)
	}
	if in.PriceRUB <= 0 {
		return fmt.Errorf("%w: price must be positive", ErrInvalidTokenPackage)
	}
	return nil
}
