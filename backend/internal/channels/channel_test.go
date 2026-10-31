package channels

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"was/backend/internal/evolution"
	"was/backend/internal/models"
)

type memoryChannelRepo struct {
	channels map[uuid.UUID]*models.WhatsAppChannel
}

func newMemoryChannelRepo() *memoryChannelRepo {
	return &memoryChannelRepo{channels: make(map[uuid.UUID]*models.WhatsAppChannel)}
}

func (m *memoryChannelRepo) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*models.WhatsAppChannel, error) {
	var list []*models.WhatsAppChannel
	for _, ch := range m.channels {
		if ch.OrganizationID == orgID {
			list = append(list, ch)
		}
	}
	return list, nil
}

func (m *memoryChannelRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*models.WhatsAppChannel, error) {
	ch, ok := m.channels[id]
	if !ok || ch.OrganizationID != orgID {
		return nil, models.ErrNotFound
	}
	return ch, nil
}

func (m *memoryChannelRepo) GetByToken(ctx context.Context, token string) (*models.WhatsAppChannel, error) {
	for _, ch := range m.channels {
		if ch.InstanceToken == token {
			return ch, nil
		}
	}
	return nil, models.ErrNotFound
}

func (m *memoryChannelRepo) Create(ctx context.Context, channel *models.WhatsAppChannel) error {
	m.channels[channel.ID] = channel
	return nil
}

func (m *memoryChannelRepo) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status models.ChannelStatus, phone *string, qrcode *string) error {
	ch, ok := m.channels[id]
	if !ok || ch.OrganizationID != orgID {
		return models.ErrNotFound
	}
	ch.Status = status
	if phone != nil {
		ch.PhoneNumber = phone
	}
	if qrcode != nil {
		ch.QRCodeBase64 = qrcode
	}
	return nil
}

func (m *memoryChannelRepo) UpdateStatusByToken(ctx context.Context, token string, status models.ChannelStatus, phone *string, qrcode *string) error {
	for _, ch := range m.channels {
		if ch.InstanceToken == token {
			ch.Status = status
			if phone != nil {
				ch.PhoneNumber = phone
			}
			if qrcode != nil {
				ch.QRCodeBase64 = qrcode
			}
			return nil
		}
	}
	return models.ErrNotFound
}

func (m *memoryChannelRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	delete(m.channels, id)
	return nil
}

func TestChannelLifecycleWithMockAdapter(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()

	repo := newMemoryChannelRepo()
	mockGateway := evolution.NewMockAdapter()
	service := NewService(repo, mockGateway, "http://localhost:8080")

	// 1. Create channel
	req := &models.CreateChannelRequest{Name: "Sales Support"}
	channel, err := service.CreateChannel(ctx, orgID, req)
	if err != nil {
		t.Fatalf("failed to create channel: %v", err)
	}

	if channel.Name != "Sales Support" {
		t.Errorf("expected channel name 'Sales Support', got '%s'", channel.Name)
	}
	if channel.Status != models.StatusWaitingForQRScan {
		t.Errorf("expected status 'Waiting for QR Scan', got '%s'", channel.Status)
	}

	// 2. Fetch QR Code
	qrResp, err := service.GetQRCode(ctx, orgID, channel.ID)
	if err != nil {
		t.Fatalf("failed to get QR code: %v", err)
	}

	if !strings.HasPrefix(qrResp.QRCode, "data:image/svg+xml;base64,") {
		t.Errorf("expected base64 SVG QR code data URI, got: %s", qrResp.QRCode)
	}

	// 3. Status sync before connection
	statusBefore, err := service.SyncStatus(ctx, orgID, channel.ID)
	if err != nil {
		t.Fatalf("failed to sync status: %v", err)
	}
	if statusBefore.Connected {
		t.Fatal("expected channel not to be connected initially")
	}

	// 4. Simulate WhatsApp scan and connection via gateway
	mockGateway.SetConnected(channel.InstanceToken, true, "15551234567@s.whatsapp.net")

	statusAfter, err := service.SyncStatus(ctx, orgID, channel.ID)
	if err != nil {
		t.Fatalf("failed to sync status after scan: %v", err)
	}
	if !statusAfter.Connected {
		t.Fatal("expected channel to report connected after scan")
	}
	if statusAfter.Status != models.StatusConnected {
		t.Errorf("expected status 'Connected', got '%s'", statusAfter.Status)
	}

	// 5. Test Disconnect
	err = service.Disconnect(ctx, orgID, channel.ID)
	if err != nil {
		t.Fatalf("failed to disconnect: %v", err)
	}

	ch, _ := service.GetChannel(ctx, orgID, channel.ID)
	if ch.Status != models.StatusDisconnected {
		t.Errorf("expected status 'Disconnected', got '%s'", ch.Status)
	}
}

