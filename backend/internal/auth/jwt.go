package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"was/backend/internal/middleware"
	"was/backend/internal/models"
)

type Claims struct {
	UserID         uuid.UUID   `json:"user_id"`
	OrganizationID uuid.UUID   `json:"organization_id"`
	Email          string      `json:"email"`
	Role           models.Role `json:"role"`
	TokenType      string      `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type TokenManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{
		secretKey:     []byte(secret),
		accessExpiry:  60 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

// GenerateTokenPair generates both an access token and a refresh token for the user/organization context.
func (tm *TokenManager) GenerateTokenPair(user *models.User, orgID uuid.UUID, role models.Role) (accessToken string, refreshToken string, expiresIn int64, err error) {
	now := time.Now()
	accessExpiresAt := now.Add(tm.accessExpiry)

	accessClaims := &Claims{
		UserID:         user.ID,
		OrganizationID: orgID,
		Email:          user.Email,
		Role:           role,
		TokenType:      "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
			Issuer:    "WAS",
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(tm.secretKey)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshExpiresAt := now.Add(tm.refreshExpiry)
	refreshClaims := &Claims{
		UserID:         user.ID,
		OrganizationID: orgID,
		Email:          user.Email,
		Role:           role,
		TokenType:      "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
			Issuer:    "WAS",
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(tm.secretKey)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	expiresIn = int64(tm.accessExpiry.Seconds())
	return accessToken, refreshToken, expiresIn, nil
}

// ValidateToken parses and validates a JWT token string.
func (tm *TokenManager) ValidateToken(tokenString string) (*middleware.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return &middleware.TokenClaims{
		UserID:         claims.UserID,
		OrganizationID: claims.OrganizationID,
		Email:          claims.Email,
		Role:           claims.Role,
		TokenType:      claims.TokenType,
	}, nil
}
