package evolution

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

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

type InboundProcessor interface {
	ProcessInboundMessage(ctx context.Context, token, senderPhone, senderName, msgID, text string, timestamp time.Time) error
}

type WebhookHandler struct {
	updater ChannelUpdater
	inbound InboundProcessor
}

func NewWebhookHandler(updater ChannelUpdater, inbound InboundProcessor) *WebhookHandler {
	return &WebhookHandler{updater: updater, inbound: inbound}
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

	case "messages.upsert", "Message", "messages.receive":
		if h.inbound != nil {
			var remoteJid, pushName, msgText, msgID string
			if sender, ok := payload.Data["sender"].(string); ok {
				remoteJid = sender
			} else if rJid, ok := payload.Data["remoteJid"].(string); ok {
				remoteJid = rJid
			}

			if name, ok := payload.Data["pushName"].(string); ok {
				pushName = name
			}

			if key, ok := payload.Data["key"].(map[string]interface{}); ok {
				if id, ok := key["id"].(string); ok {
					msgID = id
				}
				if fromMe, ok := key["fromMe"].(bool); ok && fromMe {
					// Ignore outbound echo messages
					break
				}
			}

			if msg, ok := payload.Data["message"].(map[string]interface{}); ok {
				if text, ok := msg["conversation"].(string); ok {
					msgText = text
				} else if ext, ok := msg["extendedTextMessage"].(map[string]interface{}); ok {
					if text, ok := ext["text"].(string); ok {
						msgText = text
					}
				}
			} else if text, ok := payload.Data["text"].(string); ok {
				msgText = text
			}

			// Clean remoteJid (e.g. "1234567890@s.whatsapp.net" -> "+1234567890")
			phone := strings.Split(remoteJid, "@")[0]
			if !strings.HasPrefix(phone, "+") && phone != "" {
				phone = "+" + phone
			}

			if phone != "" && msgText != "" {
				_ = h.inbound.ProcessInboundMessage(r.Context(), token, phone, pushName, msgID, msgText, time.Now())
			}
		}
	}

	utils.JSON(w, http.StatusOK, map[string]string{"status": "processed"})
}
