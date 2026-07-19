package expert

import (
	"context"
	"errors"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
)

func (s *Service) InstallPublishedMarketApplication(
	ctx context.Context,
	req InstallMarketApplicationRequest,
) (*expertdom.Expert, bool, error) {
	if s.market == nil || s.marketInstallLock == nil {
		return nil, false, ErrMarketUnavailable
	}
	application, err := s.market.GetApplicationBySlug(ctx, req.MarketSlug)
	if err != nil {
		if errors.Is(err, expertmarket.ErrNotFound) {
			return nil, false, ErrMarketApplicationNotFound
		}
		return nil, false, err
	}
	var installed *expertdom.Expert
	var existing bool
	err = s.marketInstallLock.WithinMarketApplicationLock(
		ctx,
		application.ID,
		func() error {
			published, loadErr := s.publishedMarketApplicationByID(
				ctx,
				application.ID,
			)
			if loadErr != nil {
				return loadErr
			}
			return s.marketInstallLock.WithinMarketInstallationLock(
				ctx,
				req.OrganizationID,
				application.ID,
				func() error {
					var installErr error
					installed, existing, installErr =
						s.installPublishedMarketRelease(ctx, req, published)
					return installErr
				},
			)
		},
	)
	return installed, existing, err
}

func (s *Service) installPublishedMarketRelease(
	ctx context.Context,
	req InstallMarketApplicationRequest,
	published *PublishedMarketApplication,
) (*expertdom.Expert, bool, error) {
	existing, err := s.store.GetByMarketApplication(
		ctx,
		req.OrganizationID,
		published.Application.ID,
	)
	if err == nil {
		return existing, true, nil
	}
	if !errors.Is(err, expertdom.ErrNotFound) {
		return nil, false, err
	}
	snapshot, workerSnapshotID, err := s.prepareMarketInstallation(
		ctx,
		req.OrganizationID,
		req.OrganizationSlug,
		req.UserID,
		req.ModelResourceID,
		req.ToolModelResourceIDs,
		&published.Release,
	)
	if err != nil {
		return nil, false, err
	}
	expertType := published.Release.Category
	row, err := s.Create(ctx, &CreateExpertRequest{
		OrganizationID:            req.OrganizationID,
		UserID:                    req.UserID,
		Name:                      snapshot.Name,
		Slug:                      string(published.Application.Slug),
		Description:               snapshot.Description,
		AgentSlug:                 snapshot.AgentSlug,
		Prompt:                    snapshot.Prompt,
		InteractionMode:           snapshot.InteractionMode,
		AutomationLevel:           snapshot.AutomationLevel,
		Perpetual:                 snapshot.Perpetual,
		UsedEnvBundles:            snapshot.UsedEnvBundles,
		SkillSlugs:                snapshot.SkillSlugs,
		KnowledgeMounts:           snapshot.KnowledgeMounts,
		ConfigOverrides:           snapshot.ConfigOverrides,
		WorkerSpecSnapshotID:      &workerSnapshotID,
		SourceMarketApplicationID: &published.Application.ID,
		SourceMarketReleaseID:     &published.Release.ID,
		Metadata:                  snapshot.Metadata,
		ExpertType:                &expertType,
	})
	if err == nil {
		return row, false, nil
	}
	cleanupCtx, cancelCleanup := marketCleanupContext(ctx)
	defer cancelCleanup()
	if cleanupErr := s.removeUnusedMarketSnapshot(
		cleanupCtx,
		req.OrganizationID,
		workerSnapshotID,
	); cleanupErr != nil {
		return nil, false, errors.Join(err, cleanupErr)
	}
	existing, getErr := s.store.GetByMarketApplication(
		ctx,
		req.OrganizationID,
		published.Application.ID,
	)
	if getErr == nil {
		return existing, true, nil
	}
	return nil, false, err
}
