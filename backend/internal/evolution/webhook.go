package evolution

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"was/backend/internal/models"
	"was/backend/internal/utils"
)

type WebhookPayload struct {
	Event         string                 `json:"event"`
	InstanceToken string                 `json:"instanceToken"`
	InstanceID    string                 `json:"instanceId"`
	InstanceName  string                 `json:"instanceName"`
	Data          map[string]interface{} `json:"data"`
}

type ChannelUpdater interface {
	UpdateStatusByToken(ctx context.Context, token string, status models.ChannelStatus, phone *string, qrcode *string) error
}

type WebhookHandler struct {
	updater ChannelUpdater
}

func NewWebhookHandler(updater ChannelUpdater) *WebhookHandler {
	return &WebhookHandler{updater: updater}
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		utils.Error(w, http.StatusBadRequest, "INVALID_TOKEN", "Webhook token missing")
		return
	}

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to decode webhook payload")
		return
	}

	log.Printf("[evolution webhook] Received event '%s' for instance '%s'", payload.Event, payload.InstanceName)

	switch payload.Event {
	case "Connected", "PairSuccess":
		var phone *string
		if jid, ok := payload.Data["jid"].(string); ok && jid != "" {
			phone = &jid
		}
		_ = h.updater.UpdateStatusByToken(r.Context(), token, models.StatusConnected, phone, nil)

	case "Disconnected", "LoggedOut":
		_ = h.updater.UpdateStatusByToken(r.Context(), token, models.StatusDisconnected, nil, nil)

	case "QRCode":
		if qr, ok := payload.Data["qrcode"].(string); ok && qr != "" {
			_ = h.updater.UpdateStatusByToken(r.Context(), token, models.StatusWaitingForQRScan, nil, &qr)
		}

	case "Connecting":
		_ = h.updater.UpdateStatusByToken(r.Context(), token, models.StatusConnecting, nil, nil)
	}

	utils.JSON(w, http.StatusOK, map[string]string{"status": "processed"})
}
