package auth

import (
	"errors"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

// TokenType enumerates the explicit relay/gateway token categories.
type TokenType string

const (
	TokenTypeRunner  TokenType = "runner"
	TokenTypeBrowser TokenType = "browser"
	TokenTypeTunnel  TokenType = "tunnel"
	TokenTypePreview TokenType = "preview"
)

// RelayClaims represents JWT claims for relay token
// Note: SessionID has been removed - channels are now identified by PodKey only
type RelayClaims struct {
	PodKey        string    `json:"pod_key"`
	RunnerID      int64     `json:"runner_id"`
	UserID        int64     `json:"user_id"` // 0 for runner tokens
	OrgID         int64     `json:"org_id"`
	TokenType     TokenType `json:"token_type,omitempty"`
	PreviewTarget string    `json:"preview_target,omitempty"` // e.g. 127.0.0.1:3000
	PreviewPath   string    `json:"preview_path,omitempty"`
	jwt.RegisteredClaims
}

// IsRunnerToken returns true if this is a runner-issued token (UserID == 0).
func (c *RelayClaims) IsRunnerToken() bool { return c.UserID == 0 }

// IsBrowserToken returns true if this is a browser-issued token (UserID != 0).
func (c *RelayClaims) IsBrowserToken() bool { return c.UserID != 0 }

// ResolvedType resolves the effective token type. When no explicit token_type
// claim is present (legacy tokens), it falls back to the old rule
// (UserID==0 → runner, otherwise browser) for backward compatibility.
func (c *RelayClaims) ResolvedType() TokenType {
	if c.TokenType != "" {
		return c.TokenType
	}
	if c.UserID == 0 {
		return TokenTypeRunner
	}
	return TokenTypeBrowser
}

// TokenValidator validates relay tokens
type TokenValidator struct {
	secretKey []byte
	issuer    string
}

// NewTokenValidator creates a new token validator.
// Panics if secret is empty to prevent validating tokens with a zero-length HMAC key.
func NewTokenValidator(secret, issuer string) *TokenValidator {
	if secret == "" {
		panic("relay token validator secret must not be empty")
	}
	return &TokenValidator{
		secretKey: []byte(secret),
		issuer:    issuer,
	}
}

// ValidateToken validates a relay token and returns claims
func (v *TokenValidator) ValidateToken(tokenString string) (*RelayClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RelayClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return v.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*RelayClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify issuer if configured
	if v.issuer != "" && claims.Issuer != v.issuer {
		return nil, ErrInvalidToken
	}
	if claims.ResolvedType() == TokenTypePreview {
		normalized, err := normalizePreviewPath(claims.PreviewPath)
		if err != nil || normalized != claims.PreviewPath || claims.PreviewTarget == "" {
			return nil, ErrInvalidToken
		}
	}

	return claims, nil
}

func normalizePreviewPath(raw string) (string, error) {
	if raw == "" {
		return "", ErrInvalidToken
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil || !strings.HasPrefix(decoded, "/") || strings.ContainsAny(decoded, "?#") {
		return "", ErrInvalidToken
	}
	for _, segment := range strings.Split(decoded, "/") {
		if segment == ".." {
			return "", ErrInvalidToken
		}
	}
	return path.Clean(decoded), nil
}

// GenerateToken generates a relay token (used by Backend)
// Note: SessionID parameter has been removed - channels are identified by PodKey only
func GenerateToken(secret, issuer, podKey string, runnerID, userID, orgID int64, expiry time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := &RelayClaims{
		PodKey:   podKey,
		RunnerID: runnerID,
		UserID:   userID,
		OrgID:    orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   podKey,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateTypedToken generates a relay token with an explicit token_type
// (and optional preview target). Used by backend and tests; the original
// GenerateToken remains unchanged for legacy callers.
func GenerateTypedToken(secret, issuer string, tokenType TokenType, previewTarget string, runnerID, userID, orgID int64, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &RelayClaims{
		RunnerID:      runnerID,
		UserID:        userID,
		OrgID:         orgID,
		TokenType:     tokenType,
		PreviewTarget: previewTarget,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
