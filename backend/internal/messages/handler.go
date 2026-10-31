package messages

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"was/backend/internal/middleware"
	"was/backend/internal/models"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/{id}/messages", h.ListMessages)
	r.Post("/{id}/messages", h.SendMessage)
	r.Get("/{id}/notes", h.ListNotes)
	r.Post("/{id}/notes", h.CreateNote)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid conversation id"}`, http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	messages, err := h.service.ListMessages(r.Context(), claims.OrganizationID, convID, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid conversation id"}`, http.StatusBadRequest)
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	msg, err := h.service.SendTextMessage(r.Context(), claims.OrganizationID, convID, claims.UserID, req.Body)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid conversation id"}`, http.StatusBadRequest)
		return
	}

	notes, err := h.service.ListNotes(r.Context(), claims.OrganizationID, convID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid conversation id"}`, http.StatusBadRequest)
		return
	}

	var req models.CreateInternalNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	note, err := h.service.CreateNote(r.Context(), claims.OrganizationID, convID, claims.UserID, req.Content)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

