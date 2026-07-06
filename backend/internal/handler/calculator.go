package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type CalculatorHandler struct {
	calc    *service.CalculatorService
	billing *service.BillingService
	auth    *service.AuthService
	cfg     *config.Config
}

func NewCalculatorHandler(calc *service.CalculatorService, billing *service.BillingService, auth *service.AuthService, cfg *config.Config) *CalculatorHandler {
	return &CalculatorHandler{calc: calc, billing: billing, auth: auth, cfg: cfg}
}

func (h *CalculatorHandler) Run(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input model.CalculatorInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	run, output, err := h.calc.Run(r.Context(), userID, input, user.Locale)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}
		writeError(w, http.StatusInternalServerError, "calculation failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"run_id": run.ID,
		"input":  input,
		"output": output,
	})
}

type exportRequest struct {
	RunID string `json:"run_id"`
}

func (h *CalculatorHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.billing.CanExportPDF(r.Context(), userID); err != nil {
		if errors.Is(err, service.ErrFeatureNotAvailable) {
			writeError(w, http.StatusPaymentRequired, "pdf export requires pro plan")
			return
		}
		writeError(w, http.StatusInternalServerError, "billing check failed")
		return
	}

	var req exportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RunID == "" {
		writeError(w, http.StatusBadRequest, "run_id required")
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	run, err := h.calc.GetRunForUser(r.Context(), req.RunID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}

	in, out, err := service.ParseCalculatorRun(run)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid run data")
		return
	}

	pdfBytes, err := service.GenerateCalculatorPDF(in, out, user.Locale)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "pdf generation failed")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="calculator-report.pdf"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

type ShareHandler struct {
	share   *service.ShareService
	auth    *service.AuthService
	billing *service.BillingService
	cfg     *config.Config
	runs    *repository.ToolRunRepository
}

func NewShareHandler(share *service.ShareService, auth *service.AuthService, billing *service.BillingService, cfg *config.Config, runs *repository.ToolRunRepository) *ShareHandler {
	return &ShareHandler{share: share, auth: auth, billing: billing, cfg: cfg, runs: runs}
}

func (h *ShareHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	runID := chi.URLParam(r, "id")
	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := h.share.CreateShare(r.Context(), userID, runID, h.cfg.PublicBaseURL(), user)
	if err != nil {
		h.writeShareError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h *ShareHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	payload, err := h.share.GetPublicReport(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusNotFound, "report not found or expired")
		return
	}

	in, out, err := service.ParseCalculatorRun(payload.Run)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid report data")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"process_name":      in.ProcessName,
		"input":             in,
		"output":            out,
		"created_at":        payload.Run.CreatedAt,
		"show_platform_cta": payload.ShowPlatformCta,
	})
}

func (h *ShareHandler) writeShareError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrFeatureNotAvailable):
		writeError(w, http.StatusPaymentRequired, "share report requires pro plan")
	case errors.Is(err, service.ErrShareLimitExceeded):
		writeError(w, http.StatusPaymentRequired, "monthly share limit exceeded")
	case errors.Is(err, service.ErrPartnerOnly):
		writeError(w, http.StatusForbidden, "share available for partner accounts only")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "run not found")
	default:
		writeError(w, http.StatusInternalServerError, "share failed")
	}
}

type LeadHandler struct {
	leads *service.LeadService
	auth  *service.AuthService
}

func NewLeadHandler(leads *service.LeadService, auth *service.AuthService) *LeadHandler {
	return &LeadHandler{leads: leads, auth: auth}
}

type leadRequest struct {
	RunID   string  `json:"run_id"`
	Name    string  `json:"name"`
	Company *string `json:"company"`
	Phone   *string `json:"phone"`
	Email   string  `json:"email"`
	Message *string `json:"message"`
}

func (h *LeadHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.auth.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req leadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	lead, err := h.leads.Create(r.Context(), user, req.RunID, req.Name, req.Email, req.Company, req.Phone, req.Message)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid input")
			return
		}
		writeError(w, http.StatusForbidden, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":      lead.ID,
		"status":  "ok",
		"message": "request received",
	})
}

type BillingHandler struct {
	billing       *service.BillingService
	checkout      *service.CheckoutService
	tokenPackages *service.TokenPackageService
	plans         *repository.PlanRepository
	runs          *repository.ToolRunRepository
}

func NewBillingHandler(
	billing *service.BillingService,
	checkout *service.CheckoutService,
	tokenPackages *service.TokenPackageService,
	plans *repository.PlanRepository,
	runs *repository.ToolRunRepository,
) *BillingHandler {
	return &BillingHandler{
		billing:       billing,
		checkout:      checkout,
		tokenPackages: tokenPackages,
		plans:         plans,
		runs:          runs,
	}
}

