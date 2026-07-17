package embedtoken

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var redeemContextScript = redis.NewScript(`
local stored = redis.call("GET", KEYS[1])
if not stored then
  return 0
end
if stored ~= ARGV[1] then
  return -1
end
redis.call("DEL", KEYS[1])
return 1
`)

func (s *Service) IssueContext(
	ctx context.Context,
	input ContextInput,
) (ContextGrant, error) {
	token, expiresAt, id, err := s.issue(input, ContextTokenUse, contextLifetime)
	if err != nil {
		return ContextGrant{}, err
	}
	if s.redis == nil {
		return ContextGrant{}, ErrContextStore
	}
	proof, err := redemptionProof()
	if err != nil {
		return ContextGrant{}, err
	}
	stored, err := s.redis.SetNX(
		ctx,
		contextStoreKey(id),
		hashProof(proof),
		time.Until(expiresAt),
	).Result()
	if err != nil {
		return ContextGrant{}, fmt.Errorf("%w: %v", ErrContextStore, err)
	}
	if !stored {
		return ContextGrant{}, ErrInvalidToken
	}
	return ContextGrant{
		Token:           token,
		RedemptionProof: proof,
		ExpiresAt:       expiresAt,
	}, nil
}

func (s *Service) InspectContext(ctx context.Context, token string) (*Claims, error) {
	claims, err := s.ValidateContext(token)
	if err != nil {
		return nil, err
	}
	if s.redis == nil {
		return nil, ErrContextStore
	}
	if _, err := s.redis.Get(ctx, contextStoreKey(claims.ID)).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("%w: %v", ErrContextStore, err)
	}
	return claims, nil
}

func (s *Service) RedeemContext(
	ctx context.Context,
	token string,
	proof string,
) (string, time.Time, error) {
	claims, err := s.ValidateContext(token)
	if err != nil || proof == "" {
		return "", time.Time{}, ErrInvalidToken
	}
	if s.redis == nil {
		return "", time.Time{}, ErrContextStore
	}
	result, err := redeemContextScript.Run(
		ctx,
		s.redis,
		[]string{contextStoreKey(claims.ID)},
		hashProof(proof),
	).Int64()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("%w: %v", ErrContextStore, err)
	}
	if result != 1 {
		return "", time.Time{}, ErrInvalidToken
	}
	sessionToken, expiresAt, _, err := s.issue(contextInputFromClaims(claims), SessionTokenUse, sessionLifetime)
	return sessionToken, expiresAt, err
}

func contextInputFromClaims(claims *Claims) ContextInput {
	return ContextInput{
		SessionID:            claims.SessionID,
		OrganizationID:       claims.OrganizationID,
		OrganizationSlug:     claims.OrganizationSlug,
		UserID:               claims.UserID,
		Email:                claims.Email,
		Capabilities:         claims.Capabilities,
		AllowedParentOrigins: claims.AllowedParentOrigins,
	}
}

func contextStoreKey(id string) string {
	return "agent-embed-context:" + id
}

func redemptionProof() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate embed redemption proof: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func hashProof(proof string) string {
	sum := sha256.Sum256([]byte(proof))
	return hex.EncodeToString(sum[:])
}
