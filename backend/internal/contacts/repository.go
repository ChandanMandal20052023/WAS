package contacts

import (
	"context"
	"errors"
	"fmt"

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

// FindOrCreateByPhone atomically looks up or creates a contact by normalized phone number
func (r *Repository) FindOrCreateByPhone(ctx context.Context, orgID uuid.UUID, phone string, name string) (*models.Contact, error) {
	query := `
		INSERT INTO contacts (organization_id, phone_number, full_name, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (organization_id, phone_number)
		DO UPDATE SET
			full_name = CASE WHEN contacts.full_name IS NULL OR contacts.full_name = '' THEN EXCLUDED.full_name ELSE contacts.full_name END,
			updated_at = NOW()
		RETURNING id, organization_id, phone_number, COALESCE(full_name, ''), COALESCE(email, ''), COALESCE(avatar_url, ''), metadata, created_at, updated_at;
	`
	var c models.Contact
	err := r.db.QueryRow(ctx, query, orgID, phone, name).Scan(
		&c.ID,
		&c.OrganizationID,
		&c.PhoneNumber,
		&c.FullName,
		&c.Email,
		&c.AvatarURL,
		&c.Metadata,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert contact by phone: %w", err)
	}
	return &c, nil
}

func (r *Repository) GetByID(ctx context.Context, orgID, contactID uuid.UUID) (*models.ContactWithTags, error) {
	query := `
		SELECT id, organization_id, phone_number, COALESCE(full_name, ''), COALESCE(email, ''), COALESCE(avatar_url, ''), metadata, created_at, updated_at
		FROM contacts
		WHERE organization_id = $1 AND id = $2;
	`
	var c models.ContactWithTags
	err := r.db.QueryRow(ctx, query, orgID, contactID).Scan(
		&c.ID,
		&c.OrganizationID,
		&c.PhoneNumber,
		&c.FullName,
		&c.Email,
		&c.AvatarURL,
		&c.Metadata,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("contact not found")
		}
		return nil, fmt.Errorf("failed to query contact: %w", err)
	}

	tags, err := r.GetTagsForContact(ctx, contactID)
	if err != nil {
		return nil, err
	}
	c.Tags = tags

	return &c, nil
}

