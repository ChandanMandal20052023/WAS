package channels

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"was/backend/internal/middleware"
	"was/backend/internal/models"
	"was/backend/internal/utils"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())

	channels, err := h.service.ListChannels(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, channels)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid channel ID")
		return
	}

	channel, err := h.service.GetChannel(r.Context(), orgID, channelID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, channel)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())

	var req models.CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	channel, err := h.service.CreateChannel(r.Context(), orgID, &req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, channel)
}

func (h *Handler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid channel ID")
		return
	}

	qrResponse, err := h.service.GetQRCode(r.Context(), orgID, channelID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "QR_FETCH_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, qrResponse)
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid channel ID")
		return
	}

	if err := h.service.Connect(r.Context(), orgID, channelID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "CONNECT_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Connection initiated"})
}

func (h *Handler) SyncStatus(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid channel ID")
		return
	}

	statusResponse, err := h.service.SyncStatus(r.Context(), orgID, channelID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "STATUS_CHECK_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, statusResponse)
}

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid channel ID")
		return
	}

	if err := h.service.Disconnect(r.Context(), orgID, channelID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "DISCONNECT_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Channel disconnected"})
}

func (h *Handler) Reconnect(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid channel ID")
		return
	}

	if err := h.service.Reconnect(r.Context(), orgID, channelID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "RECONNECT_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Reconnection triggered"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid channel ID")
		return
	}

	if err := h.service.DeleteChannel(r.Context(), orgID, channelID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Channel deleted successfully"})
}

func (h *Handler) PairPhone(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	channelID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid channel ID")
		return
	}

	var req models.PairPhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	code, err := h.service.PairPhone(r.Context(), orgID, channelID, req.Phone)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "PAIR_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"pairing_code": code})
}
