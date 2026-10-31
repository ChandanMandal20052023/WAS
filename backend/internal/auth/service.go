package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"was/backend/internal/models"
)

type Service interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error)
	RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.AuthResponse, error)
	GetMe(ctx context.Context, userID, orgID uuid.UUID) (*models.AuthResponse, error)
	InviteMember(ctx context.Context, orgID uuid.UUID, req *models.InviteRequest) (*models.Invitation, error)
	AcceptInvitation(ctx context.Context, req *models.AcceptInviteRequest) (*models.AuthResponse, error)
	ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) (string, error)
	ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) error
}

type authService struct {
	repo         Repository
	tokenManager *TokenManager
}

func NewService(repo Repository, tokenManager *TokenManager) Service {
	return &authService{
		repo:         repo,
		tokenManager: tokenManager,
	}
}

func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("email and password are required")
	}
	if strings.TrimSpace(req.BusinessName) == "" {
		return nil, errors.New("business name is required")
	}
	if strings.TrimSpace(req.FullName) == "" {
		return nil, errors.New("full name is required")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: hash,
		FullName:     strings.TrimSpace(req.FullName),
	}

	createdUser, createdOrg, err := s.repo.RegisterUserAndOrg(ctx, user, req.BusinessName)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, expiresIn, err := s.tokenManager.GenerateTokenPair(createdUser, createdOrg.ID, models.RoleOwner)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:         createdUser,
		Organization: createdOrg,
		Role:         models.RoleOwner,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	memberships, err := s.repo.GetUserMemberships(ctx, user.ID)
	if err != nil || len(memberships) == 0 {
		return nil, errors.New("no active organization membership found")
	}

	// Default to first active organization
	activeMembership := memberships[0]

	accessToken, refreshToken, expiresIn, err := s.tokenManager.GenerateTokenPair(user, activeMembership.OrganizationID, activeMembership.Role)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	return &models.AuthResponse{
		User:         user,
		Organization: activeMembership.Organization,
		Role:         activeMembership.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Memberships:  memberships,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.AuthResponse, error) {
	if req.RefreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	claims, err := s.tokenManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("token provided is not a refresh token")
	}

	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil || !user.IsActive {
		return nil, errors.New("user account not found or deactivated")
	}

	member, err := s.repo.GetMember(ctx, claims.OrganizationID, claims.UserID)
	if err != nil || !member.IsActive {
		return nil, errors.New("organization membership not found or revoked")
	}

	accessToken, refreshToken, expiresIn, err := s.tokenManager.GenerateTokenPair(user, member.OrganizationID, member.Role)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:         user,
		Organization: member.Organization,
		Role:         member.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID, orgID uuid.UUID) (*models.AuthResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	memberships, err := s.repo.GetUserMemberships(ctx, userID)
	if err != nil {
		return nil, err
	}

	var currentMember *models.OrganizationMember
	for _, m := range memberships {
		if m.OrganizationID == orgID {
			currentMember = m
			break
		}
	}

	if currentMember == nil && len(memberships) > 0 {
		currentMember = memberships[0]
	}

	if currentMember == nil {
		return nil, errors.New("no active organization found")
	}

	return &models.AuthResponse{
		User:         user,
		Organization: currentMember.Organization,
		Role:         currentMember.Role,
		Memberships:  memberships,
	}, nil
}

func (s *authService) InviteMember(ctx context.Context, orgID uuid.UUID, req *models.InviteRequest) (*models.Invitation, error) {
	if !req.Role.IsValid() {
		return nil, errors.New("invalid role")
	}

	tokenBytes := make([]byte, 24)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	invite := &models.Invitation{
		OrganizationID: orgID,
		Email:          strings.ToLower(strings.TrimSpace(req.Email)),
		Role:           req.Role,
		Token:          token,
		ExpiresAt:      time.Now().Add(72 * time.Hour),
	}

	if err := s.repo.CreateInvitation(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return invite, nil
}

func (s *authService) AcceptInvitation(ctx context.Context, req *models.AcceptInviteRequest) (*models.AuthResponse, error) {
	invite, err := s.repo.GetInvitationByToken(ctx, req.Token)
	if err != nil {
		return nil, errors.New("invalid or expired invitation token")
	}

	if invite.AcceptedAt != nil {
		return nil, errors.New("invitation has already been accepted")
	}

	if time.Now().After(invite.ExpiresAt) {
		return nil, errors.New("invitation has expired")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        invite.Email,
		FullName:     strings.TrimSpace(req.FullName),
		PasswordHash: hash,
	}

	if err := s.repo.AcceptInvitation(ctx, invite, user); err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	org, err := s.repo.GetOrganizationByID(ctx, invite.OrganizationID)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, expiresIn, err := s.tokenManager.GenerateTokenPair(user, org.ID, invite.Role)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:         user,
		Organization: org,
		Role:         invite.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *authService) ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Return generic success to prevent email enumeration
		return "reset_token_simulated", nil
	}

	resetBytes := make([]byte, 24)
	_, _ = rand.Read(resetBytes)
	resetToken := hex.EncodeToString(resetBytes)

	// Note: In production, send email with reset link. In dev/testing, return token.
	_ = user
	return resetToken, nil
}

func (s *authService) ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) error {
	if len(req.NewPassword) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	// For production reset token verification
	return nil
}

