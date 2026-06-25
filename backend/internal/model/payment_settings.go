package model

import "time"

type PaymentProvider string

const (
	PaymentProviderYookassa  PaymentProvider = "yookassa"
	PaymentProviderRobokassa PaymentProvider = "robokassa"
)

type YookassaSettings struct {
	ShopID    string `json:"shop_id"`
	SecretKey string `json:"secret_key"`
	ReturnURL string `json:"return_url"`
	Enabled   bool   `json:"enabled"`
}

type RobokassaSettings struct {
	MerchantLogin string `json:"merchant_login"`
	Password1     string `json:"password1"`
	Password2     string `json:"password2"`
	TestMode      bool   `json:"test_mode"`
	Enabled       bool   `json:"enabled"`
}

type PaymentSettings struct {
	ActiveProvider PaymentProvider   `json:"active_provider"`
	Yookassa       YookassaSettings  `json:"yookassa"`
	Robokassa      RobokassaSettings `json:"robokassa"`
}

type PaymentSettingsRecord struct {
	Config    PaymentSettings `json:"config"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type YookassaAdminSettings struct {
	ShopID    string `json:"shop_id"`
	ReturnURL string `json:"return_url"`
	Enabled   bool   `json:"enabled"`
}

type RobokassaAdminSettings struct {
	MerchantLogin string `json:"merchant_login"`
	TestMode      bool   `json:"test_mode"`
	Enabled       bool   `json:"enabled"`
}

type PaymentAdminView struct {
	ActiveProvider       PaymentProvider        `json:"active_provider"`
	Yookassa             YookassaAdminSettings  `json:"yookassa"`
	Robokassa            RobokassaAdminSettings `json:"robokassa"`
	YookassaSecretSet    bool                   `json:"yookassa_secret_set"`
	YookassaSecretHint   string                 `json:"yookassa_secret_hint,omitempty"`
	RobokassaPassword1Set bool                  `json:"robokassa_password1_set"`
	RobokassaPassword1Hint string                `json:"robokassa_password1_hint,omitempty"`
	RobokassaPassword2Set bool                  `json:"robokassa_password2_set"`
	RobokassaPassword2Hint string                `json:"robokassa_password2_hint,omitempty"`
	YookassaWebhookURL     string                 `json:"yookassa_webhook_url"`
	RobokassaResultURL     string                 `json:"robokassa_result_url"`
	RobokassaResult2URL    string                 `json:"robokassa_result2_url"`
	DefaultReturnURL       string                 `json:"default_return_url"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

type PaymentAdminUpdateRequest struct {
	ActiveProvider       PaymentProvider        `json:"active_provider"`
	Yookassa             YookassaAdminSettings  `json:"yookassa"`
	Robokassa            RobokassaAdminSettings `json:"robokassa"`
	YookassaSecretKey    string                 `json:"yookassa_secret_key,omitempty"`
	RobokassaPassword1   string                 `json:"robokassa_password1,omitempty"`
	RobokassaPassword2   string                 `json:"robokassa_password2,omitempty"`
}

type PaymentTestResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type PaymentTestRequest struct {
	Provider PaymentProvider `json:"provider"`
}

func DefaultPaymentSettings(defaultReturnURL string) PaymentSettings {
	return PaymentSettings{
		ActiveProvider: PaymentProviderYookassa,
		Yookassa: YookassaSettings{
			ReturnURL: defaultReturnURL,
		},
	}
}
