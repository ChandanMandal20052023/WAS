package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	PhoneNumber    string          `json:"phone_number"`
	FullName       string          `json:"full_name"`
	Email          string          `json:"email,omitempty"`
	AvatarURL      string          `json:"avatar_url,omitempty"`
	Metadata       json.RawMessage `json:"metadata"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type Tag struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	ColorHex       string    `json:"color_hex"`
	CreatedAt      time.Time `json:"created_at"`
}

type ContactWithTags struct {
	Contact
	Tags []Tag `json:"tags"`
}

type CreateContactRequest struct {
	PhoneNumber string    `json:"phone_number"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	TagIDs      []uuid.UUID `json:"tag_ids,omitempty"`
}

type UpdateContactRequest struct {
	FullName  string      `json:"full_name"`
	Email     string      `json:"email,omitempty"`
	AvatarURL string      `json:"avatar_url,omitempty"`
	TagIDs    []uuid.UUID `json:"tag_ids,omitempty"`
}

type CreateTagRequest struct {
	Name     string `json:"name"`
	ColorHex string `json:"color_hex"`
}

type ContactFilter struct {
	Search string
	TagID  *uuid.UUID
	Limit  int
	Offset int
}
