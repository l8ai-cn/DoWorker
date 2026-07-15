package expert

import (
	"context"
	"strings"

	"github.com/lib/pq"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
)

func (s *Service) Update(
	ctx context.Context,
	req *UpdateExpertRequest,
) (*expertdom.Expert, error) {
	row, err := s.store.GetByID(ctx, req.OrganizationID, req.ExpertID)
	if err != nil {
		return nil, err
	}
	if row.SourceMarketApplicationID == nil || s.marketInstallLock == nil {
		return s.updateExpert(ctx, req, row)
	}
	var updated *expertdom.Expert
	err = s.marketInstallLock.WithinMarketInstallationLock(
		ctx,
		req.OrganizationID,
		*row.SourceMarketApplicationID,
		func() error {
			current, loadErr := s.store.GetByID(
				ctx, req.OrganizationID, req.ExpertID,
			)
			if loadErr != nil {
				return loadErr
			}
			updated, loadErr = s.updateExpert(ctx, req, current)
			return loadErr
		},
	)
	return updated, err
}

func (s *Service) updateExpert(
	ctx context.Context,
	req *UpdateExpertRequest,
	row *expertdom.Expert,
) (*expertdom.Expert, error) {
	before := *row
	if err := applyExpertUpdate(row, req); err != nil {
		return nil, err
	}
	createdSnapshotID, err := s.refreshExpertWorkerSpec(ctx, &before, row)
	if err != nil {
		return nil, err
	}
	if err := s.persistExpertUpdate(ctx, row, req.Avatar); err != nil {
		s.cleanupExpertSnapshot(ctx, row.OrganizationID, createdSnapshotID)
		return nil, err
	}
	return row, nil
}

func applyExpertUpdate(
	row *expertdom.Expert,
	req *UpdateExpertRequest,
) error {
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return ErrExpertNameRequired
		}
		row.Name = name
	}
	if req.Description != nil {
		row.Description = trimOptional(req.Description)
	}
	if req.AgentSlug != nil {
		slug := strings.TrimSpace(*req.AgentSlug)
		if slug == "" {
			return ErrExpertAgentRequired
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
	if req.AutomationLevel != nil {
		row.AutomationLevel = expertdom.NormalizeAutomationLevel(*req.AutomationLevel)
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
		path := req.Avatar.repoPath()
		avatarPath = &path
	}
	if avatarPath != nil || req.ExpertType != nil {
		row.Metadata = mergeMetadata(row.Metadata, avatarPath, req.ExpertType)
	}
	return nil
}

func (s *Service) persistExpertUpdate(
	ctx context.Context,
	row *expertdom.Expert,
	avatar *AvatarInput,
) error {
	if s.gitops != nil {
		layer := s.buildAgentfileLayer(ctx, row)
		provisioned, err := s.ensureExpertRepo(ctx, row, layer, avatar)
		if err != nil {
			return err
		}
		if !provisioned {
			if err := s.commitExpertChanges(ctx, row, layer, avatar); err != nil {
				return err
			}
		}
	}
	return s.store.Update(ctx, row)
}
