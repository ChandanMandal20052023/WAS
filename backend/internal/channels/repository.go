package channels

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"was/backend/internal/models"
)

type Repository interface {
	ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*models.WhatsAppChannel, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*models.WhatsAppChannel, error)
	GetByToken(ctx context.Context, token string) (*models.WhatsAppChannel, error)
	Create(ctx context.Context, channel *models.WhatsAppChannel) error
	UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status models.ChannelStatus, phone *string, qrcode *string) error
	UpdateStatusByToken(ctx context.Context, token string, status models.ChannelStatus, phone *string, qrcode *string) error
	Delete(ctx context.Context, orgID, id uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*models.WhatsAppChannel, error) {
	query := `
		SELECT id, organization_id, name, phone_number, instance_name, instance_token,
		       status, qrcode_base64, settings, last_connected_at, created_at, updated_at
		FROM whatsapp_channels
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query channels: %w", err)
	}
	defer rows.Close()

	var channels []*models.WhatsAppChannel
	for rows.Next() {
		ch := &models.WhatsAppChannel{}
		err := rows.Scan(
			&ch.ID, &ch.OrganizationID, &ch.Name, &ch.PhoneNumber, &ch.InstanceName, &ch.InstanceToken,
			&ch.Status, &ch.QRCodeBase64, &ch.Settings, &ch.LastConnectedAt, &ch.CreatedAt, &ch.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func (r *pgRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*models.WhatsAppChannel, error) {
	query := `
		SELECT id, organization_id, name, phone_number, instance_name, instance_token,
		       status, qrcode_base64, settings, last_connected_at, created_at, updated_at
		FROM whatsapp_channels
		WHERE organization_id = $1 AND id = $2
	`
	ch := &models.WhatsAppChannel{}
	err := r.pool.QueryRow(ctx, query, orgID, id).Scan(
		&ch.ID, &ch.OrganizationID, &ch.Name, &ch.PhoneNumber, &ch.InstanceName, &ch.InstanceToken,
		&ch.Status, &ch.QRCodeBase64, &ch.Settings, &ch.LastConnectedAt, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("whatsapp channel not found")
		}
		return nil, err
	}
	return ch, nil
}

func (r *pgRepository) GetByToken(ctx context.Context, token string) (*models.WhatsAppChannel, error) {
	query := `
		SELECT id, organization_id, name, phone_number, instance_name, instance_token,
		       status, qrcode_base64, settings, last_connected_at, created_at, updated_at
		FROM whatsapp_channels
		WHERE instance_token = $1
	`
	ch := &models.WhatsAppChannel{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&ch.ID, &ch.OrganizationID, &ch.Name, &ch.PhoneNumber, &ch.InstanceName, &ch.InstanceToken,
		&ch.Status, &ch.QRCodeBase64, &ch.Settings, &ch.LastConnectedAt, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("whatsapp channel not found")
		}
		return nil, err
	}
	return ch, nil
}

func (r *pgRepository) Create(ctx context.Context, channel *models.WhatsAppChannel) error {
	query := `
		INSERT INTO whatsapp_channels (id, organization_id, name, instance_name, instance_token, status, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	if channel.ID == uuid.Nil {
		channel.ID = uuid.New()
	}
	if channel.Settings == nil {
		channel.Settings = make(map[string]interface{})
	}

	return r.pool.QueryRow(ctx, query,
		channel.ID, channel.OrganizationID, channel.Name, channel.InstanceName, channel.InstanceToken,
		channel.Status, channel.Settings,
	).Scan(&channel.CreatedAt, &channel.UpdatedAt)
}

func (r *pgRepository) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status models.ChannelStatus, phone *string, qrcode *string) error {
	query := `
		UPDATE whatsapp_channels
		SET status = $1,
		    phone_number = COALESCE($2, phone_number),
		    qrcode_base64 = $3,
		    last_connected_at = CASE WHEN $1 = 'Connected' THEN NOW() ELSE last_connected_at END,
		    updated_at = NOW()
		WHERE organization_id = $4 AND id = $5
	`
	_, err := r.pool.Exec(ctx, query, status, phone, qrcode, orgID, id)
	return err
}

func (r *pgRepository) UpdateStatusByToken(ctx context.Context, token string, status models.ChannelStatus, phone *string, qrcode *string) error {
	var now *time.Time
	if status == models.StatusConnected {
		t := time.Now()
		now = &t
	}

	query := `
		UPDATE whatsapp_channels
		SET status = $1,
		    phone_number = COALESCE($2, phone_number),
		    qrcode_base64 = COALESCE($3, qrcode_base64),
		    last_connected_at = COALESCE($4, last_connected_at),
		    updated_at = NOW()
		WHERE instance_token = $5
	`
	_, err := r.pool.Exec(ctx, query, status, phone, qrcode, now, token)
	return err
}

func (r *pgRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	query := `DELETE FROM whatsapp_channels WHERE organization_id = $1 AND id = $2`
	_, err := r.pool.Exec(ctx, query, orgID, id)
	return err
}

