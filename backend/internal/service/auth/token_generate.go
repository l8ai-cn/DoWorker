package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
)

func (s *Service) GenerateTokenPair(u *user.User, orgID int64, role string) (*TokenPair, error) {
	return s.GenerateTokenPairWithContext(context.Background(), u, orgID, role)
}

func (s *Service) GenerateTokenPairWithContext(ctx context.Context, u *user.User, orgID int64, role string) (*TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.JWTExpiration)
	refreshExpiresAt := now.Add(s.config.RefreshExpiration)

	if s.config.AccessTokens == nil {
		return nil, ErrAccessTokenConfig
	}
	accessToken, err := s.config.AccessTokens.GenerateToken(
		u.ID,
		u.Email,
		u.Username,
		orgID,
		role,
	)
	if err != nil {
		slog.ErrorContext(ctx, "failed to sign access token", "user_id", u.ID, "error", err)
		return nil, err
	}

	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		slog.ErrorContext(ctx, "failed to generate refresh token bytes", "user_id", u.ID, "error", err)
		return nil, err
	}
	refreshToken := base64.URLEncoding.EncodeToString(refreshBytes)

	if s.redis != nil {
		tokenData := &RefreshTokenData{
			UserID:         u.ID,
			OrganizationID: orgID,
			Role:           role,
			CreatedAt:      now,
			ExpiresAt:      refreshExpiresAt,
		}
		if err := s.storeRefreshToken(ctx, refreshToken, tokenData); err != nil {
			slog.ErrorContext(ctx, "failed to store refresh token in redis", "user_id", u.ID, "error", err)
			return nil, fmt.Errorf("failed to store refresh token: %w", err)
		}
	}

	slog.InfoContext(ctx, "token pair generated", "user_id", u.ID, "org_id", orgID)
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) storeRefreshToken(ctx context.Context, refreshToken string, data *RefreshTokenData) error {
	tokenHash := hashToken(refreshToken)
	key := refreshTokenPrefix + tokenHash

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	ttl := time.Until(data.ExpiresAt)
	return s.redis.Set(ctx, key, jsonData, ttl).Err()
}

func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *Service) GenerateTokens(ctx context.Context, u *user.User) (*LoginResult, error) {
	tokens, err := s.GenerateTokenPairWithContext(ctx, u, 0, "")
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate tokens after email verification", "user_id", u.ID, "error", err)
		return nil, err
	}

	return &LoginResult{
		User:         u,
		Token:        tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    int64(s.config.JWTExpiration.Seconds()),
	}, nil
}
