package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var (
	ErrConsultationSlotUnavailable = errors.New("consultation slot unavailable")
	ErrConsultationPaymentsOff     = errors.New("consultation payments disabled")
)

type ConsultationService struct {
	repo     *repository.ConsultationRepository
	users    *repository.UserRepository
	payments *PaymentSettingsService
	telegram *TelegramService
	cfg      *config.Config
}

type ConsultationBookingInput struct {
	ServiceID        string
	StartsAt         time.Time
	CustomerName     string
	CustomerEmail    string
	CustomerPhone    string
	CustomerTelegram string
	CustomerNote     string
	Timezone         string
}

type ConsultationAdminServiceInput = repository.ConsultationServiceInput
type ConsultationAdminAvailabilityInput = repository.ConsultationAvailabilityInput

func NewConsultationService(
	repo *repository.ConsultationRepository,
	users *repository.UserRepository,
	payments *PaymentSettingsService,
	telegram *TelegramService,
	cfg *config.Config,
) *ConsultationService {
	return &ConsultationService{repo: repo, users: users, payments: payments, telegram: telegram, cfg: cfg}
}

func (s *ConsultationService) ListPublicServices(ctx context.Context) ([]model.ConsultationService, error) {
	return s.repo.ListServices(ctx, false)
}

func (s *ConsultationService) ListAdminServices(ctx context.Context) ([]model.ConsultationService, error) {
	return s.repo.ListServices(ctx, true)
}

func (s *ConsultationService) CreateAdminService(ctx context.Context, in ConsultationAdminServiceInput) (*model.ConsultationService, error) {
	if err := normalizeConsultationServiceInput(&in); err != nil {
		return nil, err
	}
	return s.repo.CreateService(ctx, in)
}

func (s *ConsultationService) UpdateAdminService(ctx context.Context, id string, in ConsultationAdminServiceInput) (*model.ConsultationService, error) {
	if err := normalizeConsultationServiceInput(&in); err != nil {
		return nil, err
	}
	return s.repo.UpdateService(ctx, id, in)
}

func (s *ConsultationService) ListAvailability(ctx context.Context, serviceID string) ([]model.ConsultationAvailabilityRule, error) {
	return s.repo.ListAvailability(ctx, serviceID)
}

func (s *ConsultationService) ReplaceAvailability(ctx context.Context, serviceID string, rules []ConsultationAdminAvailabilityInput) ([]model.ConsultationAvailabilityRule, error) {
	for i := range rules {
		if rules[i].Weekday < 0 || rules[i].Weekday > 6 || strings.TrimSpace(rules[i].StartTime) == "" || strings.TrimSpace(rules[i].EndTime) == "" {
			return nil, ErrInvalidInput
		}
		if rules[i].SlotStepMinutes <= 0 {
			rules[i].SlotStepMinutes = 15
		}
	}
	return s.repo.ReplaceAvailability(ctx, serviceID, rules)
}

