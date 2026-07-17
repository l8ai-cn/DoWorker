package relay

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (g *TokenGenerator) GeneratePreviewBootstrapToken(
	podKey string,
	runnerID, userID, orgID int64,
	previewTarget, previewPath, previewOrigin string,
	expiry time.Duration,
) (string, error) {
	if podKey == "" {
		return "", fmt.Errorf("preview bootstrap requires pod key")
	}
	if previewTarget == "" {
		return "", fmt.Errorf("preview bootstrap requires target")
	}
	if previewOrigin == "" {
		return "", fmt.Errorf("preview bootstrap requires origin")
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
		TokenType:     "preview_bootstrap",
		PreviewTarget: previewTarget,
		PreviewPath:   normalizedPath,
		PreviewOrigin: previewOrigin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    g.issuer,
			Subject:   podKey,
			ID:        uuid.NewString(),
			Audience:  jwt.ClaimStrings{previewOrigin},
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(g.secretKey)
}
