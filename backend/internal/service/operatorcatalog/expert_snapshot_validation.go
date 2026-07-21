package operatorcatalog

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	specdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
	specservice "github.com/l8ai-cn/agentcloud/backend/internal/service/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
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
	if !workerConfigMatchesCatalog(spec.TypeConfig) {
		return bootstrapper.rebuildExpertSnapshotForDefinition(
			ctx,
			request,
			definition,
			expert,
			expectedSkillIDs,
		)
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
	return bootstrapper.rebuildOrBackfillExpertArtifact(
		ctx,
		request,
		definition,
		expert,
		expectedSkillIDs,
		artifactFound,
		snapshot.ID,
	)
}

func workerConfigMatchesCatalog(config specdom.TypeConfig) bool {
	return config.InteractionMode == specdom.InteractionModePTY &&
		config.AutomationLevel == specdom.AutomationLevelAutoEdit &&
		reflect.DeepEqual(config.Values, videoExpertConfigOverrides())
}

func (bootstrapper *Bootstrapper) rebuildExpertSnapshotForDefinition(
	ctx context.Context,
	request BootstrapRequest,
	definition ExpertDefinition,
	expert *expertdom.Expert,
	expectedSkillIDs []int64,
) error {
	prepared, err := bootstrapper.prepareExpertSnapshot(
		ctx,
		request,
		definition,
		expectedSkillIDs,
	)
	if err != nil {
		return err
	}
	return bootstrapper.rebuildExpertSnapshot(ctx, request, expert, prepared)
}

func (bootstrapper *Bootstrapper) rebuildOrBackfillExpertArtifact(
	ctx context.Context,
	request BootstrapRequest,
	definition ExpertDefinition,
	expert *expertdom.Expert,
	expectedSkillIDs []int64,
	artifactFound bool,
	snapshotID int64,
) error {
	prepared, err := bootstrapper.prepareExpertSnapshot(
		ctx,
		request,
		definition,
		expectedSkillIDs,
	)
	if err != nil {
		return err
	}
	if artifactFound {
		return bootstrapper.rebuildExpertSnapshot(ctx, request, expert, prepared)
	}
	return bootstrapper.createSnapshotArtifact(ctx, request, snapshotID, prepared)
}

func (bootstrapper *Bootstrapper) prepareExpertSnapshot(
	ctx context.Context,
	request BootstrapRequest,
	definition ExpertDefinition,
	expectedSkillIDs []int64,
) (workercreation.Prepared, error) {
	workerType, err := slugkit.NewFromTrusted("video-studio")
	if err != nil {
		return workercreation.Prepared{}, err
	}
	return bootstrapper.workers.Prepare(
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
}

func artifactMatchesInstructionContract(document workerdependency.Document) bool {
	source := strings.TrimSpace(document.Worker.AgentfileSource)
	return source != "" &&
		!strings.Contains("\n"+source+"\n", "\nPROMPT ") &&
		strings.Contains(source, `"/AGENTS.md"`)
}
