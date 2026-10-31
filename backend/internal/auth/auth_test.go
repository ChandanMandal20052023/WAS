package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"was/backend/internal/middleware"
	"was/backend/internal/models"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretPass123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error hashing password, got: %v", err)
	}

	if hash == password {
		t.Fatal("hash should not equal plain password")
	}

	if !CheckPasswordHash(password, hash) {
		t.Fatal("expected correct password to match hash")
	}

	if CheckPasswordHash("WrongPassword", hash) {
		t.Fatal("expected wrong password to fail match")
	}

	// Short password constraint
	_, err = HashPassword("short")
	if err == nil {
		t.Fatal("expected error on password shorter than 8 characters")
	}
}

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := "test_jwt_secret_key_12345678901234"
	tm := NewTokenManager(secret)

	userID := uuid.New()
	orgID := uuid.New()
	user := &models.User{
		ID:       userID,
		Email:    "test@was-crm.com",
		FullName: "Test User",
	}

	access, refresh, expiresIn, err := tm.GenerateTokenPair(user, orgID, models.RoleOwner)
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	if access == "" || refresh == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}

	if expiresIn <= 0 {
		t.Fatalf("expected positive expiration duration, got: %d", expiresIn)
	}

	// Validate access token
	claims, err := tm.ValidateToken(access)
	if err != nil {
		t.Fatalf("expected valid access token, got: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected user ID %v, got %v", userID, claims.UserID)
	}

	if claims.OrganizationID != orgID {
		t.Errorf("expected org ID %v, got %v", orgID, claims.OrganizationID)
	}

	if claims.Role != models.RoleOwner {
		t.Errorf("expected role %v, got %v", models.RoleOwner, claims.Role)
	}

	if claims.TokenType != "access" {
		t.Errorf("expected token_type access, got %v", claims.TokenType)
	}

	// Tampered token check
	tampered := access + "tampered"
	_, err = tm.ValidateToken(tampered)
	if err == nil {
		t.Fatal("expected error validating tampered token")
	}
}

func TestRBACMiddleware(t *testing.T) {
	secret := "test_jwt_secret_key_12345678901234"
	tm := NewTokenManager(secret)

	user := &models.User{
		ID:       uuid.New(),
		Email:    "agent@was-crm.com",
		FullName: "Agent Smith",
	}
	orgID := uuid.New()

	// 1. Agent token attempting to access Admin-only route
	agentToken, _, _, _ := tm.GenerateTokenPair(user, orgID, models.RoleAgent)

	protectedHandler := middleware.Authenticate(tm)(
		middleware.RequireRoles(models.RoleOwner, models.RoleAdmin)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest("GET", "/api/v1/organizations/current", nil)
	req.Header.Set("Authorization", "Bearer "+agentToken)
	rec := httptest.NewRecorder()

	protectedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403 Forbidden for agent accessing admin endpoint, got %d", rec.Code)
	}

	// 2. Owner token accessing Admin-or-Owner route
	ownerToken, _, _, _ := tm.GenerateTokenPair(user, orgID, models.RoleOwner)
	reqOwner := httptest.NewRequest("GET", "/api/v1/organizations/current", nil)
	reqOwner.Header.Set("Authorization", "Bearer "+ownerToken)
	recOwner := httptest.NewRecorder()

	protectedHandler.ServeHTTP(recOwner, reqOwner)

	if recOwner.Code != http.StatusOK {
		t.Errorf("expected status 200 OK for owner accessing admin endpoint, got %d", recOwner.Code)
	}
}
