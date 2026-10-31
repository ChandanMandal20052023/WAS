package auth

import (
	"encoding/json"
	"net/http"

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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.service.Register(r.Context(), &req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.service.Login(r.Context(), &req)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "AUTH_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, res)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.service.RefreshToken(r.Context(), &req)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "REFRESH_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, res)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Stateless JWT logout (client deletes token; optional Redis token blacklist)
	utils.JSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	orgID := middleware.GetOrganizationID(r.Context())

	res, err := h.service.GetMe(r.Context(), userID, orgID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, res)
}

func (h *Handler) Invite(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())

	var req models.InviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	invitation, err := h.service.InviteMember(r.Context(), orgID, &req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVITE_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusCreated, invitation)
}

func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req models.AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.service.AcceptInvitation(r.Context(), &req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "ACCEPT_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, res)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	token, err := h.service.ForgotPassword(r.Context(), &req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "FORGOT_PASSWORD_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{
		"message":     "If your email is registered, you will receive password reset instructions.",
		"reset_token": token,
	})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.ResetPassword(r.Context(), &req); err != nil {
		utils.Error(w, http.StatusBadRequest, "RESET_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Password updated successfully"})
}

