package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type PreviewSessionIssuer struct {
	secretKey []byte
	issuer    string
}

func NewPreviewSessionIssuer(secret, issuer string) *PreviewSessionIssuer {
	if secret == "" {
		panic("preview session issuer secret must not be empty")
	}
	return &PreviewSessionIssuer{secretKey: []byte(secret), issuer: issuer}
}

type IssuedPreviewSession struct {
	Token     string
	ID        string
	ExpiresAt time.Time
}

func (i *PreviewSessionIssuer) Issue(bootstrap *RelayClaims, expiry time.Duration) (*IssuedPreviewSession, error) {
	if bootstrap == nil || bootstrap.TokenType != TokenTypePreviewBootstrap {
		return nil, ErrInvalidToken
	}
	if expiry <= 0 {
		return nil, fmt.Errorf("expiry must be positive, got %v", expiry)
	}
	now := time.Now()
	sessionID := uuid.NewString()
	expiresAt := now.Add(expiry)
	claims := &RelayClaims{
		PodKey:        bootstrap.PodKey,
		RunnerID:      bootstrap.RunnerID,
		UserID:        bootstrap.UserID,
		OrgID:         bootstrap.OrgID,
		TokenType:     TokenTypePreviewSession,
		PreviewTarget: bootstrap.PreviewTarget,
		PreviewPath:   bootstrap.PreviewPath,
		PreviewOrigin: bootstrap.PreviewOrigin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    i.issuer,
			Subject:   bootstrap.PodKey,
			ID:        sessionID,
			Audience:  jwt.ClaimStrings{bootstrap.PreviewOrigin},
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secretKey)
	if err != nil {
		return nil, err
	}
	return &IssuedPreviewSession{Token: token, ID: sessionID, ExpiresAt: expiresAt}, nil
}