func (h *BillingHandler) Plan(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	up, err := h.billing.GetUserPlan(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load plan")
		return
	}

	shareUsed, _ := h.runs.GetUsageCount(r.Context(), userID, "share_report")
	shareLimit := h.billing.ShareReportLimit(&up.Plan)

	toolUsage := make(map[string]map[string]int)
	for _, slug := range []string{"calculator", "strategy", "proposal", "legal-scan"} {
		limit := -1
		if up.Plan.ToolLimits != nil {
			if v, ok := up.Plan.ToolLimits[slug]; ok {
				limit = v
			}
		}
		used, _ := h.runs.GetUsageCount(r.Context(), userID, slug)
		toolUsage[slug] = map[string]int{"used": used, "limit": limit}
	}

	paymentsEnabled := false
	var paymentProvider string
	if h.checkout != nil {
		var provider model.PaymentProvider
		paymentsEnabled, provider = h.checkout.PaymentsEnabled(r.Context())
		paymentProvider = string(provider)
	}

	tokenBalance := int64(0)
	if h.tokenPackages != nil {
		tokenBalance, _ = h.tokenPackages.GetUserBalance(r.Context(), userID)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"plan_id":          up.Plan.ID,
		"plan_slug":        up.PlanSlug,
		"plan_name":        up.Plan.Name,
		"features":         up.Plan.Features,
		"tool_limits":      up.Plan.ToolLimits,
		"payments_enabled": paymentsEnabled,
		"payment_provider": paymentProvider,
		"token_balance":    tokenBalance,
		"usage": map[string]any{
			"share_report_used":  shareUsed,
			"share_report_limit": shareLimit,
			"tools":              toolUsage,
		},
	})
}

func (h *BillingHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.billing.ListPublicPlans(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load plans")
		return
	}

	paymentsEnabled := false
	var paymentProvider string
	if h.checkout != nil {
		var provider model.PaymentProvider
		paymentsEnabled, provider = h.checkout.PaymentsEnabled(r.Context())
		paymentProvider = string(provider)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":            plans,
		"payments_enabled": paymentsEnabled,
		"payment_provider": paymentProvider,
	})
}

type switchPlanRequest struct {
	PlanID string `json:"plan_id"`
}

type checkoutRequest struct {
	PlanID string `json:"plan_id"`
}

func (h *BillingHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PlanID == "" {
		writeError(w, http.StatusBadRequest, "plan_id required")
		return
	}

	result, err := h.checkout.Create(r.Context(), userID, req.PlanID)
	if err != nil {
		h.writeBillingError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *BillingHandler) SwitchPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req switchPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.PlanID == "" {
		writeError(w, http.StatusBadRequest, "plan_id required")
		return
	}

	role, _ := middleware.UserRoleFromContext(r.Context())
	plan, err := h.billing.SwitchPlan(r.Context(), userID, req.PlanID, role)
	if err != nil {
		h.writeBillingError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"plan_id":   plan.ID,
		"plan_slug": plan.Slug,
		"plan_name": plan.Name,
		"status":    "activated",
	})
}

func (h *BillingHandler) writeBillingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPaymentRequired):
		writeError(w, http.StatusPaymentRequired, "payment required")
	case errors.Is(err, service.ErrCheckoutUnavailable):
		writeError(w, http.StatusServiceUnavailable, "checkout unavailable")
	case errors.Is(err, service.ErrUserBlocked):
		writeError(w, http.StatusForbidden, "account blocked")
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid plan")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "plan not found")
	default:
		writeError(w, http.StatusInternalServerError, "billing operation failed")
	}
}

type ToolsHandler struct {
	plans   *repository.PlanRepository
	runs    *repository.ToolRunRepository
	billing *service.BillingService
}

func NewToolsHandler(plans *repository.PlanRepository, runs *repository.ToolRunRepository, billing *service.BillingService) *ToolsHandler {
	return &ToolsHandler{plans: plans, runs: runs, billing: billing}
}

func (h *ToolsHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tools, err := h.plans.ListTools(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tools")
		return
	}

	up, err := h.billing.GetUserPlan(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load plan")
		return
	}

	type toolItem struct {
		Slug        string `json:"slug"`
		Name        string `json:"name"`
		Description string `json:"description"`
		RunsUsed    int    `json:"runs_used"`
		RunsLimit   int    `json:"runs_limit"`
	}

	items := make([]toolItem, 0, len(tools))
	for _, t := range tools {
		limit := -1
		if up.Plan.ToolLimits != nil {
			if v, ok := up.Plan.ToolLimits[t.Slug]; ok {
				limit = v
			}
		}
		used, _ := h.runs.GetUsageCount(r.Context(), userID, t.Slug)
		items = append(items, toolItem{
			Slug: t.Slug, Name: t.Name, Description: t.Description,
			RunsUsed: used, RunsLimit: limit,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"tools": items})
}
