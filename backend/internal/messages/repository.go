package messages

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"was/backend/internal/models"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByConversation(ctx context.Context, orgID, convID uuid.UUID, limit, offset int) ([]models.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT 
			m.id, m.organization_id, m.conversation_id, m.whatsapp_message_id,
			m.sender_type, m.sender_user_id, u.full_name,
			m.message_type, m.body, m.media_url, m.media_mime_type,
			m.delivery_status, m.raw_payload, m.sent_at, m.created_at
		FROM messages m
		LEFT JOIN users u ON u.id = m.sender_user_id
		WHERE m.organization_id = $1 AND m.conversation_id = $2
		ORDER BY m.sent_at ASC
		LIMIT $3 OFFSET $4;
	`

	rows, err := r.db.Query(ctx, query, orgID, convID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(
			&m.ID,
			&m.OrganizationID,
			&m.ConversationID,
			&m.WhatsAppMessageID,
			&m.SenderType,
			&m.SenderUserID,
			&m.SenderUserName,
			&m.MessageType,
			&m.Body,
			&m.MediaURL,
			&m.MediaMimeType,
			&m.DeliveryStatus,
			&m.RawPayload,
			&m.SentAt,
			&m.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *Repository) Create(ctx context.Context, msg *models.Message) error {
	query := `
		INSERT INTO messages (
			organization_id, conversation_id, whatsapp_message_id, sender_type,
			sender_user_id, message_type, body, media_url, media_mime_type,
			delivery_status, raw_payload, sent_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW()
		)
		RETURNING id, created_at;
	`
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now()
	}

	return r.db.QueryRow(ctx, query,
		msg.OrganizationID,
		msg.ConversationID,
		msg.WhatsAppMessageID,
		string(msg.SenderType),
		msg.SenderUserID,
		string(msg.MessageType),
		msg.Body,
		msg.MediaURL,
		msg.MediaMimeType,
		string(msg.DeliveryStatus),
		msg.RawPayload,
		msg.SentAt,
	).Scan(&msg.ID, &msg.CreatedAt)
}

func (r *Repository) UpdateDeliveryStatus(ctx context.Context, whatsappMsgID string, status models.DeliveryStatus) error {
	_, err := r.db.Exec(ctx, `
		UPDATE messages
		SET delivery_status = $1
		WHERE whatsapp_message_id = $2;
	`, string(status), whatsappMsgID)
	return err
}

func (r *Repository) ListNotesByConversation(ctx context.Context, orgID, convID uuid.UUID) ([]models.InternalNote, error) {
	query := `
		SELECT n.id, n.organization_id, n.conversation_id, n.contact_id, n.user_id, COALESCE(u.full_name, 'Agent'), n.content, n.created_at, n.updated_at
		FROM internal_notes n
		JOIN users u ON u.id = n.user_id
		WHERE n.organization_id = $1 AND n.conversation_id = $2
		ORDER BY n.created_at ASC;
	`
	rows, err := r.db.Query(ctx, query, orgID, convID)
	if err != nil {
		return nil, fmt.Errorf("failed to query internal notes: %w", err)
	}
	defer rows.Close()

	var notes []models.InternalNote
	for rows.Next() {
		var n models.InternalNote
		if err := rows.Scan(
			&n.ID,
			&n.OrganizationID,
			&n.ConversationID,
			&n.ContactID,
			&n.UserID,
			&n.UserName,
			&n.Content,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *Repository) CreateNote(ctx context.Context, note *models.InternalNote) error {
	query := `
		INSERT INTO internal_notes (organization_id, conversation_id, contact_id, user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at;
	`
	return r.db.QueryRow(ctx, query,
		note.OrganizationID,
		note.ConversationID,
		note.ContactID,
		note.UserID,
		note.Content,
	).Scan(&note.ID, &note.CreatedAt, &note.UpdatedAt)
}

