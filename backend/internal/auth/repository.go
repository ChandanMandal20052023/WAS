package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"was/backend/internal/models"
)

type Repository interface {
	RegisterUserAndOrg(ctx context.Context, user *models.User, orgName string) (*models.User, *models.Organization, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	GetMember(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationMember, error)
	GetUserMemberships(ctx context.Context, userID uuid.UUID) ([]*models.OrganizationMember, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	CreateInvitation(ctx context.Context, invite *models.Invitation) error
	GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error)
	AcceptInvitation(ctx context.Context, invite *models.Invitation, user *models.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) RegisterUserAndOrg(ctx context.Context, user *models.User, orgName string) (*models.User, *models.Organization, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insert User
	userQuery := `
		INSERT INTO users (id, email, password_hash, full_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		RETURNING id, email, full_name, is_active, created_at, updated_at
	`
	userID := uuid.New()
	err = tx.QueryRow(ctx, userQuery, userID, strings.ToLower(strings.TrimSpace(user.Email)), user.PasswordHash, user.FullName).
		Scan(&user.ID, &user.Email, &user.FullName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert user (email may already exist): %w", err)
	}

	// 2. Generate Organization Slug and Insert Organization
	baseSlug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(orgName), " ", "-"))
	slug := fmt.Sprintf("%s-%s", baseSlug, user.ID.String()[:8])

	orgQuery := `
		INSERT INTO organizations (id, name, slug, status, settings, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', '{}'::jsonb, NOW(), NOW())
		RETURNING id, name, slug, status, settings, created_at, updated_at
	`
	org := &models.Organization{}
	orgID := uuid.New()
	err = tx.QueryRow(ctx, orgQuery, orgID, orgName, slug).
		Scan(&org.ID, &org.Name, &org.Slug, &org.Status, &org.Settings, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert organization: %w", err)
	}

	// 3. Insert Organization Member as Owner
	memberQuery := `
		INSERT INTO organization_members (id, organization_id, user_id, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
	`
	_, err = tx.Exec(ctx, memberQuery, uuid.New(), org.ID, user.ID, models.RoleOwner)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to insert organization member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit registration: %w", err)
	}

	return user, org, nil
}

func (r *pgRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, phone, avatar_url, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, strings.ToLower(strings.TrimSpace(email))).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Phone, &user.AvatarURL, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (r *pgRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, phone, avatar_url, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &models.User{}
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Phone, &user.AvatarURL, &user.IsActive, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (r *pgRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, status, settings, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	org := &models.Organization{}
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&org.ID, &org.Name, &org.Slug, &org.Status, &org.Settings, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}
	return org, nil
}

func (r *pgRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*models.OrganizationMember, error) {
	query := `
		SELECT m.id, m.organization_id, m.user_id, m.role, m.is_active, m.created_at, m.updated_at,
		       o.id, o.name, o.slug, o.status, o.settings, o.created_at, o.updated_at
		FROM organization_members m
		JOIN organizations o ON m.organization_id = o.id
		WHERE m.organization_id = $1 AND m.user_id = $2
	`
	member := &models.OrganizationMember{Organization: &models.Organization{}}
	err := r.pool.QueryRow(ctx, query, orgID, userID).
		Scan(&member.ID, &member.OrganizationID, &member.UserID, &member.Role, &member.IsActive, &member.CreatedAt, &member.UpdatedAt,
			&member.Organization.ID, &member.Organization.Name, &member.Organization.Slug, &member.Organization.Status, &member.Organization.Settings, &member.Organization.CreatedAt, &member.Organization.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("organization member not found")
		}
		return nil, err
	}
	return member, nil
}

func (r *pgRepository) GetUserMemberships(ctx context.Context, userID uuid.UUID) ([]*models.OrganizationMember, error) {
	query := `
		SELECT m.id, m.organization_id, m.user_id, m.role, m.is_active, m.created_at, m.updated_at,
		       o.id, o.name, o.slug, o.status, o.settings, o.created_at, o.updated_at
		FROM organization_members m
		JOIN organizations o ON m.organization_id = o.id
		WHERE m.user_id = $1 AND m.is_active = true
		ORDER BY m.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.OrganizationMember
	for rows.Next() {
		member := &models.OrganizationMember{Organization: &models.Organization{}}
		err := rows.Scan(&member.ID, &member.OrganizationID, &member.UserID, &member.Role, &member.IsActive, &member.CreatedAt, &member.UpdatedAt,
			&member.Organization.ID, &member.Organization.Name, &member.Organization.Slug, &member.Organization.Status, &member.Organization.Settings, &member.Organization.CreatedAt, &member.Organization.UpdatedAt)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

func (r *pgRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *pgRepository) CreateInvitation(ctx context.Context, invite *models.Invitation) error {
	query := `
		INSERT INTO invitations (id, organization_id, email, role, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err := r.pool.Exec(ctx, query, uuid.New(), invite.OrganizationID, strings.ToLower(invite.Email), invite.Role, invite.Token, invite.ExpiresAt)
	return err
}

func (r *pgRepository) GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error) {
	query := `
		SELECT id, organization_id, email, role, token, expires_at, accepted_at, created_at
		FROM invitations
		WHERE token = $1
	`
	invite := &models.Invitation{}
	err := r.pool.QueryRow(ctx, query, token).
		Scan(&invite.ID, &invite.OrganizationID, &invite.Email, &invite.Role, &invite.Token, &invite.ExpiresAt, &invite.AcceptedAt, &invite.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return invite, nil
}

func (r *pgRepository) AcceptInvitation(ctx context.Context, invite *models.Invitation, user *models.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert or fetch user
	var existingUserID uuid.UUID
	checkUserQuery := `SELECT id FROM users WHERE email = $1`
	err = tx.QueryRow(ctx, checkUserQuery, strings.ToLower(invite.Email)).Scan(&existingUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		// New user
		insertUserQuery := `
			INSERT INTO users (id, email, password_hash, full_name, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, true, NOW(), NOW())
			RETURNING id
		`
		user.ID = uuid.New()
		err = tx.QueryRow(ctx, insertUserQuery, user.ID, strings.ToLower(invite.Email), user.PasswordHash, user.FullName).Scan(&user.ID)
		if err != nil {
			return err
		}
	} else if err == nil {
		user.ID = existingUserID
	} else {
		return err
	}

	// Add to organization_members
	memberQuery := `
		INSERT INTO organization_members (id, organization_id, user_id, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role = EXCLUDED.role, is_active = true
	`
	_, err = tx.Exec(ctx, memberQuery, uuid.New(), invite.OrganizationID, user.ID, invite.Role)
	if err != nil {
		return err
	}

	// Mark invitation accepted
	markQuery := `UPDATE invitations SET accepted_at = NOW() WHERE id = $1`
	_, err = tx.Exec(ctx, markQuery, invite.ID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *pgRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, passwordHash, userID)
	return err
}
