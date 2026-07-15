package operatorcatalog

import (
	"context"
	"errors"
	"fmt"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func (bootstrapper *Bootstrapper) ensureExpert(
	ctx context.Context,
	request BootstrapRequest,
	definition ExpertDefinition,
	skills map[string]*skilldom.Skill,
) (bool, bool, error) {
	row, err := bootstrapper.experts.GetBySlug(
		ctx,
		request.OrganizationID,
		definition.Slug,
	)
	created := false
	switch {
	case err == nil:
		if !expertMatches(row, definition) {
			return false, false, ErrCatalogConflict
		}
		if err := bootstrapper.validateExpertSnapshot(
			ctx,
			request,
			definition,
			row,
			skills,
		); err != nil {
			return false, false, err
		}
	case errors.Is(err, expertdom.ErrNotFound):
		row, err = bootstrapper.createExpert(
			ctx,
			request,
			definition,
			skills,
		)
		if err != nil {
			return false, false, err
		}
		created = true
	default:
		return false, false, err
	}
	published, err := bootstrapper.ensurePublished(
		ctx,
		request,
		definition,
		row,
	)
	return created, published, err
}

func (bootstrapper *Bootstrapper) createExpert(
	ctx context.Context,
	request BootstrapRequest,
	definition ExpertDefinition,
	skills map[string]*skilldom.Skill,
) (*expertdom.Expert, error) {
	skillIDs := make([]int64, 0, len(definition.SkillSlugs))
	for _, slug := range definition.SkillSlugs {
		row := skills[slug]
		if row == nil {
			return nil, fmt.Errorf("%w: skill %s is missing", ErrCatalogConflict, slug)
		}
		skillIDs = append(skillIDs, row.ID)
	}
	workerType, err := slugkit.NewFromTrusted("video-studio")
	if err != nil {
		return nil, err
	}
	prepared, err := bootstrapper.workers.Prepare(
		ctx,
		specservice.Scope{
			OrgID:  request.OrganizationID,
			UserID: request.PublisherUserID,
		},
		workerDraft(
			bootstrapper.workers.Revision(),
			request,
			definition,
			skillIDs,
			workerType,
		),
	)
	if err != nil {
		return nil, err
	}
	snapshot, err := bootstrapper.snapshots.Create(ctx, prepared.Snapshot)
	if err != nil {
		return nil, err
	}
	description := definition.Description
	prompt := definition.Prompt
	expert, err := bootstrapper.experts.Create(ctx, &expertsvc.CreateExpertRequest{
		OrganizationID:       request.OrganizationID,
		UserID:               request.PublisherUserID,
		Name:                 definition.Name,
		Slug:                 definition.Slug,
		Description:          &description,
		AgentSlug:            "video-studio",
		Prompt:               &prompt,
		InteractionMode:      expertdom.InteractionModePTY,
		AutomationLevel:      expertdom.AutomationLevelAutoEdit,
		SkillSlugs:           definition.SkillSlugs,
		WorkerSpecSnapshotID: &snapshot.ID,
		ExpertType:           stringPointer("video"),
	})
	if err == nil {
		return expert, nil
	}
	cleanupErr := bootstrapper.snapshots.Delete(
		context.WithoutCancel(ctx),
		request.OrganizationID,
		snapshot.ID,
	)
	return nil, errors.Join(err, cleanupErr)
}

func workerDraft(
	revision string,
	request BootstrapRequest,
	definition ExpertDefinition,
	skillIDs []int64,
	workerType slugkit.Slug,
) workercreation.Draft {
	return workercreation.Draft{
		OptionsRevision: revision,
		WorkerSpec: specservice.Draft{
			ModelResourceID: request.ModelResourceID,
			WorkerTypeSlug:  workerType,
			Runtime: specservice.RuntimeSelection{
				RuntimeImageID:    request.RuntimeImageID,
				PlacementPolicy:   specdomain.PlacementPolicyExplicit,
				ComputeTargetID:   1,
				DeploymentMode:    specdomain.DeploymentModePooled,
				ResourceProfileID: 2,
			},
			TypeConfig: specdomain.TypeConfig{
				SchemaVersion:   1,
				Values:          map[string]any{},
				SecretRefs:      map[string]specdomain.SecretReference{},
				InteractionMode: specdomain.InteractionModePTY,
				AutomationLevel: specdomain.AutomationLevelAutoEdit,
			},
			Workspace: specdomain.Workspace{
				SkillIDs:     skillIDs,
				Instructions: definition.Prompt,
			},
			Lifecycle: specdomain.Lifecycle{
				TerminationPolicy: specdomain.TerminationPolicyManual,
			},
			Metadata: specdomain.Metadata{Alias: definition.Slug},
		},
	}
}

func stringPointer(value string) *string { return &value }
