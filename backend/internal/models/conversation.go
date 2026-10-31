package models

import (
	"time"

	"github.com/google/uuid"
)

type ConversationStatus string

const (
	ConversationStatusOpen     ConversationStatus = "open"
	ConversationStatusPending  ConversationStatus = "pending"
	ConversationStatusResolved ConversationStatus = "resolved"
)

type Conversation struct {
	ID                 uuid.UUID          `json:"id"`
	OrganizationID     uuid.UUID          `json:"organization_id"`
	ChannelID          uuid.UUID          `json:"channel_id"`
	ContactID          uuid.UUID          `json:"contact_id"`
	AssignedAgentID    *uuid.UUID         `json:"assigned_agent_id,omitempty"`
	Status             ConversationStatus `json:"status"`
	LastMessagePreview *string            `json:"last_message_preview,omitempty"`
	LastMessageTime    time.Time          `json:"last_message_time"`
	UnreadCount        int                `json:"unread_count"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type ConversationDetail struct {
	Conversation
	ContactName      string    `json:"contact_name"`
	ContactPhone     string    `json:"contact_phone"`
	ContactAvatar    string    `json:"contact_avatar,omitempty"`
	ChannelName      string    `json:"channel_name"`
	AssignedAgentName *string  `json:"assigned_agent_name,omitempty"`
	AssignedAgentEmail *string `json:"assigned_agent_email,omitempty"`
	Tags             []Tag     `json:"tags"`
}

type UpdateConversationStatusRequest struct {
	Status ConversationStatus `json:"status"`
}

type AssignConversationRequest struct {
	AgentID *uuid.UUID `json:"agent_id"`
	Reason  string     `json:"reason,omitempty"`
}

type ConversationFilter struct {
	Status    *ConversationStatus
	ChannelID *uuid.UUID
	AgentID   *uuid.UUID
	Search    string
	Limit     int
	Offset    int
}

