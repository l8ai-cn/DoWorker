package expert

import (
	"context"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/gitops"
)

func (s *Service) publishedMarketApplicationByID(
	ctx context.Context,
	applicationID int64,
) (*PublishedMarketApplication, error) {
	if s.market == nil {
		return nil, ErrMarketUnavailable
	}
	application, err := s.market.GetApplicationByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if application.LatestPublishedReleaseID == nil {
		return nil, ErrMarketApplicationNotFound
	}
	release, err := s.market.GetReleaseByID(
		ctx,
		*application.LatestPublishedReleaseID,
	)
	if err != nil {
		return nil, err
	}
	if release.Status != expertmarket.ReleaseStatusPublished {
		return nil, ErrMarketReleaseNotPublished
	}
	return &PublishedMarketApplication{
		Application: *application,
		Release:     *release,
	}, nil
}

func (s *Service) commitMarketUpgrade(
	ctx context.Context,
	installed *expertdom.Expert,
	update expertdom.MarketReleaseUpdate,
) (func(context.Context) error, error) {
	if s.gitops == nil || installed.GitRepoPath == nil {
		return nil, nil
	}
	previousLayer := s.buildAgentfileLayer(ctx, installed)
	previousChanges, err := s.renderExpertFiles(
		installed,
		previousLayer,
		nil,
		false,
	)
	if err != nil {
		return nil, err
	}
	candidate := *installed
	applyMarketReleaseUpdate(&candidate, update)
	layer := s.buildAgentfileLayer(ctx, &candidate)
	if err := s.commitExpertChanges(ctx, &candidate, layer, nil); err != nil {
		return nil, err
	}
	repoName := s.gitops.RepoNameFromPath(*installed.GitRepoPath)
	branch := s.branchOf(installed)
	return func(rollbackContext context.Context) error {
		return s.gitops.Commit(
			rollbackContext,
			repoName,
			branch,
			"rollback: restore expert configuration",
			gitops.Author{},
			previousChanges,
		)
	}, nil
}

func applyMarketReleaseUpdate(
	expert *expertdom.Expert,
	update expertdom.MarketReleaseUpdate,
) {
	expert.Name = update.Name
	expert.Description = update.Description
	expert.AgentSlug = update.AgentSlug
	expert.Prompt = update.Prompt
	expert.InteractionMode = update.InteractionMode
	expert.AutomationLevel = update.AutomationLevel
	expert.Perpetual = update.Perpetual
	expert.UsedEnvBundles = update.UsedEnvBundles
	expert.SkillSlugs = update.SkillSlugs
	expert.KnowledgeMounts = update.KnowledgeMounts
	expert.ConfigOverrides = update.ConfigOverrides
	expert.AgentfileLayer = nil
	expert.Metadata = update.Metadata
	expert.WorkerSpecSnapshotID = &update.WorkerSpecSnapshotID
	expert.SourceMarketReleaseID = &update.SourceMarketReleaseID
}
