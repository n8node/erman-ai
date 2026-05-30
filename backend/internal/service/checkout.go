package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var (
	ErrCheckoutUnavailable = errors.New("checkout unavailable")
	ErrCheckoutNotFound    = errors.New("checkout not found")
)

type CheckoutService struct {
	checkouts *repository.PlanCheckoutRepository
	users     *repository.UserRepository
	plans     *repository.PlanRepository
	payments  *PaymentSettingsService
	telegram  *TelegramService
	cfg       *config.Config
}

func NewCheckoutService(
	checkouts *repository.PlanCheckoutRepository,
	users *repository.UserRepository,
	plans *repository.PlanRepository,
	payments *PaymentSettingsService,
	telegram *TelegramService,
	cfg *config.Config,
) *CheckoutService {
	return &CheckoutService{
		checkouts: checkouts,
		users:     users,
		plans:     plans,
		payments:  payments,
		telegram:  telegram,
		cfg:       cfg,
	}
}

func (s *CheckoutService) PaymentsEnabled(ctx context.Context) (bool, model.PaymentProvider) {
	enabled, provider, err := s.payments.PaymentsEnabled(ctx)
	if err != nil {
		return false, ""
	}
	return enabled, provider
}

func (s *CheckoutService) Create(ctx context.Context, userID, planID string) (*model.CheckoutResult, error) {
	enabled, provider := s.PaymentsEnabled(ctx)
	if !enabled {
		return nil, ErrCheckoutUnavailable
	}

	plan, err := s.plans.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	if !plan.IsPublic || plan.IsArchived {
		return nil, ErrInvalidInput
	}
	if plan.PriceMonthlyRUB <= 0 {
		return nil, ErrInvalidInput
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.IsBlocked {
		return nil, ErrUserBlocked
	}

	checkout, err := s.checkouts.Create(ctx, userID, planID, string(provider), plan.PriceMonthlyRUB)
	if err != nil {
		return nil, err
	}

	cfg, err := s.payments.GetEffective(ctx)
	if err != nil {
		return nil, err
	}

	switch provider {
	case model.PaymentProviderYookassa:
		return s.createYookassa(ctx, cfg, checkout, plan, user)
	case model.PaymentProviderRobokassa:
		return s.createRobokassa(ctx, cfg, checkout, plan)
	default:
		return nil, ErrCheckoutUnavailable
	}
}

func (s *CheckoutService) createYookassa(ctx context.Context, cfg model.PaymentSettings, checkout *model.PlanCheckout, plan *model.Plan, user *model.User) (*model.CheckoutResult, error) {
	yk := cfg.Yookassa
	returnURL := strings.TrimSpace(yk.ReturnURL)
	if returnURL == "" {
		returnURL = s.payments.defaultReturnURL()
	}
	returnURL = appendQuery(returnURL, "payment", "success")

	amountStr := formatRubAmount(plan.PriceMonthlyRUB)
	desc := fmt.Sprintf("Erman AI — %s", plan.Name)

	body := map[string]any{
		"amount": map[string]string{
			"value":    amountStr,
			"currency": "RUB",
		},
		"capture": true,
		"confirmation": map[string]string{
			"type":       "redirect",
			"return_url": returnURL,
		},
		"description": desc,
		"metadata": map[string]string{
			"checkout_id": checkout.ID,
			"user_id":     user.ID,
			"plan_id":     plan.ID,
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.yookassa.ru/v3/payments", strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(strings.TrimSpace(yk.ShopID), strings.TrimSpace(yk.SecretKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", checkout.ID)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yookassa request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("yookassa status %d: %s", resp.StatusCode, string(respBody))
	}

	var payment struct {
		ID           string `json:"id"`
		Confirmation struct {
			ConfirmationURL string `json:"confirmation_url"`
		} `json:"confirmation"`
	}
	if err := json.Unmarshal(respBody, &payment); err != nil {
		return nil, err
	}
	if payment.ID == "" || payment.Confirmation.ConfirmationURL == "" {
		return nil, fmt.Errorf("yookassa: missing payment id or confirmation url")
	}

	if err := s.checkouts.SetExternal(ctx, checkout.ID, payment.ID, nil); err != nil {
		return nil, err
	}

	return &model.CheckoutResult{
		CheckoutID:  checkout.ID,
		Provider:    string(model.PaymentProviderYookassa),
		CheckoutURL: payment.Confirmation.ConfirmationURL,
	}, nil
}

func (s *CheckoutService) createRobokassa(ctx context.Context, cfg model.PaymentSettings, checkout *model.PlanCheckout, plan *model.Plan) (*model.CheckoutResult, error) {
	rk := cfg.Robokassa
	invID, err := s.checkouts.NextInvID(ctx)
	if err != nil {
		return nil, err
	}

	outSum := formatRubAmount(plan.PriceMonthlyRUB)
	login := strings.TrimSpace(rk.MerchantLogin)
	pass1 := strings.TrimSpace(rk.Password1)
	invStr := strconv.FormatInt(invID, 10)

	signature := fmt.Sprintf("%x", md5.Sum([]byte(login+":"+outSum+":"+invStr+":"+pass1)))

	returnURL := s.payments.defaultReturnURL()
	successURL := appendQuery(returnURL, "payment", "success")
	failURL := appendQuery(returnURL, "payment", "failed")

	params := url.Values{}
	params.Set("MerchantLogin", login)
	params.Set("OutSum", outSum)
	params.Set("InvId", invStr)
	params.Set("Description", fmt.Sprintf("Erman AI — %s", plan.Name))
	params.Set("SignatureValue", signature)
	params.Set("SuccessURL", successURL)
	params.Set("FailURL", failURL)
	if rk.TestMode {
		params.Set("IsTest", "1")
	}

	if err := s.checkouts.SetExternal(ctx, checkout.ID, invStr, &invID); err != nil {
		return nil, err
	}

	checkoutURL := "https://auth.robokassa.ru/Merchant/Index.aspx?" + params.Encode()
	return &model.CheckoutResult{
		CheckoutID:  checkout.ID,
		Provider:    string(model.PaymentProviderRobokassa),
		CheckoutURL: checkoutURL,
	}, nil
}

func (s *CheckoutService) HandleYookassaWebhook(ctx context.Context, event, paymentID, status string, metadata map[string]string) error {
	if event != "payment.succeeded" || status != "succeeded" {
		return nil
	}
	checkoutID := strings.TrimSpace(metadata["checkout_id"])
	if checkoutID == "" && paymentID != "" {
		c, err := s.checkouts.GetByExternal(ctx, string(model.PaymentProviderYookassa), paymentID)
		if err == nil {
			checkoutID = c.ID
		}
	}
	if checkoutID == "" {
		return ErrCheckoutNotFound
	}
	return s.Fulfill(ctx, checkoutID)
}

func (s *CheckoutService) HandleRobokassaResult(ctx context.Context, invIDStr, outSumStr string) error {
	invID, err := strconv.ParseInt(strings.TrimSpace(invIDStr), 10, 64)
	if err != nil {
		return ErrInvalidInput
	}
	checkout, err := s.checkouts.GetByInvID(ctx, invID)
	if err != nil {
		return err
	}
	if err := s.verifyRobokassaAmount(checkout.AmountRUB, outSumStr); err != nil {
		return err
	}
	return s.Fulfill(ctx, checkout.ID)
}

func (s *CheckoutService) Fulfill(ctx context.Context, checkoutID string) error {
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

	plan, err := s.plans.GetByID(ctx, checkout.PlanID)
	if err != nil {
		return err
	}
	if plan.PriceMonthlyRUB != checkout.AmountRUB {
		return ErrInvalidInput
	}

	paid, err := s.checkouts.MarkPaid(ctx, checkoutID)
	if err != nil {
		return err
	}
	if paid.Status != model.CheckoutStatusPaid {
		return nil
	}

	if err := s.users.UpdatePlanID(ctx, checkout.UserID, checkout.PlanID); err != nil {
		return err
	}
	_ = s.checkouts.CreateSubscription(ctx, checkout.UserID, checkout.PlanID, checkout.AmountRUB)

	if s.telegram != nil {
		user, err := s.users.GetByID(ctx, checkout.UserID)
		if err == nil {
			s.telegram.NotifyPayment(ctx, user, plan, checkout.AmountRUB)
		}
	}
	return nil
}

func (s *CheckoutService) verifyRobokassaAmount(expected int, outSumStr string) error {
	got, err := strconv.ParseFloat(strings.TrimSpace(outSumStr), 64)
	if err != nil {
		return ErrInvalidInput
	}
	if math.Abs(got-float64(expected)) > 0.01 {
		return ErrInvalidInput
	}
	return nil
}

func formatRubAmount(rub int) string {
	return fmt.Sprintf("%d.00", rub)
}

func appendQuery(rawURL, key, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}