func (r *Repository) List(ctx context.Context, orgID uuid.UUID, filter models.ContactFilter) ([]models.ContactWithTags, int, error) {
	baseQuery := `
		FROM contacts c
		WHERE c.organization_id = $1
	`
	args := []interface{}{orgID}
	paramIdx := 2

	if filter.Search != "" {
		baseQuery += fmt.Sprintf(" AND (c.full_name ILIKE $%d OR c.phone_number ILIKE $%d OR c.email ILIKE $%d)", paramIdx, paramIdx, paramIdx)
		args = append(args, "%"+filter.Search+"%")
		paramIdx++
	}

	if filter.TagID != nil {
		baseQuery += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM contact_tags ct WHERE ct.contact_id = c.id AND ct.tag_id = $%d)", paramIdx)
		args = append(args, *filter.TagID)
		paramIdx++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count contacts: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT c.id, c.organization_id, c.phone_number, COALESCE(c.full_name, ''), COALESCE(c.email, ''), COALESCE(c.avatar_url, ''), c.metadata, c.created_at, c.updated_at
		%s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d;
	`, baseQuery, paramIdx, paramIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query contacts list: %w", err)
	}
	defer rows.Close()

	var contacts []models.ContactWithTags
	for rows.Next() {
		var c models.ContactWithTags
		if err := rows.Scan(
			&c.ID,
			&c.OrganizationID,
			&c.PhoneNumber,
			&c.FullName,
			&c.Email,
			&c.AvatarURL,
			&c.Metadata,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan contact: %w", err)
		}
		c.Tags = []models.Tag{}
		contacts = append(contacts, c)
	}

	// Fetch tags for this batch of contacts
	for i := range contacts {
		tags, err := r.GetTagsForContact(ctx, contacts[i].ID)
		if err == nil {
			contacts[i].Tags = tags
		}
	}

	return contacts, total, nil
}

func (r *Repository) Create(ctx context.Context, orgID uuid.UUID, req models.CreateContactRequest) (*models.ContactWithTags, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO contacts (organization_id, phone_number, full_name, email, avatar_url, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, '{}'::jsonb, NOW(), NOW())
		RETURNING id, organization_id, phone_number, COALESCE(full_name, ''), COALESCE(email, ''), COALESCE(avatar_url, ''), metadata, created_at, updated_at;
	`
	var c models.ContactWithTags
	err = tx.QueryRow(ctx, query, orgID, req.PhoneNumber, req.FullName, req.Email, req.AvatarURL).Scan(
		&c.ID,
		&c.OrganizationID,
		&c.PhoneNumber,
		&c.FullName,
		&c.Email,
		&c.AvatarURL,
		&c.Metadata,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert contact: %w", err)
	}

	// Insert tags
	for _, tagID := range req.TagIDs {
		_, err := tx.Exec(ctx, `
			INSERT INTO contact_tags (contact_id, tag_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT DO NOTHING;
		`, c.ID, tagID)
		if err != nil {
			return nil, fmt.Errorf("failed to associate tag: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	tags, _ := r.GetTagsForContact(ctx, c.ID)
	c.Tags = tags
	return &c, nil
}

func (r *Repository) Update(ctx context.Context, orgID, contactID uuid.UUID, req models.UpdateContactRequest) (*models.ContactWithTags, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE contacts
		SET full_name = $1, email = $2, avatar_url = $3, updated_at = NOW()
		WHERE organization_id = $4 AND id = $5
		RETURNING id, organization_id, phone_number, COALESCE(full_name, ''), COALESCE(email, ''), COALESCE(avatar_url, ''), metadata, created_at, updated_at;
	`
	var c models.ContactWithTags
	err = tx.QueryRow(ctx, query, req.FullName, req.Email, req.AvatarURL, orgID, contactID).Scan(
		&c.ID,
		&c.OrganizationID,
		&c.PhoneNumber,
		&c.FullName,
		&c.Email,
		&c.AvatarURL,
		&c.Metadata,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("contact not found")
		}
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	// Update tags if provided
	if req.TagIDs != nil {
		_, _ = tx.Exec(ctx, `DELETE FROM contact_tags WHERE contact_id = $1`, contactID)
		for _, tagID := range req.TagIDs {
			_, err := tx.Exec(ctx, `
				INSERT INTO contact_tags (contact_id, tag_id, created_at)
				VALUES ($1, $2, NOW())
				ON CONFLICT DO NOTHING;
			`, contactID, tagID)
			if err != nil {
				return nil, fmt.Errorf("failed to associate tag: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	tags, _ := r.GetTagsForContact(ctx, c.ID)
	c.Tags = tags
	return &c, nil
}

func (r *Repository) Delete(ctx context.Context, orgID, contactID uuid.UUID) error {
	res, err := r.db.Exec(ctx, `
		DELETE FROM contacts
		WHERE organization_id = $1 AND id = $2;
	`, orgID, contactID)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("contact not found")
	}
	return nil
}

func (r *Repository) GetTagsForContact(ctx context.Context, contactID uuid.UUID) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.organization_id, t.name, t.color_hex, t.created_at
		FROM tags t
		JOIN contact_tags ct ON ct.tag_id = t.id
		WHERE ct.contact_id = $1
		ORDER BY t.name ASC;
	`
	rows, err := r.db.Query(ctx, query, contactID)
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

func (r *Repository) ListTags(ctx context.Context, orgID uuid.UUID) ([]models.Tag, error) {
	query := `
		SELECT id, organization_id, name, color_hex, created_at
		FROM tags
		WHERE organization_id = $1
		ORDER BY name ASC;
	`
	rows, err := r.db.Query(ctx, query, orgID)
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

func (r *Repository) CreateTag(ctx context.Context, orgID uuid.UUID, name, color string) (*models.Tag, error) {
	if color == "" {
		color = "#128C7E"
	}
	query := `
		INSERT INTO tags (organization_id, name, color_hex, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (organization_id, name)
		DO UPDATE SET color_hex = EXCLUDED.color_hex
		RETURNING id, organization_id, name, color_hex, created_at;
	`
	var t models.Tag
	err := r.db.QueryRow(ctx, query, orgID, name, color).Scan(
		&t.ID,
		&t.OrganizationID,
		&t.Name,
		&t.ColorHex,
		&t.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}
	return &t, nil
}
