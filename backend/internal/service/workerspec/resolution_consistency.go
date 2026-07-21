package workerspec

import (
	"fmt"

	workerruntime "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerruntime"
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
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
	if err := ValidateModelRequirement(resolved.ModelRequirement); err != nil {
		return err
	}
	return ValidateToolModelRequirements(resolved.ToolModelRequirements)
}

func ValidateModelRequirement(requirement domain.ModelRequirement) error {
	if !requirement.Required {
		if len(requirement.ProtocolAdapters) != 0 {
			return fmt.Errorf("non-model worker type declared protocol adapters")
		}
		return nil
	}
	if len(requirement.ProtocolAdapters) == 0 {
		return fmt.Errorf("model worker type has no protocol adapters")
	}
	for _, adapter := range requirement.ProtocolAdapters {
		if err := slugkit.Validate(adapter.String()); err != nil {
			return fmt.Errorf("model worker protocol adapter: %w", err)
		}
	}
	return nil
}

func ValidateToolModelRequirements(requirements []domain.ToolModelRequirement) error {
	roles := make(map[slugkit.Slug]struct{}, len(requirements))
	for _, requirement := range requirements {
		if err := slugkit.Validate(requirement.Role.String()); err != nil {
			return fmt.Errorf("tool model role: %w", err)
		}
		if _, exists := roles[requirement.Role]; exists {
			return fmt.Errorf("duplicate tool model role %q", requirement.Role)
		}
		roles[requirement.Role] = struct{}{}
		if len(requirement.ProviderKeys) == 0 ||
			len(requirement.ProtocolAdapters) == 0 ||
			!requirement.Modality.Valid() ||
			!requirement.Capability.Valid() {
			return fmt.Errorf("tool model %q is incomplete", requirement.Role)
		}
		for _, provider := range requirement.ProviderKeys {
			if err := slugkit.Validate(provider.String()); err != nil {
				return fmt.Errorf("tool model provider: %w", err)
			}
		}
		for _, adapter := range requirement.ProtocolAdapters {
			if err := slugkit.Validate(adapter.String()); err != nil {
				return fmt.Errorf("tool model protocol adapter: %w", err)
			}
		}
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

func ValidateRuntimeSelection(
	selection RuntimeSelection,
	resolved workerruntime.Resolved,
) error {
	return validateRuntimeResolution(selection, resolved)
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
	cloned.SkillPackages = append(
		[]domain.SkillPackageBinding{},
		workspace.SkillPackages...,
	)
	cloned.KnowledgeMounts = append(
		[]domain.KnowledgeMount{},
		workspace.KnowledgeMounts...,
	)
	cloned.EnvBundleIDs = append(
		[]domain.RuntimeEnvBundleID{},
		workspace.EnvBundleIDs...,
	)
	cloned.ConfigDocumentBindings = append(
		[]domain.ConfigDocumentBinding{},
		workspace.ConfigDocumentBindings...,
	)
	return cloned
}
