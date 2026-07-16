package orchestrationresource

import (
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
)

const workerValidationDigest = "sha256:" +
	"0000000000000000000000000000000000000000000000000000000000000000"

func validateWorkerTemplateSemantics(spec WorkerTemplateSpec) error {
	return workerspec.Validate(workerspec.Spec{
		Version: workerspec.VersionV1,
		Runtime: workerspec.Runtime{
			WorkerType: workerspec.WorkerType{
				Slug:           spec.WorkerType,
				DefinitionHash: strings.Repeat("0", 64),
			},
			Image: workerspec.RuntimeImage{
				ID:     spec.Runtime.RuntimeImageID,
				Digest: workerValidationDigest,
			},
		},
		Placement:  workerTemplateValidationPlacement(spec.Runtime),
		TypeConfig: workerTemplateValidationTypeConfig(spec.TypeConfig),
		Workspace:  workerTemplateValidationWorkspace(spec.Workspace),
		Lifecycle: workerspec.Lifecycle{
			TerminationPolicy:  spec.Lifecycle.TerminationPolicy,
			IdleTimeoutMinutes: spec.Lifecycle.IdleTimeoutMinutes,
		},
		Metadata: workerspec.Metadata{Alias: spec.Metadata.Alias},
	})
}

func workerTemplateValidationPlacement(
	runtime WorkerTemplateRuntimeSpec,
) workerspec.Placement {
	profile := workerspec.ResourceProfile{
		Resources: workerspec.ResourceRequestsLimits{
			CPURequestMilliCPU: 1,
			CPULimitMilliCPU:   1,
			MemoryRequestBytes: 1,
			MemoryLimitBytes:   1,
		},
	}
	switch {
	case runtime.CustomResources != nil:
		profile.Custom = true
		profile.Resources = *runtime.CustomResources
	case runtime.ResourceProfileRef != nil:
		profile.ID = 1
	}
	return workerspec.Placement{
		Policy: runtime.PlacementPolicy,
		ComputeTarget: workerspec.ComputeTarget{
			ID:   1,
			Kind: workerspec.ComputeTargetKindRunnerPool,
		},
		DeploymentMode:  runtime.DeploymentMode,
		ResourceProfile: profile,
	}
}

func workerTemplateValidationTypeConfig(
	config WorkerTemplateTypeConfigSpec,
) workerspec.TypeConfig {
	var secretRefs map[string]workerspec.SecretReference
	if config.SecretRefs != nil {
		secretRefs = make(
			map[string]workerspec.SecretReference,
			len(config.SecretRefs),
		)
		id := int64(1)
		for field := range config.SecretRefs {
			secretRefs[field] = workerspec.SecretReference{
				Kind: "env-bundle",
				ID:   id,
			}
			id++
		}
	}
	return workerspec.TypeConfig{
		SchemaVersion:   config.SchemaVersion,
		Values:          config.Values,
		SecretRefs:      secretRefs,
		InteractionMode: config.InteractionMode,
		AutomationLevel: config.AutomationLevel,
	}
}

func workerTemplateValidationWorkspace(
	workspace WorkerTemplateWorkspaceSpec,
) workerspec.Workspace {
	var repositoryID *int64
	if workspace.RepositoryRef != nil {
		value := int64(1)
		repositoryID = &value
	}
	skillIDs := make([]int64, len(workspace.SkillRefs))
	for index := range skillIDs {
		skillIDs[index] = int64(index + 1)
	}
	knowledgeMounts := make(
		[]workerspec.KnowledgeMount,
		len(workspace.KnowledgeMounts),
	)
	for index, mount := range workspace.KnowledgeMounts {
		knowledgeMounts[index] = workerspec.KnowledgeMount{
			KnowledgeBaseID: int64(index + 1),
			Mode:            mount.Mode,
		}
	}
	envBundleIDs := make(
		[]workerspec.RuntimeEnvBundleID,
		len(workspace.EnvironmentBundleRefs),
	)
	for index := range envBundleIDs {
		envBundleIDs[index] = workerspec.RuntimeEnvBundleID(index + 1)
	}
	return workerspec.Workspace{
		RepositoryID:    repositoryID,
		Branch:          workspace.Branch,
		SkillIDs:        skillIDs,
		KnowledgeMounts: knowledgeMounts,
		EnvBundleIDs:    envBundleIDs,
		Instructions:    workspace.Instructions,
	}
}
