package workercreation

import (
	"context"
	"errors"
	"fmt"

	skilldomain "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	workerspecdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workerspec "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

func (service *Service) PrepareMarketSnapshot(
	ctx context.Context,
	scope workerspec.Scope,
	source workerspecdomain.Spec,
	modelResourceID int64,
	toolModelResourceIDs map[string]int64,
) (workerspec.ResolvedSnapshot, error) {
	normalized, err := workerspecdomain.NormalizeAndValidate(source)
	if err != nil {
		return workerspec.ResolvedSnapshot{}, err
	}
	if modelResourceID <= 0 {
		return workerspec.ResolvedSnapshot{}, invalidMarketSnapshotDraft(
			"model_resource_id",
			"must be positive",
		)
	}
	if scope.OrgSlug == "" {
		return workerspec.ResolvedSnapshot{}, workerspec.ErrInvalidScope
	}
	if err := rejectPrivateMarketSnapshotReferences(normalized); err != nil {
		return workerspec.ResolvedSnapshot{}, err
	}
	skillIDs := append([]int64{}, normalized.Workspace.SkillIDs...)
	if len(normalized.Workspace.SkillPackages) == 0 {
		skillIDs, err = service.platformSkillIDs(ctx, skillIDs)
		if err != nil {
			return workerspec.ResolvedSnapshot{}, err
		}
	}
	runtimeSelection := workerspec.RuntimeSelection{
		RuntimeImageID:    normalized.Runtime.Image.ID,
		PlacementPolicy:   normalized.Placement.Policy,
		ComputeTargetID:   normalized.Placement.ComputeTarget.ID,
		DeploymentMode:    normalized.Placement.DeploymentMode,
		ResourceProfileID: normalized.Placement.ResourceProfile.ID,
	}
	if normalized.Placement.ResourceProfile.Custom {
		resources := normalized.Placement.ResourceProfile.Resources
		runtimeSelection.CustomResources = &resources
	}
	prepared, err := service.Prepare(ctx, scope, Draft{
		OptionsRevision:  service.Revision(),
		OrganizationSlug: scope.OrgSlug,
		WorkerSpec: workerspec.Draft{
			ModelResourceID:      modelResourceID,
			ToolModelResourceIDs: cloneToolModelResourceIDs(toolModelResourceIDs),
			WorkerTypeSlug:       normalized.Runtime.WorkerType.Slug,
			Runtime:              runtimeSelection,
			TypeConfig:           normalized.TypeConfig,
			Workspace: workerspecdomain.Workspace{
				SkillIDs: skillIDs,
				SkillPackages: append(
					[]workerspecdomain.SkillPackageBinding{},
					normalized.Workspace.SkillPackages...,
				),
				Instructions: normalized.Workspace.Instructions,
				InitialTask:  normalized.Workspace.InitialTask,
			},
			Lifecycle: normalized.Lifecycle,
			Metadata:  workerspecdomain.Metadata{Alias: normalized.Metadata.Alias},
		},
	})
	if err != nil {
		return workerspec.ResolvedSnapshot{}, err
	}
	return prepared.Snapshot, nil
}

func cloneToolModelResourceIDs(source map[string]int64) map[string]int64 {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[string]int64, len(source))
	for role, resourceID := range source {
		cloned[role] = resourceID
	}
	return cloned
}

func rejectPrivateMarketSnapshotReferences(spec workerspecdomain.Spec) error {
	switch {
	case spec.Workspace.RepositoryID != nil:
		return invalidMarketSnapshotDraft(
			"workspace.repository_id",
			"market snapshots cannot carry repository references",
		)
	case len(spec.Workspace.KnowledgeMounts) > 0:
		return invalidMarketSnapshotDraft(
			"workspace.knowledge_mounts",
			"market snapshots cannot carry knowledge mounts",
		)
	case len(spec.Workspace.EnvBundleIDs) > 0:
		return invalidMarketSnapshotDraft(
			"workspace.env_bundle_ids",
			"market snapshots cannot carry environment bundles",
		)
	case len(spec.Workspace.ConfigBundleIDs) > 0:
		return invalidMarketSnapshotDraft(
			"workspace.config_bundle_ids",
			"market snapshots cannot carry configuration bundles",
		)
	case len(spec.TypeConfig.SecretRefs) > 0:
		return invalidMarketSnapshotDraft(
			"type_config.secret_refs",
			"market snapshots cannot carry secret references",
		)
	default:
		return nil
	}
}

func (service *Service) platformSkillIDs(
	ctx context.Context,
	ids []int64,
) ([]int64, error) {
	if len(ids) == 0 {
		return []int64{}, nil
	}
	if service == nil || service.workspaceDeps.Skills == nil {
		return nil, workerspec.ErrResolverUnavailable
	}
	platformIDs := make([]int64, 0, len(ids))
	for _, id := range ids {
		row, err := service.workspaceDeps.Skills.GetAnyByID(ctx, id)
		if err != nil {
			if errors.Is(err, skilldomain.ErrNotFound) {
				return nil, invalidMarketSnapshotDraft(
					"workspace.skill_ids",
					fmt.Sprintf("skill %d not found", id),
				)
			}
			return nil, err
		}
		if row == nil || row.ID != id {
			return nil, invalidMarketSnapshotDraft(
				"workspace.skill_ids",
				fmt.Sprintf("skill %d is invalid", id),
			)
		}
		if !row.IsPlatformLevel() || !row.IsActive ||
			row.ContentSha == "" || row.StorageKey == "" {
			return nil, invalidMarketSnapshotDraft(
				"workspace.skill_ids",
				fmt.Sprintf("skill %d is not an active packaged platform skill", id),
			)
		}
		platformIDs = append(platformIDs, id)
	}
	return platformIDs, nil
}

func invalidMarketSnapshotDraft(field, reason string) error {
	return &workerspec.InvalidDraftFieldError{
		Field:  field,
		Reason: reason,
	}
}
