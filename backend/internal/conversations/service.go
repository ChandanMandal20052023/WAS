package conversations

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"was/backend/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetConversation(ctx context.Context, orgID, convID uuid.UUID) (*models.ConversationDetail, error) {
	return s.repo.GetByID(ctx, orgID, convID)
}

func (s *Service) ListConversations(ctx context.Context, orgID uuid.UUID, filter models.ConversationFilter) ([]models.ConversationDetail, int, error) {
	return s.repo.List(ctx, orgID, filter)
}

func (s *Service) UpdateStatus(ctx context.Context, orgID, convID uuid.UUID, status models.ConversationStatus) error {
	switch status {
	case models.ConversationStatusOpen, models.ConversationStatusPending, models.ConversationStatusResolved:
		return s.repo.UpdateStatus(ctx, orgID, convID, status)
	default:
		return errors.New("invalid status: must be open, pending, or resolved")
	}
}

func (s *Service) AssignAgent(ctx context.Context, orgID, convID uuid.UUID, agentID *uuid.UUID, assignedBy uuid.UUID, reason string) error {
	return s.repo.AssignAgent(ctx, orgID, convID, agentID, assignedBy, reason)
}

func (s *Service) MarkAsRead(ctx context.Context, orgID, convID uuid.UUID) error {
	return s.repo.ResetUnread(ctx, orgID, convID)
}

