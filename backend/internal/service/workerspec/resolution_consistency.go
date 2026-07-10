package workerspec

import (
	"fmt"

	workerruntime "github.com/anthropics/agentsmesh/backend/internal/domain/workerruntime"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func validateScope(scope Scope) error {
	if scope.OrgID <= 0 || scope.UserID <= 0 {
		return ErrInvalidScope
	}
	return nil
}

func validateWorkerTypeResolution(
	requested slugkit.Slug,
	resolved WorkerTypeResolution,
) error {
	if resolved.WorkerType.Slug != requested {
		return fmt.Errorf(
			"worker type resolver substituted %q with %q",
			requested,
			resolved.WorkerType.Slug,
		)
	}
	return nil
}

func validateRuntimeResolution(
	selection RuntimeSelection,
	resolved workerruntime.Resolved,
) error {
	switch {
	case resolved.RuntimeImage.ID != selection.RuntimeImageID:
		return fmt.Errorf("runtime resolver substituted runtime image")
	case resolved.Placement.Policy != selection.PlacementPolicy:
		return fmt.Errorf("runtime resolver substituted placement policy")
	case resolved.Placement.ComputeTarget.ID != selection.ComputeTargetID:
		return fmt.Errorf("runtime resolver substituted compute target")
	case resolved.Placement.DeploymentMode != selection.DeploymentMode:
		return fmt.Errorf("runtime resolver substituted deployment mode")
	case resolved.Placement.ResourceProfile.ID != selection.ResourceProfileID:
		return fmt.Errorf("runtime resolver substituted resource profile")
	default:
		return nil
	}
}

func validateModelResolution(
	requestedResourceID int64,
	binding domain.ModelBinding,
) error {
	if binding.ResourceID != requestedResourceID {
		return fmt.Errorf("model resolver substituted model resource")
	}
	return nil
}

func cloneWorkspace(workspace domain.Workspace) domain.Workspace {
	cloned := workspace
	if workspace.RepositoryID != nil {
		repositoryID := *workspace.RepositoryID
		cloned.RepositoryID = &repositoryID
	}
	cloned.SkillIDs = append([]int64{}, workspace.SkillIDs...)
	cloned.KnowledgeMounts = append(
		[]domain.KnowledgeMount{},
		workspace.KnowledgeMounts...,
	)
	cloned.EnvBundleIDs = append(
		[]domain.RuntimeEnvBundleID{},
		workspace.EnvBundleIDs...,
	)
	return cloned
}
