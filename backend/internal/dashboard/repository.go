package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Metrics struct {
	TotalContacts       int `json:"total_contacts"`
	TotalConversations  int `json:"total_conversations"`
	OpenConversations   int `json:"open_conversations"`
	PendingConversations int `json:"pending_conversations"`
	ResolvedConversations int `json:"resolved_conversations"`
	UnreadMessages      int `json:"unread_messages"`
	ConnectedChannels   int `json:"connected_channels"`
	ActiveAgents        int `json:"active_agents"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetMetrics(ctx context.Context, orgID uuid.UUID) (*Metrics, error) {
	m := &Metrics{}

	// 1. Total Contacts
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM contacts WHERE organization_id = $1`, orgID).Scan(&m.TotalContacts)

	// 2. Conversation counts by status & unread messages sum
	rows, err := r.db.Query(ctx, `
		SELECT status, COUNT(*), COALESCE(SUM(unread_count), 0)
		FROM conversations
		WHERE organization_id = $1
		GROUP BY status
	`, orgID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			var unread int
			if err := rows.Scan(&status, &count, &unread); err == nil {
				m.TotalConversations += count
				m.UnreadMessages += unread
				switch status {
				case "open":
					m.OpenConversations = count
				case "pending":
					m.PendingConversations = count
				case "resolved":
					m.ResolvedConversations = count
				}
			}
		}
	}

	// 3. Connected Channels
	_ = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM whatsapp_channels
		WHERE organization_id = $1 AND status = 'Connected'
	`, orgID).Scan(&m.ConnectedChannels)

	// 4. Active Agents
	_ = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM organization_members
		WHERE organization_id = $1
	`, orgID).Scan(&m.ActiveAgents)

	return m, nil
}