func (s *ConsultationService) ListSlots(ctx context.Context, serviceID string, from, to time.Time) ([]model.ConsultationSlot, error) {
	svc, err := s.repo.GetService(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	if !svc.IsActive {
		return []model.ConsultationSlot{}, nil
	}
	if to.IsZero() || to.Before(from) {
		to = from.AddDate(0, 0, svc.MaxAdvanceDays)
	}
	loc := consultationLocation("Europe/Moscow")
	fromLocal := from.In(loc)
	toLocal := to.In(loc)
	maxTo := time.Now().In(loc).AddDate(0, 0, svc.MaxAdvanceDays)
	if toLocal.After(maxTo) {
		toLocal = maxTo
	}

	rules, err := s.repo.ListAvailability(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	bookings, err := s.repo.ListBookingsForService(ctx, serviceID, fromLocal, toLocal)
	if err != nil {
		return nil, err
	}
	blackouts, err := s.repo.ListBlackouts(ctx, serviceID, fromLocal, toLocal)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)
	minStart := now.Add(time.Duration(svc.MinNoticeMinutes) * time.Minute)
	slots := []model.ConsultationSlot{}
	for day := startOfDay(fromLocal, loc); day.Before(toLocal); day = day.AddDate(0, 0, 1) {
		weekday := int(day.Weekday())
		for _, rule := range rules {
			if !rule.IsActive || rule.Weekday != weekday {
				continue
			}
			startClock, err := time.Parse("15:04:05", normalizeClock(rule.StartTime))
			if err != nil {
				continue
			}
			endClock, err := time.Parse("15:04:05", normalizeClock(rule.EndTime))
			if err != nil {
				continue
			}
			step := time.Duration(rule.SlotStepMinutes) * time.Minute
			for start := time.Date(day.Year(), day.Month(), day.Day(), startClock.Hour(), startClock.Minute(), 0, 0, loc); ; start = start.Add(step) {
				end := start.Add(time.Duration(svc.DurationMinutes) * time.Minute)
				ruleEnd := time.Date(day.Year(), day.Month(), day.Day(), endClock.Hour(), endClock.Minute(), 0, 0, loc)
				if end.After(ruleEnd) {
					break
				}
				if start.Before(fromLocal) || start.Before(minStart) {
					continue
				}
				available := !overlapsAnyBooking(start, end, bookings) && !overlapsAnyBlackout(start, end, blackouts)
				slots = append(slots, model.ConsultationSlot{StartsAt: start, EndsAt: end, Available: available})
			}
		}
	}
	return slots, nil
}

func (s *ConsultationService) CreateBooking(ctx context.Context, userID *string, input ConsultationBookingInput) (*model.ConsultationBooking, error) {
	input.CustomerName = strings.TrimSpace(input.CustomerName)
	input.CustomerEmail = strings.ToLower(strings.TrimSpace(input.CustomerEmail))
	input.CustomerPhone = strings.TrimSpace(input.CustomerPhone)
	input.CustomerTelegram = strings.TrimSpace(input.CustomerTelegram)
	input.CustomerNote = strings.TrimSpace(input.CustomerNote)
	if input.Timezone == "" {
		input.Timezone = "Europe/Moscow"
	}
	if input.ServiceID == "" || input.CustomerName == "" || input.CustomerEmail == "" || input.StartsAt.IsZero() {
		return nil, ErrInvalidInput
	}
	if !strings.Contains(input.CustomerEmail, "@") {
		return nil, ErrInvalidInput
	}
	if userID == nil && input.CustomerEmail != "" {
		if user, _, err := s.users.GetByEmail(ctx, input.CustomerEmail); err == nil && !user.IsBlocked {
			userID = &user.ID
		}
	}

	svc, err := s.repo.GetService(ctx, input.ServiceID)
	if err != nil {
		return nil, err
	}
	if !svc.IsActive || svc.PriceRUB <= 0 {
		return nil, ErrInvalidInput
	}
	end := input.StartsAt.Add(time.Duration(svc.DurationMinutes) * time.Minute)
	slots, err := s.ListSlots(ctx, svc.ID, input.StartsAt.Add(-time.Minute), end.Add(time.Minute))
	if err != nil {
		return nil, err
	}
	if !slotIsAvailable(slots, input.StartsAt) {
		return nil, ErrConsultationSlotUnavailable
	}

	return s.repo.CreateBooking(ctx, repository.ConsultationBookingInput{
		ServiceID:        svc.ID,
		UserID:           userID,
		CustomerName:     input.CustomerName,
		CustomerEmail:    input.CustomerEmail,
		CustomerPhone:    input.CustomerPhone,
		CustomerTelegram: input.CustomerTelegram,
		CustomerNote:     input.CustomerNote,
		StartsAt:         input.StartsAt,
		EndsAt:           end,
		Timezone:         input.Timezone,
		AmountRUB:        svc.PriceRUB,
		MeetingURL:       svc.MeetingURL,
		ExpiresAt:        time.Now().Add(15 * time.Minute),
	})
}

func (s *ConsultationService) CreateCheckout(ctx context.Context, bookingID string) (*model.CheckoutResult, error) {
	booking, err := s.repo.GetBookingDetail(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.Status == model.ConsultationBookingStatusPaid {
		return &model.CheckoutResult{CheckoutID: booking.ID, Provider: valueOrEmpty(booking.Provider)}, nil
	}
	if booking.Status != model.ConsultationBookingStatusPendingPayment || time.Now().After(booking.ExpiresAt) {
		return nil, ErrConsultationSlotUnavailable
	}

	enabled, provider, err := s.payments.PaymentsEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, ErrConsultationPaymentsOff
	}
	cfg, err := s.payments.GetEffective(ctx)
	if err != nil {
		return nil, err
	}
	switch provider {
	case model.PaymentProviderYookassa:
		return s.createYookassaCheckout(ctx, cfg, booking)
	case model.PaymentProviderRobokassa:
		return s.createRobokassaCheckout(ctx, cfg, booking)
	default:
		return nil, ErrCheckoutUnavailable
	}
}

func (s *ConsultationService) createYookassaCheckout(ctx context.Context, cfg model.PaymentSettings, booking *model.ConsultationBookingDetail) (*model.CheckoutResult, error) {
	returnURL := s.cfg.PublicBaseURL() + "/dashboard/consultations?booking=success"
	amountStr := formatRubAmount(booking.AmountRUB)
	body := map[string]any{
		"amount":       map[string]string{"value": amountStr, "currency": "RUB"},
		"capture":      true,
		"confirmation": map[string]string{"type": "redirect", "return_url": returnURL},
		"description":  fmt.Sprintf("Erman AI — %s", booking.ServiceName),
		"metadata": map[string]string{
			"kind":       "consultation",
			"booking_id": booking.ID,
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.yookassa.ru/v3/payments", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(strings.TrimSpace(cfg.Yookassa.ShopID), strings.TrimSpace(cfg.Yookassa.SecretKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", booking.ID)
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
	if err := s.repo.SetBookingExternal(ctx, booking.ID, string(model.PaymentProviderYookassa), payment.ID, nil); err != nil {
		return nil, err
	}
	return &model.CheckoutResult{CheckoutID: booking.ID, Provider: string(model.PaymentProviderYookassa), CheckoutURL: payment.Confirmation.ConfirmationURL}, nil
}

func (s *ConsultationService) createRobokassaCheckout(ctx context.Context, cfg model.PaymentSettings, booking *model.ConsultationBookingDetail) (*model.CheckoutResult, error) {
	invID, err := s.repo.NextInvID(ctx)
	if err != nil {
		return nil, err
	}
	outSum := formatRubAmount(booking.AmountRUB)
	login := strings.TrimSpace(cfg.Robokassa.MerchantLogin)
	pass1 := strings.TrimSpace(cfg.Robokassa.Password1)
	invStr := strconv.FormatInt(invID, 10)
	signature := BuildRobokassaPaymentSignature(login, outSum, invStr, pass1, "")

	returnURL := s.cfg.PublicBaseURL() + "/dashboard/consultations"
	params := url.Values{}
	params.Set("MerchantLogin", login)
	params.Set("OutSum", outSum)
	params.Set("InvId", invStr)
	params.Set("Description", fmt.Sprintf("Erman AI — %s", booking.ServiceName))
	params.Set("SignatureValue", signature)
	params.Set("SuccessURL", appendQuery(returnURL, "booking", "success"))
	params.Set("FailURL", appendQuery(returnURL, "booking", "failed"))
	if cfg.Robokassa.TestMode {
		params.Set("IsTest", "1")
	}
	if err := s.repo.SetBookingExternal(ctx, booking.ID, string(model.PaymentProviderRobokassa), invStr, &invID); err != nil {
		return nil, err
	}
	return &model.CheckoutResult{
		CheckoutID:  booking.ID,
		Provider:    string(model.PaymentProviderRobokassa),
		CheckoutURL: "https://auth.robokassa.ru/Merchant/Index.aspx?" + params.Encode(),
	}, nil
}

func (s *ConsultationService) HandleYookassaWebhook(ctx context.Context, event, paymentID, status string, metadata map[string]string) error {
	if event != "payment.succeeded" || status != "succeeded" {
		return nil
	}
	bookingID := strings.TrimSpace(metadata["booking_id"])
	if bookingID == "" && paymentID != "" {
		booking, err := s.repo.GetBookingByExternal(ctx, string(model.PaymentProviderYookassa), paymentID)
		if err == nil {
			bookingID = booking.ID
		}
	}
	if bookingID == "" {
		return repository.ErrNotFound
	}
	return s.Fulfill(ctx, bookingID)
}

func (s *ConsultationService) HandleRobokassaResult(ctx context.Context, invIDStr, outSumStr string) error {
	invID, err := strconv.ParseInt(strings.TrimSpace(invIDStr), 10, 64)
	if err != nil {
		return ErrInvalidInput
	}
	booking, err := s.repo.GetBookingByInvID(ctx, invID)
	if err != nil {
		return err
	}
	got, err := strconv.ParseFloat(strings.TrimSpace(outSumStr), 64)
	if err != nil {
		return ErrInvalidInput
	}
	if math.Abs(got-float64(booking.AmountRUB)) > 0.01 {
		return ErrInvalidInput
	}
	return s.Fulfill(ctx, booking.ID)
}

func (s *ConsultationService) Fulfill(ctx context.Context, bookingID string) error {
	booking, err := s.repo.GetBookingDetail(ctx, bookingID)
	if err != nil {
		return err
	}
	if booking.Status == model.ConsultationBookingStatusPaid {
		return nil
	}
	if booking.Status != model.ConsultationBookingStatusPendingPayment {
		return ErrInvalidInput
	}
	paid, err := s.repo.MarkBookingPaid(ctx, bookingID)
	if err != nil {
		return err
	}
	if s.telegram != nil {
		s.telegram.NotifyConsultationBooking(ctx, paid, s.cfg.PublicBaseURL())
	}
	return nil
}

func (s *ConsultationService) ListBookingsAdmin(ctx context.Context, limit, offset int) ([]model.ConsultationBookingDetail, int, error) {
	return s.repo.ListBookingsAdmin(ctx, limit, offset)
}

func (s *ConsultationService) UpdateBookingStatus(ctx context.Context, id, status string) (*model.ConsultationBookingDetail, error) {
	switch status {
	case model.ConsultationBookingStatusPendingPayment, model.ConsultationBookingStatusPaid, model.ConsultationBookingStatusCancelled, model.ConsultationBookingStatusExpired, model.ConsultationBookingStatusCompleted, model.ConsultationBookingStatusNoShow:
	default:
		return nil, ErrInvalidInput
	}
	return s.repo.UpdateBookingStatus(ctx, id, status)
}

func normalizeConsultationServiceInput(in *ConsultationAdminServiceInput) error {
	in.Slug = strings.TrimSpace(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.MeetingURL = strings.TrimSpace(in.MeetingURL)
	if in.Slug == "" {
		in.Slug = slugify(in.Name)
	}
	if in.Name == "" || in.Slug == "" || in.DurationMinutes <= 0 || in.PriceRUB < 0 {
		return ErrInvalidInput
	}
	if in.MinNoticeMinutes < 0 {
		in.MinNoticeMinutes = 180
	}
	if in.MaxAdvanceDays <= 0 {
		in.MaxAdvanceDays = 30
	}
	if in.SortOrder == 0 {
		in.SortOrder = 100
	}
	return nil
}

func slugify(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = regexp.MustCompile(`[^a-z0-9а-яё]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "consultation"
	}
	return slug
}

func consultationLocation(name string) *time.Location {
	loc, err := time.LoadLocation(strings.TrimSpace(name))
	if err != nil {
		return time.FixedZone("Europe/Moscow", 3*60*60)
	}
	return loc
}

func startOfDay(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func normalizeClock(value string) string {
	value = strings.TrimSpace(value)
	if len(value) == 5 {
		return value + ":00"
	}
	return value
}

func overlapsAnyBooking(start, end time.Time, bookings []model.ConsultationBooking) bool {
	for _, b := range bookings {
		if start.Before(b.EndsAt) && end.After(b.StartsAt) {
			return true
		}
	}
	return false
}

func overlapsAnyBlackout(start, end time.Time, blackouts []model.ConsultationBlackout) bool {
	for _, b := range blackouts {
		if start.Before(b.EndsAt) && end.After(b.StartsAt) {
			return true
		}
	}
	return false
}

func slotIsAvailable(slots []model.ConsultationSlot, startsAt time.Time) bool {
	for _, slot := range slots {
		if slot.Available && slot.StartsAt.Equal(startsAt) {
			return true
		}
	}
	return false
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
