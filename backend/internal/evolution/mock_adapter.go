package evolution

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type mockInstance struct {
	Name      string
	Token     string
	Connected bool
	JID       string
}

type MockAdapter struct {
	mu        sync.RWMutex
	instances map[string]*mockInstance
}

func NewMockAdapter() *MockAdapter {
	return &MockAdapter{
		instances: make(map[string]*mockInstance),
	}
}

func (m *MockAdapter) CreateInstance(ctx context.Context, name, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.instances[token] = &mockInstance{
		Name:      name,
		Token:     token,
		Connected: false,
		JID:       "",
	}
	return nil
}

func (m *MockAdapter) ConnectInstance(ctx context.Context, token, webhookURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.instances[token]; !ok {
		m.instances[token] = &mockInstance{Token: token}
	}
	return nil
}

func (m *MockAdapter) GetQRCode(ctx context.Context, token string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Generate a clean, authentic SVG QR code representation encoded in base64 data URI
	svgQR := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="256" height="256">
		<rect width="256" height="256" fill="#FFFFFF"/>
		<rect x="24" y="24" width="64" height="64" fill="#000000" rx="8"/>
		<rect x="36" y="36" width="40" height="40" fill="#FFFFFF" rx="4"/>
		<rect x="44" y="44" width="24" height="24" fill="#128C7E"/>
		<rect x="168" y="24" width="64" height="64" fill="#000000" rx="8"/>
		<rect x="180" y="36" width="40" height="40" fill="#FFFFFF" rx="4"/>
		<rect x="188" y="44" width="24" height="24" fill="#128C7E"/>
		<rect x="24" y="168" width="64" height="64" fill="#000000" rx="8"/>
		<rect x="36" y="180" width="40" height="40" fill="#FFFFFF" rx="4"/>
		<rect x="44" y="188" width="24" height="24" fill="#128C7E"/>
		<circle cx="128" cy="128" r="28" fill="#128C7E"/>
		<text x="128" y="134" font-family="sans-serif" font-size="16" font-weight="bold" fill="#FFFFFF" text-anchor="middle">WAS</text>
	</svg>`)

	encoded := base64.StdEncoding.EncodeToString([]byte(svgQR))
	return fmt.Sprintf("data:image/svg+xml;base64,%s", encoded), nil
}

func (m *MockAdapter) GetStatus(ctx context.Context, token string) (bool, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inst, ok := m.instances[token]
	if !ok {
		return false, "", nil
	}
	return inst.Connected, inst.JID, nil
}

func (m *MockAdapter) Disconnect(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if inst, ok := m.instances[token]; ok {
		inst.Connected = false
		inst.JID = ""
	}
	return nil
}

func (m *MockAdapter) Reconnect(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if inst, ok := m.instances[token]; ok {
		inst.Connected = true
		inst.JID = "15552345678@s.whatsapp.net"
	}
	return nil
}

func (m *MockAdapter) DeleteInstance(ctx context.Context, instanceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.instances, instanceID)
	return nil
}

func (m *MockAdapter) PairPhone(ctx context.Context, token, phone string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return "WAS-7829-4102", nil
}

func (m *MockAdapter) SendTextMessage(ctx context.Context, token, phone, message string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return fmt.Sprintf("MOCK-WA-%s", uuid.New().String()[:8]), nil
}

// SetConnected allows test suites to toggle connected status
func (m *MockAdapter) SetConnected(token string, connected bool, jid string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if inst, ok := m.instances[token]; ok {
		inst.Connected = connected
		inst.JID = jid
	} else {
		m.instances[token] = &mockInstance{Token: token, Connected: connected, JID: jid}
	}
}
