package organizations

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"was/backend/internal/middleware"
	"was/backend/internal/models"
	"was/backend/internal/utils"
)

type MemberWithWorkload struct {
	*models.OrganizationMember
	OpenConversations     int `json:"open_conversations"`
	PendingConversations  int `json:"pending_conversations"`
	ResolvedConversations int `json:"resolved_conversations"`
}

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	GetMembers(ctx context.Context, orgID uuid.UUID) ([]*models.OrganizationMember, error)
	GetMembersWithWorkload(ctx context.Context, orgID uuid.UUID) ([]*MemberWithWorkload, error)
	UpdateMemberRole(ctx context.Context, orgID, memberID uuid.UUID, role models.Role) error
	UpdateMemberStatus(ctx context.Context, orgID, memberID uuid.UUID, isActive bool) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, status, settings, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	org := &models.Organization{}
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&org.ID, &org.Name, &org.Slug, &org.Status, &org.Settings, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return nil, errors.New("organization not found")
	}
	return org, nil
}

func (r *pgRepository) Update(ctx context.Context, org *models.Organization) error {
	query := `
		UPDATE organizations
		SET name = $1, settings = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.pool.Exec(ctx, query, org.Name, org.Settings, org.ID)
	return err
}

func (r *pgRepository) GetMembers(ctx context.Context, orgID uuid.UUID) ([]*models.OrganizationMember, error) {
	query := `
		SELECT m.id, m.organization_id, m.user_id, m.role, m.is_active, m.created_at, m.updated_at,
		       u.id, u.email, u.full_name, u.phone, u.avatar_url, u.is_active, u.created_at, u.updated_at
		FROM organization_members m
		JOIN users u ON m.user_id = u.id
		WHERE m.organization_id = $1
		ORDER BY m.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.OrganizationMember
	for rows.Next() {
		member := &models.OrganizationMember{User: &models.User{}}
		err := rows.Scan(
			&member.ID, &member.OrganizationID, &member.UserID, &member.Role, &member.IsActive, &member.CreatedAt, &member.UpdatedAt,
			&member.User.ID, &member.User.Email, &member.User.FullName, &member.User.Phone, &member.User.AvatarURL, &member.User.IsActive, &member.User.CreatedAt, &member.User.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

func (r *pgRepository) GetMembersWithWorkload(ctx context.Context, orgID uuid.UUID) ([]*MemberWithWorkload, error) {
	members, err := r.GetMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	result := make([]*MemberWithWorkload, 0, len(members))
	for _, m := range members {
		mw := &MemberWithWorkload{OrganizationMember: m}
		// Query open, pending, resolved counts
		rows, err := r.pool.Query(ctx, `
			SELECT status, COUNT(*)
			FROM conversations
			WHERE organization_id = $1 AND assigned_agent_id = $2
			GROUP BY status
		`, orgID, m.UserID)
		if err == nil {
			for rows.Next() {
				var status string
				var count int
				if err := rows.Scan(&status, &count); err == nil {
					switch status {
					case "open":
						mw.OpenConversations = count
					case "pending":
						mw.PendingConversations = count
					case "resolved":
						mw.ResolvedConversations = count
					}
				}
			}
			rows.Close()
		}
		result = append(result, mw)
	}

	return result, nil
}

func (r *pgRepository) UpdateMemberRole(ctx context.Context, orgID, memberID uuid.UUID, role models.Role) error {
	res, err := r.pool.Exec(ctx, `
		UPDATE organization_members
		SET role = $1, updated_at = NOW()
		WHERE organization_id = $2 AND id = $3;
	`, string(role), orgID, memberID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("member not found")
	}
	return nil
}

func (r *pgRepository) UpdateMemberStatus(ctx context.Context, orgID, memberID uuid.UUID, isActive bool) error {
	res, err := r.pool.Exec(ctx, `
		UPDATE organization_members
		SET is_active = $1, updated_at = NOW()
		WHERE organization_id = $2 AND id = $3;
	`, isActive, orgID, memberID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("member not found")
	}
	return nil
}

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	if orgID == uuid.Nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ORG", "Organization context missing")
		return
	}

	org, err := h.repo.GetByID(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, org)
}

func (h *Handler) UpdateCurrent(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())

	var body struct {
		Name     string                 `json:"name"`
		Settings map[string]interface{} `json:"settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	org, err := h.repo.GetByID(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusNotFound, "NOT_FOUND", "Organization not found")
		return
	}

	if body.Name != "" {
		org.Name = body.Name
	}
	if body.Settings != nil {
		org.Settings = body.Settings
	}

	if err := h.repo.Update(r.Context(), org); err != nil {
		utils.Error(w, http.StatusInternalServerError, "UPDATE_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, org)
}

func (h *Handler) GetMembers(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())

	members, err := h.repo.GetMembersWithWorkload(r.Context(), orgID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "FETCH_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, members)
}

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	memberIDStr := chi.URLParam(r, "id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid member id")
		return
	}

	var body struct {
		Role models.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	switch body.Role {
	case models.RoleOwner, models.RoleAdmin, models.RoleManager, models.RoleAgent:
		// valid
	default:
		utils.Error(w, http.StatusBadRequest, "INVALID_ROLE", "Role must be owner, admin, manager, or agent")
		return
	}

	if err := h.repo.UpdateMemberRole(r.Context(), orgID, memberID, body.Role); err != nil {
		utils.Error(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "role updated successfully"})
}

func (h *Handler) UpdateMemberStatus(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrganizationID(r.Context())
	memberIDStr := chi.URLParam(r, "id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid member id")
		return
	}

	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.Error(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if err := h.repo.UpdateMemberStatus(r.Context(), orgID, memberID, body.IsActive); err != nil {
		utils.Error(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "member status updated successfully"})
}

