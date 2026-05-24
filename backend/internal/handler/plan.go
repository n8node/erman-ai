package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type PlanHandler struct {
	plans *service.PlanService
}

func NewPlanHandler(plans *service.PlanService) *PlanHandler {
	return &PlanHandler{plans: plans}
}

type planRequest struct {
	Slug            string         `json:"slug"`
	Name            string         `json:"name"`
	PriceMonthlyRUB int            `json:"price_monthly_rub"`
	PriceYearlyRUB  int            `json:"price_yearly_rub"`
	ToolLimits      map[string]int `json:"tool_limits"`
	Features        map[string]any `json:"features"`
	SupportLevel    string         `json:"support_level"`
	IsPublic        bool           `json:"is_public"`
	IsArchived      bool           `json:"is_archived"`
}

func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.plans.ListAdmin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load plans")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *PlanHandler) Meta(w http.ResponseWriter, r *http.Request) {
	tools, err := h.plans.ListTools(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load tools")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tools":             tools,
		"default_features":  service.DefaultPlanFeatures(),
		"support_levels":    []string{"community", "email", "priority"},
	})
}

func (h *PlanHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.plans.GetAdmin(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "plan not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load plan")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req planRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	plan, err := h.plans.Create(r.Context(), planRequestToInput(req))
	if err != nil {
		h.writePlanError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, plan)
}

func (h *PlanHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req planRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	plan, err := h.plans.Update(r.Context(), id, planRequestToInput(req))
	if err != nil {
		h.writePlanError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, plan)
}

func planRequestToInput(req planRequest) service.PlanInput {
	return service.PlanInput{
		Slug:            req.Slug,
		Name:            req.Name,
		PriceMonthlyRUB: req.PriceMonthlyRUB,
		PriceYearlyRUB:  req.PriceYearlyRUB,
		ToolLimits:      req.ToolLimits,
		Features:        req.Features,
		SupportLevel:    req.SupportLevel,
		IsPublic:        req.IsPublic,
		IsArchived:      req.IsArchived,
	}
}

func (h *PlanHandler) writePlanError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid plan data")
	case errors.Is(err, service.ErrPlanSlugTaken):
		writeError(w, http.StatusConflict, "plan slug already exists")
	case errors.Is(err, service.ErrPlanInvalidSlug):
		writeError(w, http.StatusBadRequest, "invalid plan slug")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "plan not found")
	default:
		writeError(w, http.StatusInternalServerError, "plan operation failed")
	}
}
