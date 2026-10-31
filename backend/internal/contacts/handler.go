package contacts

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
	r.Get("/", h.ListContacts)
	r.Post("/", h.CreateContact)
	r.Get("/{id}", h.GetContact)
	r.Put("/{id}", h.UpdateContact)
	r.Delete("/{id}", h.DeleteContact)
}

func (h *Handler) RegisterTagRoutes(r chi.Router) {
	r.Get("/", h.ListTags)
	r.Post("/", h.CreateTag)
}

func (h *Handler) ListContacts(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	search := r.URL.Query().Get("search")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	tagStr := r.URL.Query().Get("tag_id")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	filter := models.ContactFilter{
		Search: search,
		Limit:  limit,
		Offset: offset,
	}

	if tagStr != "" {
		if tID, err := uuid.Parse(tagStr); err == nil {
			filter.TagID = &tID
		}
	}

	contacts, total, err := h.service.ListContacts(r.Context(), claims.OrganizationID, filter)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"contacts": contacts,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req models.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	contact, err := h.service.CreateContact(r.Context(), claims.OrganizationID, req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contact)
}

func (h *Handler) GetContact(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	contactID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid contact id"}`, http.StatusBadRequest)
		return
	}

	contact, err := h.service.GetContact(r.Context(), claims.OrganizationID, contactID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contact)
}

func (h *Handler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	contactID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid contact id"}`, http.StatusBadRequest)
		return
	}

	var req models.UpdateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	contact, err := h.service.UpdateContact(r.Context(), claims.OrganizationID, contactID, req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contact)
}

func (h *Handler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	contactID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"invalid contact id"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteContact(r.Context(), claims.OrganizationID, contactID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	tags, err := h.service.ListTags(r.Context(), claims.OrganizationID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req models.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request payload"}`, http.StatusBadRequest)
		return
	}

	tag, err := h.service.CreateTag(r.Context(), claims.OrganizationID, req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

