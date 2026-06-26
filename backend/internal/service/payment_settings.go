package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var ErrInvalidPaymentSettings = errors.New("invalid payment settings")

type PaymentSettingsService struct {
	repo *repository.PaymentSettingsRepository
	cfg  *config.Config
}

func NewPaymentSettingsService(repo *repository.PaymentSettingsRepository, cfg *config.Config) *PaymentSettingsService {
	return &PaymentSettingsService{repo: repo, cfg: cfg}
}

func (s *PaymentSettingsService) GetStored(ctx context.Context) (*model.PaymentSettingsRecord, error) {
	rec, err := s.repo.Get(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			def := model.DefaultPaymentSettings(s.defaultReturnURL())
			return &model.PaymentSettingsRecord{Config: def}, nil
		}
		return nil, err
	}
	return rec, nil
}

func (s *PaymentSettingsService) GetEffective(ctx context.Context) (model.PaymentSettings, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return model.PaymentSettings{}, err
	}
	return rec.Config, nil
}

func (s *PaymentSettingsService) GetAdminView(ctx context.Context) (*model.PaymentAdminView, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildAdminView(rec), nil
}

func (s *PaymentSettingsService) Update(ctx context.Context, req model.PaymentAdminUpdateRequest) (*model.PaymentAdminView, error) {
	if err := validatePaymentSettings(req); err != nil {
		return nil, err
	}

	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}

	cfg := rec.Config
	cfg.ActiveProvider = req.ActiveProvider
	cfg.Yookassa.ShopID = strings.TrimSpace(req.Yookassa.ShopID)
	cfg.Yookassa.ReturnURL = strings.TrimSpace(req.Yookassa.ReturnURL)
	cfg.Yookassa.Enabled = req.Yookassa.Enabled
	cfg.Robokassa.MerchantLogin = strings.TrimSpace(req.Robokassa.MerchantLogin)
	cfg.Robokassa.TestMode = req.Robokassa.TestMode
	cfg.Robokassa.Enabled = req.Robokassa.Enabled

	if strings.TrimSpace(req.YookassaSecretKey) != "" {
		cfg.Yookassa.SecretKey = strings.TrimSpace(req.YookassaSecretKey)
	}
	if strings.TrimSpace(req.RobokassaPassword1) != "" {
		cfg.Robokassa.Password1 = strings.TrimSpace(req.RobokassaPassword1)
	}
	if strings.TrimSpace(req.RobokassaPassword2) != "" {
		cfg.Robokassa.Password2 = strings.TrimSpace(req.RobokassaPassword2)
	}

	if cfg.Yookassa.ReturnURL == "" {
		cfg.Yookassa.ReturnURL = s.defaultReturnURL()
	}

	if cfg.ActiveProvider == model.PaymentProviderRobokassa && robokassaConfigured(cfg.Robokassa) {
		cfg.Robokassa.Enabled = true
	}

	updated, err := s.repo.Update(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return s.buildAdminView(updated), nil
}

func (s *PaymentSettingsService) TestConnection(ctx context.Context, provider model.PaymentProvider) (*model.PaymentTestResult, error) {
	switch provider {
	case model.PaymentProviderYookassa:
		return s.testYookassa(ctx)
	case model.PaymentProviderRobokassa:
		return s.testRobokassa(ctx)
	default:
		return nil, fmt.Errorf("%w: unknown provider", ErrInvalidPaymentSettings)
	}
}

