package knowledgebase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/knowledgebase"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/gitea"
	"github.com/l8ai-cn/agentcloud/backend/pkg/crypto"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

var (
	ErrNotFound      = knowledgebase.ErrNotFound
	ErrInvalidInput  = errors.New("knowledgebase: invalid input")
	ErrNotConfigured = gitea.ErrNotConfigured
)

type Service struct {
	repo    knowledgebase.Repository
	git     *gitea.Client
	log     *slog.Logger
	secrets *crypto.Encryptor
}

func NewService(repo knowledgebase.Repository, git *gitea.Client, log *slog.Logger) *Service {
	if git == nil {
		return nil
	}
	if log == nil {
		log = slog.Default()
	}
	return &Service{repo: repo, git: git, log: log}
}

type CreateParams struct {
	OrganizationID  int64
	CreatedByUserID int64
	Name            string
	Description     string
	SourceType      string
	SourceConfig    json.RawMessage
}

func (s *Service) Create(ctx context.Context, p *CreateParams) (*knowledgebase.KnowledgeBase, error) {
	if strings.TrimSpace(p.Name) == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	sourceType := p.SourceType
	if sourceType == "" {
		sourceType = knowledgebase.SourceTypeGit
	}
	if !knowledgebase.ValidSourceType(sourceType) {
		return nil, fmt.Errorf("%w: unknown source_type %q", ErrInvalidInput, sourceType)
	}

	slug, err := slugkit.GenerateUnique(ctx, p.Name, slugkit.FromExistsCheck(
		func(ctx context.Context, candidate string) (bool, error) {
			return s.repo.SlugExists(ctx, p.OrganizationID, candidate)
		}))
	if err != nil {
		return nil, fmt.Errorf("%w: cannot derive slug from name: %v", ErrInvalidInput, err)
	}

	const branch = "main"
	repo, repoName, err := s.provisionRepo(ctx, p.OrganizationID, slug, p.Name, p.Description, branch)
	if err != nil {
		return nil, err
	}
	if s.secrets == nil {
		return nil, s.failCreateAndCleanupRepo(
			ctx,
			repoName,
			fmt.Errorf("%w: deploy key encryption is not configured", ErrNotConfigured),
		)
	}
	mountKeys, err := s.provisionMountDeployKeys(ctx, repoName)
	if err != nil {
		return nil, s.failCreateAndCleanupRepo(ctx, repoName, err)
	}

	sourceConfig := p.SourceConfig
	if len(sourceConfig) == 0 {
		sourceConfig = json.RawMessage("{}")
	}
	sourceConfig, err = addMountDeployKeys(sourceConfig, mountKeys)
	if err != nil {
		return nil, s.failCreateAndCleanupRepo(ctx, repoName, err)
	}
	sourceConfig, err = s.encryptSourceSecrets(sourceConfig)
	if err != nil {
		return nil, s.failCreateAndCleanupRepo(ctx, repoName, err)
	}
	kb := &knowledgebase.KnowledgeBase{
		OrganizationID:  p.OrganizationID,
		Slug:            slug,
		Name:            p.Name,
		Description:     p.Description,
		GitRepoPath:     s.git.Namespace() + "/" + repoName,
		HTTPCloneURL:    s.git.CloneURL(repoName),
		DefaultBranch:   repo.DefaultBranch,
		SourceType:      sourceType,
		SourceConfig:    sourceConfig,
		SyncStatus:      knowledgebase.SyncStatusIdle,
		CreatedByUserID: p.CreatedByUserID,
	}
	if kb.DefaultBranch == "" {
		kb.DefaultBranch = branch
	}
	if err := s.repo.Create(ctx, kb); err != nil {
		return nil, s.failCreateAndCleanupRepo(ctx, repoName, err)
	}
	s.log.Info("knowledge base created", "org_id", p.OrganizationID, "slug", slug, "repo", kb.GitRepoPath)
	return kb, nil
}

func (s *Service) Get(ctx context.Context, orgID, id int64) (*knowledgebase.KnowledgeBase, error) {
	return s.repo.Get(ctx, orgID, id)
}

func (s *Service) GetBySlug(ctx context.Context, orgID int64, slug string) (*knowledgebase.KnowledgeBase, error) {
	return s.repo.GetBySlug(ctx, orgID, slug)
}

func (s *Service) List(ctx context.Context, orgID int64, sourceType string) ([]*knowledgebase.KnowledgeBase, error) {
	return s.repo.List(ctx, &knowledgebase.ListFilter{OrganizationID: orgID, SourceType: sourceType})
}

type UpdateParams struct {
	Name         *string
	Description  *string
	SourceConfig json.RawMessage
}

func (s *Service) Update(ctx context.Context, orgID, id int64, p *UpdateParams) (*knowledgebase.KnowledgeBase, error) {
	kb, err := s.repo.Get(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if p.Name != nil {
		if strings.TrimSpace(*p.Name) == "" {
			return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
		}
		updates["name"] = *p.Name
	}
	if p.Description != nil {
		updates["description"] = *p.Description
	}
	if len(p.SourceConfig) > 0 {
		if kb.SourceType == knowledgebase.SourceTypeGit {
			return nil, fmt.Errorf("%w: git knowledge bases have no external source_config", ErrInvalidInput)
		}
		merged, err := s.mergeSourceConfigUpdate(kb.SourceConfig, p.SourceConfig)
		if err != nil {
			return nil, err
		}
		encrypted, err := s.encryptSourceSecrets(merged)
		if err != nil {
			return nil, err
		}
		updates["source_config"] = encrypted
	}
	if len(updates) > 0 {
		if err := s.repo.Update(ctx, orgID, id, updates); err != nil {
			return nil, err
		}
	}
	return s.repo.Get(ctx, orgID, id)
}
