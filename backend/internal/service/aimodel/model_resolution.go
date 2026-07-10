package aimodel

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
)

var ErrNotFound = errors.New("ai model not found")

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

func (s *Service) GetVisible(ctx context.Context, id, userID, orgID int64) (*aimodel.AIModel, error) {
	m, err := s.repo.GetVisibleByID(ctx, id, userID, orgID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNotFound
	}
	return m, nil
}

func (s *Service) ResolveVisible(ctx context.Context, id, userID, orgID int64) (*ResolvedModel, error) {
	m, err := s.GetVisible(ctx, id, userID, orgID)
	if err != nil {
		return nil, err
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
