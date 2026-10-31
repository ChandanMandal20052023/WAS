package conversations

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

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetOrCreate finds or atomically creates a conversation thread for an organization, channel, and contact
func (r *Repository) GetOrCreate(ctx context.Context, orgID, channelID, contactID uuid.UUID) (*models.Conversation, error) {
	query := `
		INSERT INTO conversations (organization_id, channel_id, contact_id, status, last_message_time, created_at, updated_at)
		VALUES ($1, $2, $3, 'open', NOW(), NOW(), NOW())
		ON CONFLICT (organization_id, channel_id, contact_id)
		DO UPDATE SET updated_at = NOW()
		RETURNING id, organization_id, channel_id, contact_id, assigned_agent_id, status, last_message_preview, last_message_time, unread_count, created_at, updated_at;
	`
	var conv models.Conversation
	err := r.db.QueryRow(ctx, query, orgID, channelID, contactID).Scan(
		&conv.ID,
		&conv.OrganizationID,
		&conv.ChannelID,
		&conv.ContactID,
		&conv.AssignedAgentID,
		&conv.Status,
		&conv.LastMessagePreview,
		&conv.LastMessageTime,
		&conv.UnreadCount,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create conversation: %w", err)
	}
	return &conv, nil
}

func (r *Repository) GetByID(ctx context.Context, orgID, convID uuid.UUID) (*models.ConversationDetail, error) {
	query := `
		SELECT 
			c.id, c.organization_id, c.channel_id, c.contact_id, c.assigned_agent_id, 
			c.status, c.last_message_preview, c.last_message_time, c.unread_count, c.created_at, c.updated_at,
			COALESCE(ct.full_name, ''), ct.phone_number, COALESCE(ct.avatar_url, ''),
			ch.name,
			u.full_name, u.email
		FROM conversations c
		JOIN contacts ct ON ct.id = c.contact_id
		JOIN whatsapp_channels ch ON ch.id = c.channel_id
		LEFT JOIN users u ON u.id = c.assigned_agent_id
		WHERE c.organization_id = $1 AND c.id = $2;
	`
	var d models.ConversationDetail
	err := r.db.QueryRow(ctx, query, orgID, convID).Scan(
		&d.ID,
		&d.OrganizationID,
		&d.ChannelID,
		&d.ContactID,
		&d.AssignedAgentID,
		&d.Status,
		&d.LastMessagePreview,
		&d.LastMessageTime,
		&d.UnreadCount,
		&d.CreatedAt,
		&d.UpdatedAt,
		&d.ContactName,
		&d.ContactPhone,
		&d.ContactAvatar,
		&d.ChannelName,
		&d.AssignedAgentName,
		&d.AssignedAgentEmail,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("conversation not found")
		}
		return nil, fmt.Errorf("failed to query conversation: %w", err)
	}

	tags, err := r.GetTagsForConversation(ctx, convID)
	if err == nil {
		d.Tags = tags
	} else {
		d.Tags = []models.Tag{}
	}

	return &d, nil
}

