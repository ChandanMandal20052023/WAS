package messages

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"was/backend/internal/channels"
	"was/backend/internal/contacts"
	"was/backend/internal/conversations"
	"was/backend/internal/evolution"
	"was/backend/internal/models"
	"was/backend/internal/utils"
)

type Service struct {
	repo         *Repository
	convRepo     *conversations.Repository
	contactRepo  *contacts.Repository
	channelRepo  channels.Repository
	gateway      evolution.EvolutionGateway
}

func NewService(
	repo *Repository,
	convRepo *conversations.Repository,
	contactRepo *contacts.Repository,
	channelRepo channels.Repository,
	gateway evolution.EvolutionGateway,
) *Service {
	return &Service{
		repo:        repo,
		convRepo:    convRepo,
		contactRepo: contactRepo,
		channelRepo: channelRepo,
		gateway:     gateway,
	}
}

func (s *Service) ListMessages(ctx context.Context, orgID, convID uuid.UUID, limit, offset int) ([]models.Message, error) {
	// Verify conversation belongs to organization
	_, err := s.convRepo.GetByID(ctx, orgID, convID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByConversation(ctx, orgID, convID, limit, offset)
}

func (s *Service) SendTextMessage(ctx context.Context, orgID, convID, userID uuid.UUID, text string) (*models.Message, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("message text cannot be empty")
	}

	conv, err := s.convRepo.GetByID(ctx, orgID, convID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	channel, err := s.channelRepo.GetByID(ctx, orgID, conv.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	// Dispatch outbound message via Evolution Go
	waMsgID, err := s.gateway.SendTextMessage(ctx, channel.InstanceToken, conv.ContactPhone, text)
	if err != nil {
		return nil, fmt.Errorf("failed to transmit WhatsApp message: %w", err)
	}

	msg := &models.Message{
		OrganizationID:    orgID,
		ConversationID:    convID,
		WhatsAppMessageID: &waMsgID,
		SenderType:        models.SenderTypeAgent,
		SenderUserID:      &userID,
		MessageType:       models.MessageTypeText,
		Body:              &text,
		DeliveryStatus:    models.DeliveryStatusSent,
		SentAt:            time.Now(),
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save sent message: %w", err)
	}

	// Update conversation last message preview
	_ = s.convRepo.UpdateLastMessage(ctx, convID, text, time.Now(), false)

	return msg, nil
}

func (s *Service) ListNotes(ctx context.Context, orgID, convID uuid.UUID) ([]models.InternalNote, error) {
	return s.repo.ListNotesByConversation(ctx, orgID, convID)
}

func (s *Service) CreateNote(ctx context.Context, orgID, convID, userID uuid.UUID, content string) (*models.InternalNote, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("note content cannot be empty")
	}

	conv, err := s.convRepo.GetByID(ctx, orgID, convID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	note := &models.InternalNote{
		OrganizationID: orgID,
		ConversationID: &convID,
		ContactID:      &conv.ContactID,
		UserID:         userID,
		Content:        content,
	}

	if err := s.repo.CreateNote(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to create internal note: %w", err)
	}

	return note, nil
}

// ProcessInboundMessage implements evolution.InboundProcessor
func (s *Service) ProcessInboundMessage(ctx context.Context, token, senderPhone, senderName, msgID, text string, timestamp time.Time) error {
	channel, err := s.channelRepo.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("channel not found for token: %w", err)
	}

	normPhone, err := utils.NormalizePhoneNumber(senderPhone)
	if err != nil {
		normPhone = senderPhone
	}

	if senderName == "" {
		senderName = normPhone
	}

	// 1. Auto-upsert contact
	contact, err := s.contactRepo.FindOrCreateByPhone(ctx, channel.OrganizationID, normPhone, senderName)
	if err != nil {
		return fmt.Errorf("failed to upsert contact on inbound: %w", err)
	}

	// 2. Auto-upsert conversation thread
	conv, err := s.convRepo.GetOrCreate(ctx, channel.OrganizationID, channel.ID, contact.ID)
	if err != nil {
		return fmt.Errorf("failed to get or create conversation thread: %w", err)
	}

	// 3. Insert incoming message
	var pMsgID *string
	if msgID != "" {
		pMsgID = &msgID
	}

	msg := &models.Message{
		OrganizationID:    channel.OrganizationID,
		ConversationID:    conv.ID,
		WhatsAppMessageID: pMsgID,
		SenderType:        models.SenderTypeCustomer,
		MessageType:       models.MessageTypeText,
		Body:              &text,
		DeliveryStatus:    models.DeliveryStatusDelivered,
		SentAt:            timestamp,
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		return fmt.Errorf("failed to persist inbound message: %w", err)
	}

	// 4. Update last message preview and increment unread count
	_ = s.convRepo.UpdateLastMessage(ctx, conv.ID, text, timestamp, true)

	return nil
}
