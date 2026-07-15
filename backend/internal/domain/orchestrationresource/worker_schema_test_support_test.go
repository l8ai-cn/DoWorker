package orchestrationresource

import (
	"encoding/json"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func workerSchemaManifest(t *testing.T, kind string, spec any) Manifest {
	t.Helper()
	raw, err := json.Marshal(spec)
	require.NoError(t, err)
	return Manifest{
		TypeMeta: TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: kind},
		Metadata: Metadata{
			Name:      slugkit.MustNewForTest("resource-one"),
			Namespace: slugkit.MustNewForTest("team-alpha"),
		},
		Spec: raw,
	}
}

func workerDraftReference(kind, name string) Reference {
	return Reference{
		Kind: kind,
		Name: slugkit.MustNewForTest(name),
	}
}

func validWorkerTemplateSpec() WorkerTemplateSpec {
	modelRef := workerDraftReference(KindModelBinding, "coding-primary")
	resourceProfileRef := workerDraftReference(
		KindResourceProfile,
		"balanced-profile",
	)
	repositoryRef := workerDraftReference(
		KindRepository,
		"agents-mesh",
	)
	return WorkerTemplateSpec{
		OptionsRevision: "runtime-catalog-7",
		WorkerType:      slugkit.MustNewForTest("codex"),
		ModelRef:        &modelRef,
		ToolRefs: map[string]Reference{
			"video-generation": workerDraftReference(
				KindToolBinding,
				"seedance-tool",
			),
		},
		Runtime: WorkerTemplateRuntimeSpec{
			RuntimeImageID:  31,
			PlacementPolicy: workerspec.PlacementPolicyExplicit,
			ComputeTargetRef: workerDraftReference(
				KindComputeTarget,
				"primary-pool",
			),
			DeploymentMode:     workerspec.DeploymentModeDedicated,
			ResourceProfileRef: &resourceProfileRef,
		},
		TypeConfig: WorkerTemplateTypeConfigSpec{
			SchemaVersion: 2,
			Values: map[string]any{
				"reasoning-effort": "high",
			},
			SecretRefs: map[string]Reference{
				"api-token": workerDraftReference(
					KindEnvironmentBundle,
					"codex-secrets",
				),
			},
			InteractionMode: workerspec.InteractionModeACP,
			AutomationLevel: workerspec.AutomationLevelAutoEdit,
		},
		Workspace: WorkerTemplateWorkspaceSpec{
			RepositoryRef: &repositoryRef,
			Branch:        "main",
			SkillRefs: []Reference{
				workerDraftReference(KindSkill, "code-review"),
			},
			KnowledgeMounts: []WorkerTemplateKnowledgeMount{
				{
					Ref: workerDraftReference(
						KindKnowledgeBase,
						"engineering-docs",
					),
					Mode: workerspec.KnowledgeMountReadOnly,
				},
			},
			EnvironmentBundleRefs: []Reference{
				workerDraftReference(
					KindEnvironmentBundle,
					"runtime-environment",
				),
			},
			ConfigBundleRefs: []Reference{
				workerDraftReference(
					KindEnvironmentBundle,
					"codex-configuration",
				),
			},
			Instructions: "Review the repository before editing.",
		},
		Lifecycle: WorkerTemplateLifecycleSpec{
			TerminationPolicy:  workerspec.TerminationPolicyOnIdle,
			IdleTimeoutMinutes: 30,
		},
		Metadata: WorkerTemplateMetadataSpec{Alias: "Reviewer"},
	}
}
