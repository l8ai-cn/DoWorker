package auth

import (
	"errors"
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
	TokenTypeRunner           TokenType = "runner"
	TokenTypeBrowser          TokenType = "browser"
	TokenTypeTunnel           TokenType = "tunnel"
	TokenTypePreviewBootstrap TokenType = "preview_bootstrap"
	TokenTypePreviewSession   TokenType = "preview_session"
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
	PreviewOrigin string    `json:"preview_origin,omitempty"`
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
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return v.secretKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())

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
	if claims.ResolvedType() == TokenTypePreviewBootstrap || claims.ResolvedType() == TokenTypePreviewSession {
		normalized, err := NormalizePreviewPath(claims.PreviewPath)
		audience, audienceErr := claims.GetAudience()
		if err != nil ||
			normalized != claims.PreviewPath ||
			claims.PreviewTarget == "" ||
			claims.PreviewOrigin == "" ||
			claims.ID == "" ||
			audienceErr != nil ||
			len(audience) != 1 ||
			audience[0] != claims.PreviewOrigin {
			return nil, ErrInvalidToken
		}
	}

	return claims, nil
}

func (v *TokenValidator) ValidatePreviewToken(tokenString string, tokenType TokenType, previewOrigin string) (*RelayClaims, error) {
	if tokenType != TokenTypePreviewBootstrap && tokenType != TokenTypePreviewSession {
		return nil, ErrInvalidToken
	}
	claims, err := v.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != tokenType || claims.PreviewOrigin != previewOrigin {
		return nil, ErrInvalidToken
	}
	return claims, nil
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
	if tokenType == TokenTypePreviewBootstrap || tokenType == TokenTypePreviewSession || tokenType == TokenType("preview") {
		return "", ErrInvalidToken
	}
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
