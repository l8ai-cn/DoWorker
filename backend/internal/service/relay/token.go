package relay

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	PodKey        string `json:"pod_key"`
	RunnerID      int64  `json:"runner_id"`
	UserID        int64  `json:"user_id"` // 0 for runner tokens
	OrgID         int64  `json:"org_id"`
	TokenType     string `json:"token_type,omitempty"`
	PreviewTarget string `json:"preview_target,omitempty"` // e.g. 127.0.0.1:3000
	PreviewPath   string `json:"preview_path,omitempty"`
	jwt.RegisteredClaims
}

type TokenGenerator struct {
	secretKey []byte
	issuer    string
}

// NewTokenGenerator creates a new token generator.
// Panics if secret is empty to prevent signing tokens with a zero-length HMAC key.
func NewTokenGenerator(secret, issuer string) *TokenGenerator {
	if secret == "" {
		panic("relay token secret must not be empty")
	}
	return &TokenGenerator{
		secretKey: []byte(secret),
		issuer:    issuer,
	}
}

func (g *TokenGenerator) GenerateToken(podKey string, runnerID, userID, orgID int64, expiry time.Duration) (string, error) {
	if podKey == "" {
		return "", fmt.Errorf("podKey must not be empty")
	}
	if expiry <= 0 {
		return "", fmt.Errorf("expiry must be positive, got %v", expiry)
	}
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := &TokenClaims{
		PodKey:   podKey,
		RunnerID: runnerID,
		UserID:   userID,
		OrgID:    orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    g.issuer,
			Subject:   podKey,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secretKey)
}

// GenerateTypedToken generates non-preview tokens with an explicit token_type.
// PodKey may be empty for tunnel tokens, which are not bound to a single pod.
func (g *TokenGenerator) GenerateTypedToken(podKey string, runnerID, userID, orgID int64, tokenType, previewTarget string, expiry time.Duration) (string, error) {
	if tokenType == "preview" {
		return "", fmt.Errorf("preview tokens require GeneratePreviewToken")
	}
	if expiry <= 0 {
		return "", fmt.Errorf("expiry must be positive, got %v", expiry)
	}
	now := time.Now()
	claims := &TokenClaims{
		PodKey:        podKey,
		RunnerID:      runnerID,
		UserID:        userID,
		OrgID:         orgID,
		TokenType:     tokenType,
		PreviewTarget: previewTarget,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    g.issuer,
			Subject:   podKey,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(g.secretKey)
}

func (g *TokenGenerator) GeneratePreviewToken(podKey string, runnerID, userID, orgID int64, previewTarget, previewPath string, expiry time.Duration) (string, error) {
	if podKey == "" {
		return "", fmt.Errorf("preview token requires pod key")
	}
	if previewTarget == "" {
		return "", fmt.Errorf("preview token requires target")
	}
	normalizedPath, err := NormalizePreviewPath(previewPath)
	if err != nil {
		return "", err
	}
	if expiry <= 0 {
		return "", fmt.Errorf("expiry must be positive, got %v", expiry)
	}
	now := time.Now()
	claims := &TokenClaims{
		PodKey:        podKey,
		RunnerID:      runnerID,
		UserID:        userID,
		OrgID:         orgID,
		TokenType:     "preview",
		PreviewTarget: previewTarget,
		PreviewPath:   normalizedPath,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    g.issuer,
			Subject:   podKey,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(g.secretKey)
}
