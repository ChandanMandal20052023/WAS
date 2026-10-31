package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"was/backend/internal/models"
	"was/backend/internal/utils"
)

type contextKey string

const (
	UserIDKey         contextKey = "was_user_id"
	OrganizationIDKey contextKey = "was_org_id"
	UserEmailKey      contextKey = "was_user_email"
	UserRoleKey       contextKey = "was_user_role"
)

type TokenClaims struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Email          string
	Role           models.Role
	TokenType      string
}

type TokenValidator interface {
	ValidateToken(tokenString string) (*TokenClaims, error)
}

// Authenticate verifies the JWT Bearer token and injects User and Organization context.
func Authenticate(validator TokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				utils.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format (Bearer required)")
				return
			}

			claims, err := validator.ValidateToken(parts[1])
			if err != nil {
				utils.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Access token is invalid or expired")
				return
			}

			if claims.TokenType != "access" {
				utils.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token is not an access token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, OrganizationIDKey, claims.OrganizationID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRoles enforces that the authenticated user possesses one of the allowed roles.
func RequireRoles(allowedRoles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetUserRole(r.Context())
			if role == "" {
				utils.Error(w, http.StatusForbidden, "FORBIDDEN", "No role assigned in organization context")
				return
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			utils.Error(w, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "You do not have permission to access this resource")
		})
	}
}

// Context Extractors
func GetUserID(ctx context.Context) uuid.UUID {
	if val, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return val
	}
	return uuid.Nil
}

func GetOrganizationID(ctx context.Context) uuid.UUID {
	if val, ok := ctx.Value(OrganizationIDKey).(uuid.UUID); ok {
		return val
	}
	return uuid.Nil
}

func GetUserRole(ctx context.Context) models.Role {
	if val, ok := ctx.Value(UserRoleKey).(models.Role); ok {
		return val
	}
	return ""
}

func GetUserEmail(ctx context.Context) string {
	if val, ok := ctx.Value(UserEmailKey).(string); ok {
		return val
	}
	return ""
}

func GetClaims(ctx context.Context) (*TokenClaims, bool) {
	userID := GetUserID(ctx)
	orgID := GetOrganizationID(ctx)
	if userID == uuid.Nil || orgID == uuid.Nil {
		return nil, false
	}
	return &TokenClaims{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          GetUserEmail(ctx),
		Role:           GetUserRole(ctx),
	}, true
}
