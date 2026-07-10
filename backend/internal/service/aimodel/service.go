package aimodel

import (
	"context"
	"encoding/json"

	"github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
)

// Service manages the model pool: encrypted-credential CRUD plus resolving a
// model into the do-agent settings.json document for pod injection.
type Service struct {
	repo      aimodel.Repository
	encryptor *crypto.Encryptor
}

func NewService(repo aimodel.Repository, encryptor *crypto.Encryptor) *Service {
	return &Service{repo: repo, encryptor: encryptor}
}

// CreateInput is the scope + fields for a new pool row. Exactly one of OrgID /
// UserID identifies the owning scope; credentials are plaintext KV (encrypted
// here).
type CreateInput struct {
	OrgID        *int64
	UserID       *int64
	Name         string
	ProviderType string
	Model        string
	BaseURL      string
	Credentials  map[string]string
	IsDefault    bool
	TokenBudget  *int64
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*aimodel.AIModel, error) {
	enc, err := s.encrypt(in.Credentials)
	if err != nil {
		return nil, err
	}
	m := &aimodel.AIModel{
		OrganizationID:       in.OrgID,
		UserID:               in.UserID,
		Name:                 in.Name,
		ProviderType:         in.ProviderType,
		Model:                in.Model,
		BaseURL:              in.BaseURL,
		EncryptedCredentials: enc,
		IsDefault:            in.IsDefault,
		IsEnabled:            true,
		TokenBudget:          in.TokenBudget,
	}
	if in.IsDefault {
		if err := s.clearDefaults(ctx, m); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) ListVisible(ctx context.Context, userID, orgID int64) ([]*aimodel.AIModel, error) {
	return s.repo.ListVisible(ctx, userID, orgID)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) clearDefaults(ctx context.Context, m *aimodel.AIModel) error {
	userID := int64(0)
	orgID := int64(0)
	if m.UserID != nil {
		userID = *m.UserID
	}
	if m.OrganizationID != nil {
		orgID = *m.OrganizationID
	}
	return s.repo.ClearDefaults(ctx, userID, orgID)
}

func (s *Service) encrypt(creds map[string]string) (string, error) {
	if len(creds) == 0 {
		return "", nil
	}
	b, err := json.Marshal(creds)
	if err != nil {
		return "", err
	}
	if s.encryptor != nil {
		return s.encryptor.Encrypt(string(b))
	}
	return string(b), nil
}

func (s *Service) decrypt(enc string) (map[string]string, error) {
	if enc == "" {
		return map[string]string{}, nil
	}
	raw := enc
	if s.encryptor != nil {
		dec, err := s.encryptor.Decrypt(enc)
		if err != nil {
			return nil, err
		}
		raw = dec
	}
	var creds map[string]string
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return nil, err
	}
	return creds, nil
}
