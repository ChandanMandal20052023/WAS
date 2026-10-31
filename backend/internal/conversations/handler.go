package conversations

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
	r.Get("/", h.ListConversations)
	r.Get("/{id}", h.GetConversation)
	r.Patch("/{id}/status", h.UpdateStatus)
	r.Post("/{id}/assign", h.AssignAgent)
	r.Post("/{id}/read", h.MarkAsRead)
}

func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	search := r.URL.Query().Get("search")
	statusStr := r.URL.Query().Get("status")
	channelStr := r.URL.Query().Get("channel_id")
	agentStr := r.URL.Query().Get("agent_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	filter := models.ConversationFilter{
		Search: search,
		Limit:  limit,
		Offset: offset,
	}

	if statusStr != "" {
		st := models.ConversationStatus(statusStr)
		filter.Status = &st
	}

	if channelStr != "" {
		if chID, err := uuid.Parse(channelStr); err == nil {
			filter.ChannelID = &chID
		}
	}

	if agentStr != "" {
		if agID, err := uuid.Parse(agentStr); err == nil {
			filter.AgentID = &agID
		}
	}

	conversations, total, err := h.service.ListConversations(r.Context(), claims.OrganizationID, filter)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"conversations": conversations,
		"total":         total,
		"limit":         filter.Limit,
		"offset":        filter.Offset,
	})
}

func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
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

	conv, err := h.service.GetConversation(r.Context(), claims.OrganizationID, convID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conv)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
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

	var req models.UpdateConversationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(r.Context(), claims.OrganizationID, convID, req.Status); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "status updated successfully"})
}

func (h *Handler) AssignAgent(w http.ResponseWriter, r *http.Request) {
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

	var req models.AssignConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.AssignAgent(r.Context(), claims.OrganizationID, convID, req.AgentID, claims.UserID, req.Reason); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "agent assigned successfully"})
}

func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.MarkAsRead(r.Context(), claims.OrganizationID, convID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