func (r *Repository) List(ctx context.Context, orgID uuid.UUID, filter models.ConversationFilter) ([]models.ConversationDetail, int, error) {
	baseQuery := `
		FROM conversations c
		JOIN contacts ct ON ct.id = c.contact_id
		JOIN whatsapp_channels ch ON ch.id = c.channel_id
		LEFT JOIN users u ON u.id = c.assigned_agent_id
		WHERE c.organization_id = $1
	`
	args := []interface{}{orgID}
	paramIdx := 2

	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND c.status = $%d", paramIdx)
		args = append(args, string(*filter.Status))
		paramIdx++
	}

	if filter.ChannelID != nil {
		baseQuery += fmt.Sprintf(" AND c.channel_id = $%d", paramIdx)
		args = append(args, *filter.ChannelID)
		paramIdx++
	}

	if filter.AgentID != nil {
		baseQuery += fmt.Sprintf(" AND c.assigned_agent_id = $%d", paramIdx)
		args = append(args, *filter.AgentID)
		paramIdx++
	}

	if filter.Search != "" {
		baseQuery += fmt.Sprintf(" AND (ct.full_name ILIKE $%d OR ct.phone_number ILIKE $%d OR c.last_message_preview ILIKE $%d)", paramIdx, paramIdx, paramIdx)
		args = append(args, "%"+filter.Search+"%")
		paramIdx++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count conversations: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			c.id, c.organization_id, c.channel_id, c.contact_id, c.assigned_agent_id, 
			c.status, c.last_message_preview, c.last_message_time, c.unread_count, c.created_at, c.updated_at,
			COALESCE(ct.full_name, ''), ct.phone_number, COALESCE(ct.avatar_url, ''),
			ch.name,
			u.full_name, u.email
		%s
		ORDER BY c.last_message_time DESC
		LIMIT $%d OFFSET $%d;
	`, baseQuery, paramIdx, paramIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	var convs []models.ConversationDetail
	for rows.Next() {
		var d models.ConversationDetail
		if err := rows.Scan(
			&d.ID,
			&d.OrganizationID,
			&d.ChannelID,
			&d.ContactID,
			&d.AssignedAgentID,
			&d.Status,
			&d.LastMessagePreview,
			&d.LastMessageTime,
			&d.UnreadCount,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.ContactName,
			&d.ContactPhone,
			&d.ContactAvatar,
			&d.ChannelName,
			&d.AssignedAgentName,
			&d.AssignedAgentEmail,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan conversation: %w", err)
		}
		d.Tags = []models.Tag{}
		convs = append(convs, d)
	}

	for i := range convs {
		tags, _ := r.GetTagsForConversation(ctx, convs[i].ID)
		if tags != nil {
			convs[i].Tags = tags
		}
	}

	return convs, total, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, orgID, convID uuid.UUID, status models.ConversationStatus) error {
	res, err := r.db.Exec(ctx, `
		UPDATE conversations
		SET status = $1, updated_at = NOW()
		WHERE organization_id = $2 AND id = $3;
	`, string(status), orgID, convID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("conversation not found")
	}
	return nil
}

func (r *Repository) AssignAgent(ctx context.Context, orgID, convID uuid.UUID, agentID *uuid.UUID, assignedBy uuid.UUID, reason string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
		UPDATE conversations
		SET assigned_agent_id = $1, updated_at = NOW()
		WHERE organization_id = $2 AND id = $3;
	`, agentID, orgID, convID)
	if err != nil {
		return fmt.Errorf("failed to assign agent: %w", err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("conversation not found")
	}

	if agentID != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO conversation_assignments (organization_id, conversation_id, assigned_to_user_id, assigned_by_user_id, reason, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW());
		`, orgID, convID, *agentID, assignedBy, reason)
		if err != nil {
			return fmt.Errorf("failed to log conversation assignment: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *Repository) ResetUnread(ctx context.Context, orgID, convID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE conversations
		SET unread_count = 0, updated_at = NOW()
		WHERE organization_id = $1 AND id = $2;
	`, orgID, convID)
	return err
}

func (r *Repository) UpdateLastMessage(ctx context.Context, convID uuid.UUID, preview string, messageTime time.Time, incrementUnread bool) error {
	unreadClause := ""
	if incrementUnread {
		unreadClause = ", unread_count = unread_count + 1"
	}
	query := fmt.Sprintf(`
		UPDATE conversations
		SET last_message_preview = $1, last_message_time = $2, updated_at = NOW()%s
		WHERE id = $3;
	`, unreadClause)

	_, err := r.db.Exec(ctx, query, preview, messageTime, convID)
	return err
}

func (r *Repository) GetTagsForConversation(ctx context.Context, convID uuid.UUID) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.organization_id, t.name, t.color_hex, t.created_at
		FROM tags t
		JOIN conversation_tags ct ON ct.tag_id = t.id
		WHERE ct.conversation_id = $1
		ORDER BY t.name ASC;
	`
	rows, err := r.db.Query(ctx, query, convID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	tags := make([]models.Tag, 0)
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.Name, &t.ColorHex, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

