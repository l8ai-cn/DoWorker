package embedtoken

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

const (
	ContextTokenUse = "agent_embed_context"
	SessionTokenUse = "agent_embed_session"

	contextLifetime = 5 * time.Minute
	sessionLifetime = 15 * time.Minute
	issuer          = "agentcloud-agent-embed"
)

var (
	ErrInvalidToken = errors.New("invalid embed token")
	ErrContextStore = errors.New("embed context store unavailable")
)

type ContextInput struct {
	SessionID            string
	OrganizationID       int64
	OrganizationSlug     string
	UserID               int64
	Email                string
	Capabilities         []string
	AllowedParentOrigins []string
}

type ContextGrant struct {
	Token           string
	RedemptionProof string
	ExpiresAt       time.Time
}

type Claims struct {
	SessionID            string   `json:"session_id"`
	OrganizationID       int64    `json:"org_id"`
	OrganizationSlug     string   `json:"org_slug"`
	UserID               int64    `json:"user_id"`
	Email                string   `json:"email,omitempty"`
	Capabilities         []string `json:"capabilities"`
	AllowedParentOrigins []string `json:"allowed_parent_origins"`
	TokenUse             string   `json:"token_use"`
	jwt.RegisteredClaims
}

type Service struct {
	secret []byte
	redis  *redis.Client
}

func NewService(secret string, redisClient *redis.Client) *Service {
	return &Service{secret: []byte(secret), redis: redisClient}
}

func (s *Service) ValidateContext(token string) (*Claims, error) {
	return s.validate(token, ContextTokenUse)
}

func (s *Service) ValidateSession(token string) (*Claims, error) {
	return s.validate(token, SessionTokenUse)
}
