package handler

import (
	"encoding/json"
	"net/http"
	"ticket-reservation/service"
)

type Handler struct {
	svc *service.ReservationService
}

func New(svc *service.ReservationService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/slots", h.CreateSlot)
	mux.HandleFunc("GET /api/slots", h.ListSlots)
	mux.HandleFunc("GET /api/slots/{id}", h.GetSlot)
	mux.HandleFunc("POST /api/reservations", h.Reserve)
	mux.HandleFunc("DELETE /api/reservations/{id}", h.CancelReservation)
}

func (h *Handler) CreateSlot(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	slot, err := h.svc.CreateSlot(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, slot)
}

func (h *Handler) ListSlots(w http.ResponseWriter, r *http.Request) {
	slots := h.svc.ListSlots()
	writeJSON(w, http.StatusOK, slots)
}

func (h *Handler) GetSlot(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	slot, err := h.svc.GetSlot(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, slot)
}

func (h *Handler) Reserve(w http.ResponseWriter, r *http.Request) {
	var req service.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reservation, err := h.svc.Reserve(req)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, reservation)
}

func (h *Handler) CancelReservation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Cancel(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
