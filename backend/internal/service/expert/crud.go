package expert

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/lib/pq"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
)

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
		OrganizationID:            req.OrganizationID,
		Slug:                      slug,
		Name:                      strings.TrimSpace(req.Name),
		Description:               trimOptional(req.Description),
		AgentSlug:                 strings.TrimSpace(req.AgentSlug),
		RunnerID:                  req.RunnerID,
		RepositoryID:              req.RepositoryID,
		BranchName:                trimOptional(req.BranchName),
		Prompt:                    trimOptional(req.Prompt),
		InteractionMode:           mode,
		AutomationLevel:           expertdom.NormalizeAutomationLevel(req.AutomationLevel),
		Perpetual:                 req.Perpetual,
		UsedEnvBundles:            pq.StringArray(nonEmptyStrings(req.UsedEnvBundles)),
		SkillSlugs:                pq.StringArray(nonEmptyStrings(req.SkillSlugs)),
		KnowledgeMounts:           encodeKnowledgeMounts(req.KnowledgeMounts),
		ConfigOverrides:           encodeConfigOverrides(req.ConfigOverrides),
		AgentfileLayer:            trimOptional(req.AgentfileLayer),
		SourcePodKey:              trimOptional(req.SourcePodKey),
		WorkerSpecSnapshotID:      req.WorkerSpecSnapshotID,
		SourceMarketApplicationID: req.SourceMarketApplicationID,
		SourceMarketReleaseID:     req.SourceMarketReleaseID,
		DefaultBranch:             "main",
		CreatedByID:               req.UserID,
	}
	if len(req.Metadata) > 0 {
		row.Metadata = append(json.RawMessage(nil), req.Metadata...)
	}

	var avatarPath *string
	if req.Avatar != nil && len(req.Avatar.Data) > 0 {
		p := req.Avatar.repoPath()
		avatarPath = &p
	}
	row.Metadata = mergeMetadata(row.Metadata, avatarPath, req.ExpertType)

	// Provision first so a database failure can compensate by deleting the
	// newly created repository.
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
			cleanupCtx, cancel := marketCleanupContext(ctx)
			defer cancel()
			repoName := s.gitops.RepoNameFromPath(*row.GitRepoPath)
			if delErr := s.gitops.DeleteRepo(cleanupCtx, repoName); delErr != nil {
				s.logger.Warn("expert: compensating repo delete failed",
					"repo", repoName, "error", delErr)
			}
		}
		return nil, err
	}
	return row, nil
}
