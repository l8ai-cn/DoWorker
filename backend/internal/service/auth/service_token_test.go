package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewService(t *testing.T) {
	cfg := &Config{
		JWTExpiration:     time.Hour,
		RefreshExpiration: time.Hour * 24 * 7,
		Issuer:            "test-issuer",
	}
	configureTestAccessTokens(t, cfg)

	svc := NewService(cfg, nil)
	if svc == nil {
		t.Error("NewService returned nil")
	}
	if svc.config != cfg {
		t.Error("Service config not set correctly")
	}
}

func TestGenerateTokenPair(t *testing.T) {
	cfg := &Config{
		JWTExpiration:     time.Hour,
		RefreshExpiration: time.Hour * 24 * 7,
		Issuer:            "test-issuer",
	}
	configureTestAccessTokens(t, cfg)

	svc := NewService(cfg, nil)
	mockUser := createMockUser()

	tests := []struct {
		name  string
		orgID int64
		role  string
	}{
		{
			name:  "basic token generation",
			orgID: 0,
			role:  "",
		},
		{
			name:  "token with org and role",
			orgID: 123,
			role:  "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := svc.GenerateTokenPair(mockUser, tt.orgID, tt.role)
			if err != nil {
				t.Fatalf("GenerateTokenPair failed: %v", err)
			}

			if tokens.AccessToken == "" {
				t.Error("AccessToken is empty")
			}
			if tokens.RefreshToken == "" {
				t.Error("RefreshToken is empty")
			}
			if tokens.TokenType != "Bearer" {
				t.Errorf("TokenType = %s, want Bearer", tokens.TokenType)
			}
			if tokens.ExpiresAt.Before(time.Now()) {
				t.Error("ExpiresAt should be in the future")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	cfg := &Config{
		JWTExpiration:     time.Hour,
		RefreshExpiration: time.Hour * 24 * 7,
		Issuer:            "test-issuer",
	}
	fixture := configureTestAccessTokens(t, cfg)

	svc := NewService(cfg, nil)
	mockUser := createMockUser()

	t.Run("valid token", func(t *testing.T) {
		tokens, err := svc.GenerateTokenPair(mockUser, 123, "admin")
		if err != nil {
			t.Fatalf("Failed to generate tokens: %v", err)
		}

		claims, err := svc.ValidateToken(tokens.AccessToken)
		if err != nil {
			t.Fatalf("ValidateToken failed: %v", err)
		}

		if claims.UserID != mockUser.ID {
			t.Errorf("UserID = %d, want %d", claims.UserID, mockUser.ID)
		}
		if claims.Email != mockUser.Email {
			t.Errorf("Email = %s, want %s", claims.Email, mockUser.Email)
		}
		if claims.Username != mockUser.Username {
			t.Errorf("Username = %s, want %s", claims.Username, mockUser.Username)
		}
		if claims.OrganizationID != 123 {
			t.Errorf("OrganizationID = %d, want 123", claims.OrganizationID)
		}
		if claims.Role != "admin" {
			t.Errorf("Role = %s, want admin", claims.Role)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := svc.ValidateToken("invalid-token")
		if err == nil {
			t.Error("Expected error for invalid token")
		}
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := svc.ValidateToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.invalid")
		if err == nil {
			t.Error("Expected error for malformed token")
		}
	})

	t.Run("token with wrong secret", func(t *testing.T) {
		otherCfg := &Config{
			JWTExpiration: time.Hour,
			Issuer:        "test-issuer",
		}
		configureTestAccessTokens(t, otherCfg)
		otherSvc := NewService(otherCfg, nil)
		tokens, _ := otherSvc.GenerateTokenPair(mockUser, 0, "")

		_, err := svc.ValidateToken(tokens.AccessToken)
		if err == nil {
			t.Error("Expected error for token with wrong secret")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expired := signExpiredTestAccessToken(t, fixture, mockUser.ID)
		_, err := svc.ValidateToken(expired)
		if err == nil {
			t.Error("Expected error for expired token")
		}
		if err != ErrTokenExpired {
			t.Errorf("Expected ErrTokenExpired, got %v", err)
		}
	})

	t.Run("embed token use", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
			UserID:   mockUser.ID,
			Email:    mockUser.Email,
			Username: mockUser.Username,
			TokenUse: "agent_embed_session",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Issuer:    "test-issuer",
				Audience:  jwt.ClaimStrings{testAccessTokenAudience},
			},
		})
		token.Header["kid"] = "test-access-token-key"
		signed, err := token.SignedString(fixture.privateKey)
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}

		_, err = svc.ValidateToken(signed)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})
}
