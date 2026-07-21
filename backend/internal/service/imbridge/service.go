package imbridge

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/imbridge"
)

var (
	ErrNotFound         = errors.New("im connection not found")
	ErrInvalidProvider  = errors.New("invalid im provider")
	ErrInvalidConfig    = errors.New("invalid im connection config")
	ErrConnectionPaused = errors.New("im connection is not active")
)

type ChannelMessenger interface {
	SendMessageAsUser(ctx context.Context, channelID, userID int64, content json.RawMessage) error
}

type Service struct {
	repo     domain.Repository
	registry map[string]Provider
	baseURL  string
}

func NewService(repo domain.Repository, registry map[string]Provider, publicBaseURL string) *Service {
	return &Service{repo: repo, registry: registry, baseURL: strings.TrimRight(publicBaseURL, "/")}
}

func (s *Service) ListProviders() []map[string]string {
	return ListProviderMeta(s.registry)
}

func (s *Service) ListConnections(ctx context.Context, orgID int64) ([]*domain.Connection, error) {
	conns, err := s.repo.ListConnections(ctx, orgID)
	if err != nil {
		return nil, err
	}
	for _, c := range conns {
		s.decorate(c)
	}
	return conns, nil
}

func (s *Service) GetConnection(ctx context.Context, orgID, id int64) (*domain.Connection, error) {
	conn, err := s.repo.GetConnection(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return nil, ErrNotFound
	}
	s.decorate(conn)
	return conn, nil
}

type CreateConnectionRequest struct {
	OrganizationID  int64
	CreatedByUserID int64
	Provider        string
	Name            string
	ChannelID       *int64
	Config          json.RawMessage
	Status          string
}

func (s *Service) CreateConnection(ctx context.Context, req *CreateConnectionRequest) (*domain.Connection, error) {
	if req.Provider == domain.ProviderWeChat {
		req.Provider = domain.ProviderWeixin
	}
	provider, err := GetProvider(s.registry, req.Provider)
	if err != nil {
		return nil, ErrInvalidProvider
	}
	if err := provider.ValidateConfig(req.Config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	token, err := newWebhookToken()
	if err != nil {
		return nil, err
	}
	status := req.Status
	if status == "" {
		status = domain.StatusDisabled
	}
	conn := &domain.Connection{
		OrganizationID:  req.OrganizationID,
		Provider:        req.Provider,
		Name:            strings.TrimSpace(req.Name),
		ChannelID:       req.ChannelID,
		Config:          req.Config,
		WebhookToken:    token,
		Status:          status,
		CreatedByUserID: req.CreatedByUserID,
	}
	if conn.Name == "" {
		return nil, fmt.Errorf("%w: name required", ErrInvalidConfig)
	}
	if err := s.repo.CreateConnection(ctx, conn); err != nil {
		return nil, err
	}
	s.decorate(conn)
	return conn, nil
}

type UpdateConnectionRequest struct {
	Name      *string
	ChannelID *int64
	Config    json.RawMessage
	Status    *string
}

func (s *Service) UpdateConnection(ctx context.Context, orgID, id int64, req *UpdateConnectionRequest) (*domain.Connection, error) {
	conn, err := s.GetConnection(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		conn.Name = strings.TrimSpace(*req.Name)
	}
	if req.ChannelID != nil {
		conn.ChannelID = req.ChannelID
	}
	if len(req.Config) > 0 {
		provider, err := GetProvider(s.registry, conn.Provider)
		if err != nil {
			return nil, err
		}
		if err := provider.ValidateConfig(req.Config); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
		}
		conn.Config = req.Config
	}
	if req.Status != nil {
		conn.Status = *req.Status
	}
	if err := s.repo.UpdateConnection(ctx, conn); err != nil {
		return nil, err
	}
	s.decorate(conn)
	return conn, nil
}

func (s *Service) DeleteConnection(ctx context.Context, orgID, id int64) error {
	return s.repo.DeleteConnection(ctx, orgID, id)
}

func (s *Service) decorate(conn *domain.Connection) {
	conn.WebhookURL = fmt.Sprintf("%s/api/v1/webhooks/im/%s/%d?token=%s",
		s.baseURL, conn.Provider, conn.ID, conn.WebhookToken)
}

func newWebhookToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (s *Service) connectionForWebhook(ctx context.Context, provider, token string, connectionID int64) (*domain.Connection, error) {
	conn, err := s.repo.GetConnectionByToken(ctx, provider, token)
	if err != nil {
		return nil, err
	}
	if conn == nil || conn.ID != connectionID {
		return nil, ErrNotFound
	}
	return conn, nil
}

func (s *Service) markError(ctx context.Context, conn *domain.Connection, msg string) {
	conn.Status = domain.StatusError
	conn.LastError = &msg
	_ = s.repo.UpdateConnection(ctx, conn)
}

func (s *Service) HandleWebhook(ctx context.Context, provider string, connectionID int64, token string, headers http.Header, body []byte) (interface{}, error) {
	conn, err := s.connectionForWebhook(ctx, provider, token, connectionID)
	if err != nil {
		return nil, err
	}
	if conn.Status != domain.StatusActive {
		return nil, ErrConnectionPaused
	}
	p, err := GetProvider(s.registry, provider)
	if err != nil {
		return nil, err
	}
	if err := p.VerifyWebhook(ctx, conn.Config, headers, body); err != nil {
		s.markError(ctx, conn, err.Error())
		return nil, err
	}
	event, err := p.ParseInbound(ctx, conn.Config, headers, body)
	if err != nil {
		s.markError(ctx, conn, err.Error())
		return nil, err
	}
	if event.Challenge != "" {
		return map[string]string{"challenge": event.Challenge}, nil
	}
	if strings.TrimSpace(event.Text) == "" {
		return map[string]string{"status": "ignored"}, nil
	}
	return map[string]string{"status": "accepted", "thread": event.ExternalThreadID}, nil
}
