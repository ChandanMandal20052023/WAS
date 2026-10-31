package contacts

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"was/backend/internal/models"
	"was/backend/internal/utils"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateContact(ctx context.Context, orgID uuid.UUID, req models.CreateContactRequest) (*models.ContactWithTags, error) {
	normPhone, err := utils.NormalizePhoneNumber(req.PhoneNumber)
	if err != nil {
		return nil, err
	}
	req.PhoneNumber = normPhone

	if strings.TrimSpace(req.FullName) == "" {
		req.FullName = normPhone
	}

	return s.repo.Create(ctx, orgID, req)
}

func (s *Service) GetContact(ctx context.Context, orgID, contactID uuid.UUID) (*models.ContactWithTags, error) {
	return s.repo.GetByID(ctx, orgID, contactID)
}

func (s *Service) ListContacts(ctx context.Context, orgID uuid.UUID, filter models.ContactFilter) ([]models.ContactWithTags, int, error) {
	return s.repo.List(ctx, orgID, filter)
}

func (s *Service) UpdateContact(ctx context.Context, orgID, contactID uuid.UUID, req models.UpdateContactRequest) (*models.ContactWithTags, error) {
	if strings.TrimSpace(req.FullName) == "" {
		return nil, errors.New("full name cannot be empty")
	}
	return s.repo.Update(ctx, orgID, contactID, req)
}

func (s *Service) DeleteContact(ctx context.Context, orgID, contactID uuid.UUID) error {
	return s.repo.Delete(ctx, orgID, contactID)
}

func (s *Service) ListTags(ctx context.Context, orgID uuid.UUID) ([]models.Tag, error) {
	return s.repo.ListTags(ctx, orgID)
}

func (s *Service) CreateTag(ctx context.Context, orgID uuid.UUID, req models.CreateTagRequest) (*models.Tag, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("tag name cannot be empty")
	}
	return s.repo.CreateTag(ctx, orgID, name, req.ColorHex)
}
