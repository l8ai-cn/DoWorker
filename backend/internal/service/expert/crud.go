package expert

import (
	"context"
	"strings"

	"github.com/lib/pq"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type CreateExpertRequest struct {
	OrganizationID int64
	UserID         int64
	Name           string
	Slug           string
	Description    *string
	AgentSlug      string
	RunnerID       *int64
	RepositoryID   *int64
	BranchName     *string
	Prompt         *string
	InteractionMode string
	Perpetual      bool
	UsedEnvBundles []string
	SkillSlugs     []string
	KnowledgeMounts []expertdom.KnowledgeMount
	ConfigOverrides map[string]interface{}
	AgentfileLayer *string
	SourcePodKey   *string
	Avatar         *AvatarInput
	ExpertType     *string
}

type UpdateExpertRequest struct {
	OrganizationID int64
	ExpertID       int64
	Name           *string
	Description    *string
	AgentSlug      *string
	RunnerID       *int64
	RepositoryID   *int64
	BranchName     *string
	Prompt         *string
	InteractionMode *string
	Perpetual      *bool
	UsedEnvBundles []string
	SkillSlugs     []string
	KnowledgeMounts []expertdom.KnowledgeMount
	ConfigOverrides map[string]interface{}
	AgentfileLayer *string
	Avatar         *AvatarInput
	ExpertType     *string
}

func (s *Service) Create(ctx context.Context, req *CreateExpertRequest) (*expertdom.Expert, error) {
	if err := validateExpertBasics(req.AgentSlug, req.Name); err != nil {
		return nil, err
	}
	slug, err := s.resolveSlug(ctx, req.OrganizationID, req.Slug, req.Name, 0)
	if err != nil {
		return nil, err
	}
	mode := normalizeInteractionMode(req.InteractionMode)
	row := &expertdom.Expert{
		OrganizationID:  req.OrganizationID,
		Slug:            slug,
		Name:            strings.TrimSpace(req.Name),
		Description:     trimOptional(req.Description),
		AgentSlug:       strings.TrimSpace(req.AgentSlug),
		RunnerID:        req.RunnerID,
		RepositoryID:    req.RepositoryID,
		BranchName:      trimOptional(req.BranchName),
		Prompt:          trimOptional(req.Prompt),
		InteractionMode: mode,
		Perpetual:       req.Perpetual,
		UsedEnvBundles:  pq.StringArray(nonEmptyStrings(req.UsedEnvBundles)),
		SkillSlugs:      pq.StringArray(nonEmptyStrings(req.SkillSlugs)),
		KnowledgeMounts: encodeKnowledgeMounts(req.KnowledgeMounts),
		ConfigOverrides: encodeConfigOverrides(req.ConfigOverrides),
		AgentfileLayer:  trimOptional(req.AgentfileLayer),
		SourcePodKey:    trimOptional(req.SourcePodKey),
		DefaultBranch:   "main",
		CreatedByID:     req.UserID,
	}

	var avatarPath *string
	if req.Avatar != nil && len(req.Avatar.Data) > 0 {
		p := req.Avatar.repoPath()
		avatarPath = &p
	}
	row.Metadata = mergeMetadata(row.Metadata, avatarPath, req.ExpertType)

	// Git is the source of truth: provision + seed the repo first, then persist
	// the DB cache row. On DB failure, compensate by deleting the fresh repo.
	provisioned := false
	if s.gitops != nil {
		layer := s.buildAgentfileLayer(ctx, row)
		repo, err := s.provisionExpertRepo(ctx, row, layer, req.Avatar)
		if err != nil {
			return nil, err
		}
		applyRepo(row, repo)
		provisioned = true
	}

	if err := s.store.Create(ctx, row); err != nil {
		if provisioned && row.GitRepoPath != nil {
			repoName := s.gitops.RepoNameFromPath(*row.GitRepoPath)
			if delErr := s.gitops.DeleteRepo(ctx, repoName); delErr != nil {
				s.logger.Warn("expert: compensating repo delete failed",
					"repo", repoName, "error", delErr)
			}
		}
		return nil, err
	}
	return row, nil
}

func (s *Service) Update(ctx context.Context, req *UpdateExpertRequest) (*expertdom.Expert, error) {
	row, err := s.store.GetByID(ctx, req.OrganizationID, req.ExpertID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrExpertNameRequired
		}
		row.Name = name
	}
	if req.Description != nil {
		row.Description = trimOptional(req.Description)
	}
	if req.AgentSlug != nil {
		slug := strings.TrimSpace(*req.AgentSlug)
		if slug == "" {
			return nil, ErrExpertAgentRequired
		}
		row.AgentSlug = slug
	}
	if req.RunnerID != nil {
		row.RunnerID = req.RunnerID
	}
	if req.RepositoryID != nil {
		row.RepositoryID = req.RepositoryID
	}
	if req.BranchName != nil {
		row.BranchName = trimOptional(req.BranchName)
	}
	if req.Prompt != nil {
		row.Prompt = trimOptional(req.Prompt)
	}
	if req.InteractionMode != nil {
		row.InteractionMode = normalizeInteractionMode(*req.InteractionMode)
	}
	if req.Perpetual != nil {
		row.Perpetual = *req.Perpetual
	}
	if req.UsedEnvBundles != nil {
		row.UsedEnvBundles = pq.StringArray(nonEmptyStrings(req.UsedEnvBundles))
	}
	if req.SkillSlugs != nil {
		row.SkillSlugs = pq.StringArray(nonEmptyStrings(req.SkillSlugs))
	}
	if req.KnowledgeMounts != nil {
		row.KnowledgeMounts = encodeKnowledgeMounts(req.KnowledgeMounts)
	}
	if req.ConfigOverrides != nil {
		row.ConfigOverrides = encodeConfigOverrides(req.ConfigOverrides)
	}
	if req.AgentfileLayer != nil {
		row.AgentfileLayer = trimOptional(req.AgentfileLayer)
	}

	var avatarPath *string
	if req.Avatar != nil && len(req.Avatar.Data) > 0 {
		p := req.Avatar.repoPath()
		avatarPath = &p
	}
	if avatarPath != nil || req.ExpertType != nil {
		row.Metadata = mergeMetadata(row.Metadata, avatarPath, req.ExpertType)
	}

	// Git is the source of truth: commit changed files first (lazily
	// provisioning a repo for legacy rows), then refresh the DB cache. On a
	// commit-succeeds / store.Update-fails partial failure, Git is ahead and
	// the cache is stale; it is reconciled on the next Git-first read (Run).
	if s.gitops != nil {
		layer := s.buildAgentfileLayer(ctx, row)
		provisioned, err := s.ensureExpertRepo(ctx, row, layer, req.Avatar)
		if err != nil {
			return nil, err
		}
		if !provisioned {
			if err := s.commitExpertChanges(ctx, row, layer, req.Avatar); err != nil {
				return nil, err
			}
		}
	}

	if err := s.store.Update(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) Delete(ctx context.Context, orgID, id int64) error {
	row, err := s.store.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	// DB row is deleted first (authoritative for existence), then the repo is
	// removed best-effort (mirrors the knowledgebase Delete ordering).
	if err := s.store.Delete(ctx, orgID, id); err != nil {
		return err
	}
	if s.gitops != nil && row.GitRepoPath != nil {
		repoName := s.gitops.RepoNameFromPath(*row.GitRepoPath)
		if delErr := s.gitops.DeleteRepo(ctx, repoName); delErr != nil {
			s.logger.Warn("expert: repo delete failed", "repo", repoName, "error", delErr)
		}
	}
	return nil
}

func (s *Service) GetBySlug(ctx context.Context, orgID int64, slug string) (*expertdom.Expert, error) {
	return s.store.GetBySlug(ctx, orgID, slug)
}

func (s *Service) GetByID(ctx context.Context, orgID, id int64) (*expertdom.Expert, error) {
	return s.store.GetByID(ctx, orgID, id)
}

func (s *Service) List(ctx context.Context, orgID int64, limit, offset int) ([]expertdom.Expert, int64, error) {
	return s.store.List(ctx, orgID, limit, offset)
}

func (s *Service) resolveSlug(ctx context.Context, orgID int64, explicit, nameSeed string, excludeID int64) (string, error) {
	seed := strings.TrimSpace(explicit)
	if seed == "" {
		seed = nameSeed
	}
	return slugkit.GenerateUnique(ctx, seed, slugkit.FromExistsCheck(func(ctx context.Context, candidate string) (bool, error) {
		return s.store.SlugExists(ctx, orgID, candidate, excludeID)
	}))
}
