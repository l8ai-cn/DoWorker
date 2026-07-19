package operatorcatalog

import (
	"context"
	"errors"
	"slices"
	"strings"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
	specdom "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"gorm.io/gorm"
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
	artifact, err := bootstrapper.artifacts.GetBySnapshotID(
		ctx,
		request.OrganizationID,
		snapshot.ID,
	)
	artifactFound := err == nil
	if artifactFound && artifactMatchesInstructionContract(artifact) {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	workerType, err := slugkit.NewFromTrusted("video-studio")
	if err != nil {
		return err
	}
	prepared, err := bootstrapper.workers.Prepare(
		ctx,
		specservice.Scope{
			OrgID:   request.OrganizationID,
			OrgSlug: request.OrganizationSlug,
			UserID:  request.PublisherUserID,
		},
		workerDraft(
			bootstrapper.workers.Revision(),
			request,
			definition,
			expectedSkillIDs,
			workerType,
		),
	)
	if err != nil {
		return err
	}
	if artifactFound {
		return bootstrapper.rebuildExpertSnapshot(
			ctx,
			request,
			expert,
			prepared,
		)
	}
	return bootstrapper.createSnapshotArtifact(ctx, request, snapshot.ID, prepared)
}

func artifactMatchesInstructionContract(document workerdependency.Document) bool {
	source := strings.TrimSpace(document.Worker.AgentfileSource)
	return source != "" &&
		!strings.Contains("\n"+source+"\n", "\nPROMPT ") &&
		strings.Contains(source, `"/AGENTS.md"`)
}
