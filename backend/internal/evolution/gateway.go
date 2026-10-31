package evolution

import (
	"context"
)

type EvolutionGateway interface {
	CreateInstance(ctx context.Context, name, token string) error
	ConnectInstance(ctx context.Context, token, webhookURL string) error
	GetQRCode(ctx context.Context, token string) (string, error)
	GetStatus(ctx context.Context, token string) (bool, string, error)
	Disconnect(ctx context.Context, token string) error
	Reconnect(ctx context.Context, token string) error
	DeleteInstance(ctx context.Context, instanceID string) error
	PairPhone(ctx context.Context, token, phone string) (string, error)
	SendTextMessage(ctx context.Context, token, phone, message string) (string, error)
}
