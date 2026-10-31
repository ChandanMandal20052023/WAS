package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("resource not found")

type ChannelStatus string

const (
	StatusDisconnected     ChannelStatus = "Disconnected"
	StatusCreating         ChannelStatus = "Creating"
	StatusWaitingForQRScan ChannelStatus = "Waiting for QR Scan"
	StatusConnecting       ChannelStatus = "Connecting"
	StatusConnected        ChannelStatus = "Connected"
	StatusReconnecting     ChannelStatus = "Reconnecting"
	StatusConnectionFailed ChannelStatus = "Connection Failed"
)

func (s ChannelStatus) IsValid() bool {
	switch s {
	case StatusDisconnected, StatusCreating, StatusWaitingForQRScan,
		StatusConnecting, StatusConnected, StatusReconnecting, StatusConnectionFailed:
		return true
	}
	return false
}

type WhatsAppChannel struct {
	ID              uuid.UUID              `json:"id"`
	OrganizationID  uuid.UUID              `json:"organization_id"`
	Name            string                 `json:"name"`
	PhoneNumber     *string                `json:"phone_number,omitempty"`
	InstanceName    string                 `json:"instance_name"`
	InstanceToken   string                 `json:"-"` // Never expose raw token to frontend
	Status          ChannelStatus          `json:"status"`
	QRCodeBase64    *string                `json:"qrcode_base64,omitempty"`
	Settings        map[string]interface{} `json:"settings"`
	LastConnectedAt *time.Time             `json:"last_connected_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// Channel DTOs
type CreateChannelRequest struct {
	Name string `json:"name"`
}

type PairPhoneRequest struct {
	Phone string `json:"phone"`
}

type QRCodeResponse struct {
	QRCode    string        `json:"qrcode"`
	Status    ChannelStatus `json:"status"`
	ExpiresIn int           `json:"expires_in"` // in seconds
}

type ChannelStatusResponse struct {
	ID          uuid.UUID     `json:"id"`
	Status      ChannelStatus `json:"status"`
	PhoneNumber *string       `json:"phone_number,omitempty"`
	Connected   bool          `json:"connected"`
}
