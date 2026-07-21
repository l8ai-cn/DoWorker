package expert

import (
	"context"
	"errors"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
)

func (s *Service) UpgradeMarketApplication(
	ctx context.Context,
	req UpgradeMarketApplicationRequest,
) (*expertdom.Expert, bool, error) {
	installed, err := s.store.GetByID(ctx, req.OrganizationID, req.ExpertID)
	if err != nil {
		return nil, false, err
	}
	if installed.SourceMarketApplicationID == nil {
		return nil, false, expertdom.ErrNotFound
	}
	if s.marketInstallLock == nil {
		return nil, false, ErrMarketUnavailable
	}
	applicationID := *installed.SourceMarketApplicationID
	var upgraded *expertdom.Expert
	var changed bool
	err = s.marketInstallLock.WithinMarketApplicationLock(
		ctx,
		applicationID,
		func() error {
			return s.marketInstallLock.WithinMarketInstallationLock(
				ctx,
				req.OrganizationID,
				applicationID,
				func() error {
					var upgradeErr error
					upgraded, changed, upgradeErr =
						s.upgradeMarketApplication(ctx, req, applicationID)
					return upgradeErr
				},
			)
		},
	)
	return upgraded, changed, err
}

func (s *Service) upgradeMarketApplication(
	ctx context.Context,
	req UpgradeMarketApplicationRequest,
	applicationID int64,
) (*expertdom.Expert, bool, error) {
	installed, err := s.store.GetByID(ctx, req.OrganizationID, req.ExpertID)
	if err != nil {
		return nil, false, err
	}
	if installed.SourceMarketApplicationID == nil ||
		*installed.SourceMarketApplicationID != applicationID {
		return nil, false, expertdom.ErrNotFound
	}
	published, err := s.publishedMarketApplicationByID(
		ctx,
		applicationID,
	)
	if err != nil {
		return nil, false, err
	}
	if installed.SourceMarketReleaseID != nil &&
		*installed.SourceMarketReleaseID == published.Release.ID {
		return installed, false, nil
	}
	if installed.WorkerSpecSnapshotID == nil {
		return nil, false, ErrMarketSnapshotInvalid
	}
	currentSnapshot, err := s.workerSpecs.GetByID(
		ctx,
		req.OrganizationID,
		*installed.WorkerSpecSnapshotID,
	)
	if err != nil {
		return nil, false, err
	}
	snapshot, workerSnapshotID, err := s.prepareMarketInstallation(
		ctx,
		req.OrganizationID,
		req.OrganizationSlug,
		req.UserID,
		currentSnapshot.Spec.Runtime.ModelBinding.ResourceID,
		marketToolModelResourceIDs(
			currentSnapshot.Spec.Runtime.ToolModelBindings,
		),
		&published.Release,
	)
	if err != nil {
		return nil, false, err
	}
	update := marketReleaseUpdate(
		snapshot,
		workerSnapshotID,
		published.Release.ID,
		published.Release.Category,
	)
	update.ExpectedRevision = installed.Revision
	rollbackGit, err := s.commitMarketUpgrade(ctx, installed, update)
	if err != nil {
		cleanupCtx, cancel := marketCleanupContext(ctx)
		defer cancel()
		cleanupErr := s.removeUnusedMarketSnapshot(
			cleanupCtx, req.OrganizationID, workerSnapshotID,
		)
		return nil, false, errors.Join(err, cleanupErr)
	}
	if err := s.store.UpdateMarketRelease(
		ctx,
		req.OrganizationID,
		installed.ID,
		published.Application.ID,
		update,
	); err != nil {
		cleanupCtx, cancel := marketCleanupContext(ctx)
		defer cancel()
		cleanupErr := s.removeUnusedMarketSnapshot(
			cleanupCtx, req.OrganizationID, workerSnapshotID,
		)
		var rollbackErr error
		if rollbackGit != nil {
			rollbackErr = rollbackGit(cleanupCtx)
		}
		return nil, false, errors.Join(err, cleanupErr, rollbackErr)
	}
	row, err := s.store.GetByID(ctx, req.OrganizationID, installed.ID)
	return row, err == nil, err
}
