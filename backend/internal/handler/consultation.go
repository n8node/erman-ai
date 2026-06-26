package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/erman-ai/erman-ai/internal/middleware"
	"github.com/erman-ai/erman-ai/internal/repository"
	"github.com/erman-ai/erman-ai/internal/service"
	"github.com/go-chi/chi/v5"
)

type ConsultationHandler struct {
	consultations *service.ConsultationService
}

func NewConsultationHandler(consultations *service.ConsultationService) *ConsultationHandler {
	return &ConsultationHandler{consultations: consultations}
}

type consultationBookingBody struct {
	ServiceID        string `json:"service_id"`
	StartsAt         string `json:"starts_at"`
	CustomerName     string `json:"customer_name"`
	CustomerEmail    string `json:"customer_email"`
	CustomerPhone    string `json:"customer_phone"`
	CustomerTelegram string `json:"customer_telegram"`
	CustomerNote     string `json:"customer_note"`
	Timezone         string `json:"timezone"`
	Website          string `json:"website"`
}

func (h *ConsultationHandler) ListPublicServices(w http.ResponseWriter, r *http.Request) {
	items, err := h.consultations.ListPublicServices(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load services")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ConsultationHandler) ListSlots(w http.ResponseWriter, r *http.Request) {
	serviceID := r.URL.Query().Get("service_id")
	from, _ := time.Parse(time.RFC3339, r.URL.Query().Get("from"))
	to, _ := time.Parse(time.RFC3339, r.URL.Query().Get("to"))
	if serviceID == "" {
		writeError(w, http.StatusBadRequest, "service_id required")
		return
	}
	if from.IsZero() {
		from = time.Now()
	}
	items, err := h.consultations.ListSlots(r.Context(), serviceID, from, to)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ConsultationHandler) CreatePublicBooking(w http.ResponseWriter, r *http.Request) {
	h.createBooking(w, r, nil)
}

func (h *ConsultationHandler) CreateAuthenticatedBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	h.createBooking(w, r, &userID)
}

func (h *ConsultationHandler) createBooking(w http.ResponseWriter, r *http.Request, userID *string) {
	var req consultationBookingBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Website != "" {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
		return
	}
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid starts_at")
		return
	}
	booking, err := h.consultations.CreateBooking(r.Context(), userID, service.ConsultationBookingInput{
		ServiceID:        req.ServiceID,
		StartsAt:         startsAt,
		CustomerName:     req.CustomerName,
		CustomerEmail:    req.CustomerEmail,
		CustomerPhone:    req.CustomerPhone,
		CustomerTelegram: req.CustomerTelegram,
		CustomerNote:     req.CustomerNote,
		Timezone:         req.Timezone,
	})
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, booking)
}

type consultationCheckoutBody struct {
	BookingID string `json:"booking_id"`
}

func (h *ConsultationHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	var req consultationCheckoutBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BookingID == "" {
		writeError(w, http.StatusBadRequest, "booking_id required")
		return
	}
	result, err := h.consultations.CreateCheckout(r.Context(), req.BookingID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *ConsultationHandler) ListAdminServices(w http.ResponseWriter, r *http.Request) {
	items, err := h.consultations.ListAdminServices(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load services")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ConsultationHandler) CreateAdminService(w http.ResponseWriter, r *http.Request) {
	var req service.ConsultationAdminServiceInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.consultations.CreateAdminService(r.Context(), req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ConsultationHandler) UpdateAdminService(w http.ResponseWriter, r *http.Request) {
	var req service.ConsultationAdminServiceInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.consultations.UpdateAdminService(r.Context(), chi.URLParam(r, "id"), req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ConsultationHandler) GetAdminAvailability(w http.ResponseWriter, r *http.Request) {
	items, err := h.consultations.ListAvailability(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type consultationAvailabilityBody struct {
	Items []service.ConsultationAdminAvailabilityInput `json:"items"`
}

func (h *ConsultationHandler) PutAdminAvailability(w http.ResponseWriter, r *http.Request) {
	var req consultationAvailabilityBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	items, err := h.consultations.ReplaceAvailability(r.Context(), chi.URLParam(r, "id"), req.Items)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *ConsultationHandler) ListAdminBookings(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.consultations.ListBookingsAdmin(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load bookings")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "limit": limit, "offset": offset})
}

type consultationStatusBody struct {
	Status string `json:"status"`
}

func (h *ConsultationHandler) UpdateAdminBookingStatus(w http.ResponseWriter, r *http.Request) {
	var req consultationStatusBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		writeError(w, http.StatusBadRequest, "status required")
		return
	}
	item, err := h.consultations.UpdateBookingStatus(r.Context(), chi.URLParam(r, "id"), req.Status)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ConsultationHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, service.ErrConsultationSlotUnavailable):
		writeError(w, http.StatusConflict, "slot unavailable")
	case errors.Is(err, service.ErrConsultationPaymentsOff):
		writeError(w, http.StatusServiceUnavailable, "payments disabled")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	default:
		writeError(w, http.StatusInternalServerError, "request failed")
	}
}
