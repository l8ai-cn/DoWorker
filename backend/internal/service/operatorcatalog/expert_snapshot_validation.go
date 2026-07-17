package operatorcatalog

import (
	"context"
	"slices"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdom "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

func (bootstrapper *Bootstrapper) validateExpertSnapshot(
	ctx context.Context,
	request BootstrapRequest,
	definition ExpertDefinition,
	expert *expertdom.Expert,
	skills map[string]*skilldom.Skill,
) error {
	if expert.WorkerSpecSnapshotID == nil {
		return ErrCatalogConflict
	}
	snapshot, err := bootstrapper.snapshots.GetByID(
		ctx,
		request.OrganizationID,
		*expert.WorkerSpecSnapshotID,
	)
	if err != nil {
		return err
	}
	expectedSkillIDs := make([]int64, 0, len(definition.SkillSlugs))
	for _, slug := range definition.SkillSlugs {
		row := skills[slug]
		if row == nil {
			return ErrCatalogConflict
		}
		expectedSkillIDs = append(expectedSkillIDs, row.ID)
	}
	slices.Sort(expectedSkillIDs)
	spec := snapshot.Spec
	if snapshot.ID != *expert.WorkerSpecSnapshotID ||
		snapshot.OrganizationID != request.OrganizationID ||
		!specdom.HasResolvedProtocolAdapters(spec) ||
		spec.Runtime.ModelBinding.ResourceID != request.ModelResourceID ||
		spec.Runtime.Image.ID != request.RuntimeImageID ||
		spec.Runtime.WorkerType.Slug.String() != "video-studio" ||
		spec.Workspace.Instructions != definition.Prompt ||
		!slices.Equal(spec.Workspace.SkillIDs, expectedSkillIDs) {
		return ErrCatalogConflict
	}
	return nil
}
