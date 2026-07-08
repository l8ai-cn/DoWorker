package aimodel

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	"github.com/anthropics/agentsmesh/backend/pkg/crypto"
)

var (
	ErrNotFound = errors.New("ai model not found")
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

// ResolvedModel is a model row with its provider credentials decrypted, ready
// for do-agent settings injection.
type ResolvedModel struct {
	Model       *aimodel.AIModel
	Credentials map[string]string
}

// Resolve loads a pool row by id and decrypts its credentials.
func (s *Service) Resolve(ctx context.Context, id int64) (*ResolvedModel, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNotFound
	}
	return s.resolveRow(m)
}

// ResolveDefault resolves the default visible model for a caller, or nil when
// none is configured.
func (s *Service) ResolveDefault(ctx context.Context, userID, orgID int64) (*ResolvedModel, error) {
	m, err := s.repo.DefaultVisible(ctx, userID, orgID)
	if err != nil || m == nil {
		return nil, err
	}
	return s.resolveRow(m)
}

// ResolveDefaultForAgent picks a pool model when the client omits
// model_config_id: harness-specific provider first, else org/user default.
func (s *Service) ResolveDefaultForAgent(ctx context.Context, userID, orgID int64, agentSlug string) (*ResolvedModel, error) {
	for _, p := range aimodel.PreferredProviders(agentSlug) {
		m, err := s.repo.FirstVisibleByProvider(ctx, userID, orgID, p)
		if err != nil {
			return nil, err
		}
		if m != nil {
			return s.resolveRow(m)
		}
	}
	return s.ResolveDefault(ctx, userID, orgID)
}

func (s *Service) resolveRow(m *aimodel.AIModel) (*ResolvedModel, error) {
	creds, err := s.decrypt(m.EncryptedCredentials)
	if err != nil {
		return nil, err
	}
	return &ResolvedModel{Model: m, Credentials: creds}, nil
}

// SettingsJSON returns the do-agent settings.json document for a resolved
// model, optionally overriding the model id (empty => row default).
func (r *ResolvedModel) SettingsJSON(overrideModel string) map[string]interface{} {
	model := overrideModel
	if model == "" {
		model = r.Model.Model
	}
	return aimodel.DoAgentSettings(r.Model.ProviderType, model, r.Model.BaseURL, r.Credentials)
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
