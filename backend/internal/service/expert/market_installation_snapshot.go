package expert

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

func (s *Service) prepareMarketInstallation(
	ctx context.Context,
	organizationID, userID, modelResourceID int64,
	release *expertmarket.Release,
) (marketExpertSnapshot, int64, error) {
	if s.workerSpecWriter == nil || s.marketWorkerSpecs == nil {
		return marketExpertSnapshot{}, 0, ErrMarketUnavailable
	}
	expertSnapshot, workerSnapshot, err := decodeMarketReleaseSnapshots(release)
	if err != nil {
		return marketExpertSnapshot{}, 0, err
	}
	if err := validateMarketSkillDependencies(
		release.SkillDependencies,
		workerSnapshot.Spec.Workspace.SkillPackages,
	); err != nil {
		return marketExpertSnapshot{}, 0, err
	}
	resolved, err := s.marketWorkerSpecs.PrepareMarketSnapshot(
		ctx,
		specservice.Scope{OrgID: organizationID, UserID: userID},
		workerSnapshot.Spec,
		modelResourceID,
	)
	if err != nil {
		return marketExpertSnapshot{}, 0, err
	}
	created, err := s.workerSpecWriter.Create(ctx, resolved)
	if err != nil {
		return marketExpertSnapshot{}, 0, err
	}
	if created.ID <= 0 || created.OrganizationID != organizationID {
		return marketExpertSnapshot{}, 0, ErrMarketSnapshotInvalid
	}
	return expertSnapshot, created.ID, nil
}

func (s *Service) removeUnusedMarketSnapshot(
	ctx context.Context,
	organizationID, snapshotID int64,
) error {
	return s.workerSpecWriter.Delete(ctx, organizationID, snapshotID)
}
