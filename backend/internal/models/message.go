package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SenderType string

const (
	SenderTypeCustomer SenderType = "customer"
	SenderTypeAgent    SenderType = "agent"
	SenderTypeSystem   SenderType = "system"
)

type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeAudio    MessageType = "audio"
	MessageTypeVideo    MessageType = "video"
	MessageTypeDocument MessageType = "document"
)

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusRead      DeliveryStatus = "read"
	DeliveryStatusFailed    DeliveryStatus = "failed"
)

type Message struct {
	ID                uuid.UUID       `json:"id"`
	OrganizationID    uuid.UUID       `json:"organization_id"`
	ConversationID    uuid.UUID       `json:"conversation_id"`
	WhatsAppMessageID *string         `json:"whatsapp_message_id,omitempty"`
	SenderType        SenderType      `json:"sender_type"`
	SenderUserID      *uuid.UUID      `json:"sender_user_id,omitempty"`
	SenderUserName    *string         `json:"sender_user_name,omitempty"`
	MessageType       MessageType     `json:"message_type"`
	Body              *string         `json:"body,omitempty"`
	MediaURL          *string         `json:"media_url,omitempty"`
	MediaMimeType     *string         `json:"media_mime_type,omitempty"`
	DeliveryStatus    DeliveryStatus  `json:"delivery_status"`
	RawPayload        json.RawMessage `json:"raw_payload,omitempty"`
	SentAt            time.Time       `json:"sent_at"`
	CreatedAt         time.Time       `json:"created_at"`
}

type SendMessageRequest struct {
	Body string `json:"body"`
}

type InternalNote struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
	ContactID      *uuid.UUID `json:"contact_id,omitempty"`
	UserID         uuid.UUID  `json:"user_id"`
	UserName       string     `json:"user_name"`
	Content        string     `json:"content"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateInternalNoteRequest struct {
	Content string `json:"content"`
}

