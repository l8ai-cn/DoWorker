package expert

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func (s *Service) prepareMarketInstallation(
	ctx context.Context,
	organizationID int64,
	organizationSlug string,
	userID int64,
	modelResourceID int64,
	toolModelResourceIDs map[string]int64,
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
	targetSlug, err := slugkit.NewFromTrusted(organizationSlug)
	if err != nil {
		return marketExpertSnapshot{}, 0, err
	}
	resolved, err := s.marketWorkerSpecs.PrepareMarketSnapshot(
		ctx,
		specservice.Scope{
			OrgID: organizationID, OrgSlug: targetSlug, UserID: userID,
		},
		workerSnapshot.Spec,
		modelResourceID,
		toolModelResourceIDs,
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

func marketToolModelResourceIDs(
	bindings []specdomain.ToolModelBinding,
) map[string]int64 {
	if len(bindings) == 0 {
		return nil
	}
	resourceIDs := make(map[string]int64, len(bindings))
	for _, binding := range bindings {
		resourceIDs[binding.Role.String()] = binding.ModelBinding.ResourceID
	}
	return resourceIDs
}

func (s *Service) removeUnusedMarketSnapshot(
	ctx context.Context,
	organizationID, snapshotID int64,
) error {
	return s.workerSpecWriter.Delete(ctx, organizationID, snapshotID)
}
