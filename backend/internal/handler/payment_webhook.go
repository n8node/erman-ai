package handler

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/service"
)

type PaymentWebhookHandler struct {
	svc    *service.PaymentSettingsService
	logger *slog.Logger
}

func NewPaymentWebhookHandler(svc *service.PaymentSettingsService, logger *slog.Logger) *PaymentWebhookHandler {
	return &PaymentWebhookHandler{svc: svc, logger: logger}
}

func (h *PaymentWebhookHandler) YookassaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	var notification struct {
		Type   string `json:"type"`
		Event  string `json:"event"`
		Object struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"object"`
	}
	if err := json.Unmarshal(body, &notification); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	h.logger.Info("yookassa webhook received",
		"type", notification.Type,
		"event", notification.Event,
		"payment_id", notification.Object.ID,
		"status", notification.Object.Status,
	)

	w.WriteHeader(http.StatusOK)
}

func (h *PaymentWebhookHandler) RobokassaResult(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.svc.GetEffective(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "settings unavailable")
		return
	}

	rk := cfg.Robokassa
	if !rk.Enabled || strings.TrimSpace(rk.Password2) == "" {
		writeError(w, http.StatusServiceUnavailable, "robokassa not configured")
		return
	}

	q := r.URL.Query()
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err == nil && len(r.PostForm) > 0 {
			q = r.PostForm
		}
	}

	outSum := firstParam(q, "OutSum")
	invID := firstParam(q, "InvId")
	signature := strings.ToUpper(firstParam(q, "SignatureValue"))

	if outSum == "" || invID == "" || signature == "" {
		writeError(w, http.StatusBadRequest, "missing parameters")
		return
	}

	expected := strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(
		outSum+":"+invID+":"+strings.TrimSpace(rk.Password2),
	))))

	if signature != expected {
		h.logger.Warn("robokassa invalid signature", "inv_id", invID)
		writeError(w, http.StatusForbidden, "invalid signature")
		return
	}

	amount, _ := strconv.ParseFloat(outSum, 64)
	h.logger.Info("robokassa payment confirmed",
		"inv_id", invID,
		"amount", amount,
		"test_mode", rk.TestMode,
	)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("OK" + invID))
}

func firstParam(q interface {
	Get(string) string
}, key string) string {
	return strings.TrimSpace(q.Get(key))
}

// IsPaymentEnabled reports whether any configured provider accepts payments.
func IsPaymentEnabled(cfg model.PaymentSettings) bool {
	switch cfg.ActiveProvider {
	case model.PaymentProviderYookassa:
		return cfg.Yookassa.Enabled && strings.TrimSpace(cfg.Yookassa.ShopID) != "" && strings.TrimSpace(cfg.Yookassa.SecretKey) != ""
	case model.PaymentProviderRobokassa:
		return cfg.Robokassa.Enabled && strings.TrimSpace(cfg.Robokassa.MerchantLogin) != "" &&
			strings.TrimSpace(cfg.Robokassa.Password1) != "" && strings.TrimSpace(cfg.Robokassa.Password2) != ""
	default:
		return false
	}
}
