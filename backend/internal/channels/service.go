package channels

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"was/backend/internal/evolution"
	"was/backend/internal/models"
)

type Service interface {
	ListChannels(ctx context.Context, orgID uuid.UUID) ([]*models.WhatsAppChannel, error)
	GetChannel(ctx context.Context, orgID, id uuid.UUID) (*models.WhatsAppChannel, error)
	CreateChannel(ctx context.Context, orgID uuid.UUID, req *models.CreateChannelRequest) (*models.WhatsAppChannel, error)
	GetQRCode(ctx context.Context, orgID, id uuid.UUID) (*models.QRCodeResponse, error)
	Connect(ctx context.Context, orgID, id uuid.UUID) error
	SyncStatus(ctx context.Context, orgID, id uuid.UUID) (*models.ChannelStatusResponse, error)
	Disconnect(ctx context.Context, orgID, id uuid.UUID) error
	Reconnect(ctx context.Context, orgID, id uuid.UUID) error
	DeleteChannel(ctx context.Context, orgID, id uuid.UUID) error
	PairPhone(ctx context.Context, orgID, id uuid.UUID, phone string) (string, error)
}

type channelService struct {
	repo             Repository
	evolutionGateway evolution.EvolutionGateway
	webhookBaseURL   string
}

func NewService(repo Repository, gateway evolution.EvolutionGateway, webhookBaseURL string) Service {
	return &channelService{
		repo:             repo,
		evolutionGateway: gateway,
		webhookBaseURL:   webhookBaseURL,
	}
}

func (s *channelService) ListChannels(ctx context.Context, orgID uuid.UUID) ([]*models.WhatsAppChannel, error) {
	return s.repo.ListByOrg(ctx, orgID)
}

func (s *channelService) GetChannel(ctx context.Context, orgID, id uuid.UUID) (*models.WhatsAppChannel, error) {
	return s.repo.GetByID(ctx, orgID, id)
}

func (s *channelService) CreateChannel(ctx context.Context, orgID uuid.UUID, req *models.CreateChannelRequest) (*models.WhatsAppChannel, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("channel name is required")
	}

	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	instanceName := fmt.Sprintf("was_%s_%s", orgID.String()[:8], token[:8])

	channel := &models.WhatsAppChannel{
		OrganizationID: orgID,
		Name:           strings.TrimSpace(req.Name),
		InstanceName:   instanceName,
		InstanceToken:  token,
		Status:         models.StatusCreating,
	}

	// 1. Create instance in Evolution Go
	if err := s.evolutionGateway.CreateInstance(ctx, instanceName, token); err != nil {
		return nil, fmt.Errorf("evolution gateway instance creation failed: %w", err)
	}

	// 2. Persist in database
	if err := s.repo.Create(ctx, channel); err != nil {
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}

	// 3. Initiate connection and configure webhook
	webhookURL := fmt.Sprintf("%s/api/v1/webhooks/evolution/%s", s.webhookBaseURL, token)
	_ = s.evolutionGateway.ConnectInstance(ctx, token, webhookURL)
	_ = s.repo.UpdateStatus(ctx, orgID, channel.ID, models.StatusWaitingForQRScan, nil, nil)
	channel.Status = models.StatusWaitingForQRScan

	return channel, nil
}

func (s *channelService) GetQRCode(ctx context.Context, orgID, id uuid.UUID) (*models.QRCodeResponse, error) {
	ch, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	if ch.Status == models.StatusConnected {
		return &models.QRCodeResponse{
			QRCode:    "",
			Status:    models.StatusConnected,
			ExpiresIn: 0,
		}, nil
	}

	qr, err := s.evolutionGateway.GetQRCode(ctx, ch.InstanceToken)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve QR code from gateway: %w", err)
	}

	_ = s.repo.UpdateStatus(ctx, orgID, id, models.StatusWaitingForQRScan, nil, &qr)

	return &models.QRCodeResponse{
		QRCode:    qr,
		Status:    models.StatusWaitingForQRScan,
		ExpiresIn: 60,
	}, nil
}

func (s *channelService) Connect(ctx context.Context, orgID, id uuid.UUID) error {
	ch, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}

	webhookURL := fmt.Sprintf("%s/api/v1/webhooks/evolution/%s", s.webhookBaseURL, ch.InstanceToken)
	if err := s.evolutionGateway.ConnectInstance(ctx, ch.InstanceToken, webhookURL); err != nil {
		return err
	}

	return s.repo.UpdateStatus(ctx, orgID, id, models.StatusConnecting, nil, nil)
}

func (s *channelService) SyncStatus(ctx context.Context, orgID, id uuid.UUID) (*models.ChannelStatusResponse, error) {
	ch, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return nil, err
	}

	connected, jid, err := s.evolutionGateway.GetStatus(ctx, ch.InstanceToken)
	if err != nil {
		return nil, err
	}

	var status models.ChannelStatus
	var phone *string

	if connected {
		status = models.StatusConnected
		if jid != "" {
			phone = &jid
		}
	} else {
		if ch.Status == models.StatusConnected {
			status = models.StatusDisconnected
		} else {
			status = ch.Status
		}
	}

	_ = s.repo.UpdateStatus(ctx, orgID, id, status, phone, nil)

	return &models.ChannelStatusResponse{
		ID:          ch.ID,
		Status:      status,
		PhoneNumber: phone,
		Connected:   connected,
	}, nil
}

func (s *channelService) Disconnect(ctx context.Context, orgID, id uuid.UUID) error {
	ch, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}

	if err := s.evolutionGateway.Disconnect(ctx, ch.InstanceToken); err != nil {
		return err
	}

	return s.repo.UpdateStatus(ctx, orgID, id, models.StatusDisconnected, nil, nil)
}

func (s *channelService) Reconnect(ctx context.Context, orgID, id uuid.UUID) error {
	ch, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}

	if err := s.evolutionGateway.Reconnect(ctx, ch.InstanceToken); err != nil {
		return err
	}

	return s.repo.UpdateStatus(ctx, orgID, id, models.StatusReconnecting, nil, nil)
}

func (s *channelService) DeleteChannel(ctx context.Context, orgID, id uuid.UUID) error {
	ch, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}

	_ = s.evolutionGateway.DeleteInstance(ctx, ch.InstanceName)
	return s.repo.Delete(ctx, orgID, id)
}

func (s *channelService) PairPhone(ctx context.Context, orgID, id uuid.UUID, phone string) (string, error) {
	ch, err := s.repo.GetByID(ctx, orgID, id)
	if err != nil {
		return "", err
	}

	return s.evolutionGateway.PairPhone(ctx, ch.InstanceToken, phone)
}

