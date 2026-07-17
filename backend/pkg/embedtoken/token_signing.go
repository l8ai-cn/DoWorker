package embedtoken

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (s *Service) issue(
	input ContextInput,
	tokenUse string,
	lifetime time.Duration,
) (string, time.Time, string, error) {
	if s == nil || len(s.secret) == 0 || input.SessionID == "" || input.OrganizationID <= 0 ||
		input.OrganizationSlug == "" || input.UserID <= 0 || len(input.Capabilities) == 0 ||
		len(input.AllowedParentOrigins) == 0 {
		return "", time.Time{}, "", ErrInvalidToken
	}
	jti, err := tokenID()
	if err != nil {
		return "", time.Time{}, "", err
	}
	now := time.Now()
	expiresAt := now.Add(lifetime)
	claims := Claims{
		SessionID:            input.SessionID,
		OrganizationID:       input.OrganizationID,
		OrganizationSlug:     input.OrganizationSlug,
		UserID:               input.UserID,
		Email:                input.Email,
		Capabilities:         append([]string(nil), input.Capabilities...),
		AllowedParentOrigins: append([]string(nil), input.AllowedParentOrigins...),
		TokenUse:             tokenUse,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   input.SessionID,
			ID:        jti,
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, "", fmt.Errorf("sign embed token: %w", err)
	}
	return signed, expiresAt, jti, nil
}

func (s *Service) validate(token string, expectedUse string) (*Claims, error) {
	if s == nil || len(s.secret) == 0 || token == "" {
		return nil, ErrInvalidToken
	}
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(parsed *jwt.Token) (any, error) {
		if _, ok := parsed.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid || claims.Issuer != issuer || claims.TokenUse != expectedUse ||
		claims.Subject != claims.SessionID || claims.SessionID == "" || claims.OrganizationID <= 0 ||
		claims.OrganizationSlug == "" || claims.UserID <= 0 || len(claims.Capabilities) == 0 ||
		len(claims.AllowedParentOrigins) == 0 || claims.ID == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func tokenID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate embed token id: %w", err)
	}
	return hex.EncodeToString(value), nil
}