func (s *PaymentSettingsService) testYookassa(ctx context.Context) (*model.PaymentTestResult, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	yk := rec.Config.Yookassa
	if strings.TrimSpace(yk.ShopID) == "" || strings.TrimSpace(yk.SecretKey) == "" {
		return &model.PaymentTestResult{
			OK:      false,
			Message: "Укажите Shop ID и секретный ключ API",
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.yookassa.ru/v3/me", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(strings.TrimSpace(yk.ShopID), strings.TrimSpace(yk.SecretKey))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &model.PaymentTestResult{
			OK:      false,
			Message: "Не удалось подключиться к API ЮKassa: " + err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == http.StatusOK {
		var me struct {
			AccountID string `json:"account_id"`
			Status    string `json:"status"`
		}
		_ = json.Unmarshal(body, &me)
		msg := "Подключение успешно"
		if me.AccountID != "" {
			msg = fmt.Sprintf("Подключение успешно (account_id: %s)", me.AccountID)
		}
		return &model.PaymentTestResult{OK: true, Message: msg}, nil
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &model.PaymentTestResult{
			OK:      false,
			Message: "Неверный Shop ID или секретный ключ",
		}, nil
	}

	return &model.PaymentTestResult{
		OK:      false,
		Message: fmt.Sprintf("ЮKassa вернула статус %d", resp.StatusCode),
	}, nil
}

func (s *PaymentSettingsService) testRobokassa(ctx context.Context) (*model.PaymentTestResult, error) {
	rec, err := s.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	rk := rec.Config.Robokassa
	if strings.TrimSpace(rk.MerchantLogin) == "" {
		return &model.PaymentTestResult{OK: false, Message: "Укажите логин магазина"}, nil
	}
	if strings.TrimSpace(rk.Password1) == "" {
		return &model.PaymentTestResult{OK: false, Message: "Укажите пароль #1"}, nil
	}
	if strings.TrimSpace(rk.Password2) == "" {
		return &model.PaymentTestResult{OK: false, Message: "Укажите пароль #2"}, nil
	}
	return &model.PaymentTestResult{
		OK:      true,
		Message: "Учётные данные заполнены. В кабинете Robokassa укажите Result URL backend API. Server отвечает OK{InvId}.",
	}, nil
}

func (s *PaymentSettingsService) YookassaWebhookURL() string {
	return s.cfg.PublicBaseURL() + "/api/v1/billing/yookassa/webhook"
}

func (s *PaymentSettingsService) RobokassaResultURL() string {
	return s.cfg.PublicBaseURL() + "/api/v1/billing/robokassa/result"
}

func (s *PaymentSettingsService) RobokassaResult2URL() string {
	return s.cfg.PublicBaseURL() + "/api/v1/billing/robokassa/result2"
}

func (s *PaymentSettingsService) RobokassaLegacyResultURL() string {
	return s.cfg.PublicBaseURL() + "/api/v1/billing/robokassa/result"
}

func (s *PaymentSettingsService) defaultReturnURL() string {
	return s.cfg.PublicBaseURL() + "/dashboard/billing"
}

func (s *PaymentSettingsService) PaymentsEnabled(ctx context.Context) (bool, model.PaymentProvider, error) {
	cfg, err := s.GetEffective(ctx)
	if err != nil {
		return false, "", err
	}
	return IsPaymentEnabled(cfg), cfg.ActiveProvider, nil
}

func IsPaymentEnabled(cfg model.PaymentSettings) bool {
	switch cfg.ActiveProvider {
	case model.PaymentProviderYookassa:
		return cfg.Yookassa.Enabled && yookassaConfigured(cfg.Yookassa)
	case model.PaymentProviderRobokassa:
		// Active Robokassa + credentials is enough (no separate enable toggle was required in admin UI).
		return robokassaConfigured(cfg.Robokassa)
	default:
		return false
	}
}

func yookassaConfigured(yk model.YookassaSettings) bool {
	return strings.TrimSpace(yk.ShopID) != "" && strings.TrimSpace(yk.SecretKey) != ""
}

func robokassaConfigured(rk model.RobokassaSettings) bool {
	return strings.TrimSpace(rk.MerchantLogin) != "" &&
		strings.TrimSpace(rk.Password1) != "" &&
		strings.TrimSpace(rk.Password2) != ""
}

// RobokassaConfigured reports whether Robokassa credentials are stored.
func RobokassaConfigured(rk model.RobokassaSettings) bool {
	return robokassaConfigured(rk)
}

func (s *PaymentSettingsService) buildAdminView(rec *model.PaymentSettingsRecord) *model.PaymentAdminView {
	yk := rec.Config.Yookassa
	rk := rec.Config.Robokassa
	return &model.PaymentAdminView{
		ActiveProvider:         rec.Config.ActiveProvider,
		Yookassa:               model.YookassaAdminSettings{ShopID: yk.ShopID, ReturnURL: yk.ReturnURL, Enabled: yk.Enabled},
		Robokassa:              model.RobokassaAdminSettings{MerchantLogin: rk.MerchantLogin, TestMode: rk.TestMode, Enabled: rk.Enabled},
		YookassaSecretSet:      strings.TrimSpace(yk.SecretKey) != "",
		YookassaSecretHint:     maskSecret(yk.SecretKey),
		RobokassaPassword1Set:  strings.TrimSpace(rk.Password1) != "",
		RobokassaPassword1Hint: maskSecret(rk.Password1),
		RobokassaPassword2Set:  strings.TrimSpace(rk.Password2) != "",
		RobokassaPassword2Hint: maskSecret(rk.Password2),
		YookassaWebhookURL:     s.YookassaWebhookURL(),
		RobokassaResultURL:     s.RobokassaResultURL(),
		RobokassaResult2URL:    s.RobokassaResult2URL(),
		DefaultReturnURL:       s.defaultReturnURL(),
		UpdatedAt:              rec.UpdatedAt,
	}
}

func validatePaymentSettings(req model.PaymentAdminUpdateRequest) error {
	switch req.ActiveProvider {
	case model.PaymentProviderYookassa, model.PaymentProviderRobokassa:
	default:
		return fmt.Errorf("%w: invalid active provider", ErrInvalidPaymentSettings)
	}
	if req.Yookassa.Enabled {
		if strings.TrimSpace(req.Yookassa.ShopID) == "" {
			return fmt.Errorf("%w: yookassa shop_id required when enabled", ErrInvalidPaymentSettings)
		}
	}
	if req.Robokassa.Enabled {
		if strings.TrimSpace(req.Robokassa.MerchantLogin) == "" {
			return fmt.Errorf("%w: robokassa merchant_login required when enabled", ErrInvalidPaymentSettings)
		}
	}
	return nil
}
