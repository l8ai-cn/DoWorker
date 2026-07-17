package orchestrationworker

import (
	"context"
	"encoding/json"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerTemplatePlannerExtractsEveryDirectReference(t *testing.T) {
	planner := workerTemplatePlannerForTest(t)
	spec := workerTemplateSpecForTest()

	references, err := planner.References(&spec)

	require.NoError(t, err)
	require.Len(t, references, 10)
	assert.Equal(t, []string{
		"/spec/modelRef",
		"/spec/runtime/computeTargetRef",
		"/spec/runtime/resourceProfileRef",
		"/spec/toolRefs/video-generation",
		"/spec/typeConfig/secretRefs/api-token",
		"/spec/workspace/configBundleRefs/0",
		"/spec/workspace/environmentBundleRefs/0",
		"/spec/workspace/knowledgeMounts/0/ref",
		"/spec/workspace/repositoryRef",
		"/spec/workspace/skillRefs/0",
	}, referencePaths(references))
}

func TestWorkerTemplatePlannerCompilesResourceRefsIntoWorkerDraft(t *testing.T) {
	bindings := newBindingResolverStub()
	compiler := &workerCompilerStub{
		revision: "runtime-catalog-7",
		artifact: json.RawMessage(`{"version":1,"runtime":{"worker_type":{"slug":"codex"}}}`),
	}
	planner, err := NewWorkerTemplatePlanner(bindings, compiler)
	require.NoError(t, err)
	spec := workerTemplateSpecForTest()
	resolved := resolvedWorkerTemplateReferences(t, planner, spec)

	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope:              workerTemplateScope(),
		TypedSpec:          &spec,
		ResolvedReferences: resolved,
	})

	require.NoError(t, err)
	require.Empty(t, output.Issues)
	assert.Equal(t, "WorkerSpec", output.ArtifactKind)
	assert.JSONEq(t, string(compiler.artifact), string(output.ArtifactJSON))
	assert.Equal(t, compiler.revision, output.OptionsRevision)
	require.Equal(t, 1, compiler.calls)
	draft := compiler.draft.WorkerSpec
	assert.Equal(t, int64(101), draft.ModelResourceID)
	assert.Equal(t, map[string]int64{"video-generation": 102}, draft.ToolModelResourceIDs)
	assert.Equal(t, slugkit.MustNewForTest("codex"), draft.WorkerTypeSlug)
	assert.Equal(t, int64(31), draft.Runtime.RuntimeImageID)
	assert.Equal(t, int64(103), draft.Runtime.ComputeTargetID)
	assert.Equal(t, int64(104), draft.Runtime.ResourceProfileID)
	assert.Equal(t, int64(105), draft.TypeConfig.SecretRefs["api-token"].ID)
	assert.Equal(t, slugkit.MustNewForTest("env-bundle"), draft.TypeConfig.SecretRefs["api-token"].Kind)
	require.NotNil(t, draft.Workspace.RepositoryID)
	assert.Equal(t, int64(106), *draft.Workspace.RepositoryID)
	assert.Equal(t, []int64{107}, draft.Workspace.SkillIDs)
	assert.Equal(t, int64(108), draft.Workspace.KnowledgeMounts[0].KnowledgeBaseID)
	assert.Equal(t, []workerspec.RuntimeEnvBundleID{109}, draft.Workspace.EnvBundleIDs)
	assert.Equal(t, []int64{110}, draft.Workspace.ConfigBundleIDs)
	assert.Empty(t, draft.Workspace.InitialTask)
}

func TestWorkerTemplatePlannerRejectsStaleOptionsBeforeResolution(t *testing.T) {
	bindings := newBindingResolverStub()
	compiler := &workerCompilerStub{revision: "runtime-catalog-8"}
	planner, err := NewWorkerTemplatePlanner(bindings, compiler)
	require.NoError(t, err)
	spec := workerTemplateSpecForTest()

	_, err = planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope:     workerTemplateScope(),
		TypedSpec: &spec,
	})

	assert.ErrorIs(t, err, controlservice.ErrStaleOptions)
	assert.Zero(t, compiler.calls)
	assert.Zero(t, bindings.calls)
}

func TestWorkerTemplatePlannerReturnsSafeCompilerIssues(t *testing.T) {
	compiler := &workerCompilerStub{
		revision: "runtime-catalog-7",
		issues: []control.PlanIssue{{
			Severity: control.PlanIssueBlocking,
			Path:     "/spec/typeConfig/automationLevel",
			Code:     "invalid-draft",
			Message:  "Worker template contains an invalid field.",
		}},
	}
	planner, err := NewWorkerTemplatePlanner(newBindingResolverStub(), compiler)
	require.NoError(t, err)
	spec := workerTemplateSpecForTest()

	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope:              workerTemplateScope(),
		TypedSpec:          &spec,
		ResolvedReferences: resolvedWorkerTemplateReferences(t, planner, spec),
	})

	require.NoError(t, err)
	require.Len(t, output.Issues, 1)
	assert.Equal(t, "/spec/typeConfig/automationLevel", output.Issues[0].Path)
	assert.NotContains(t, output.Issues[0].Message, "secret")
	assert.Empty(t, output.ArtifactJSON)
}

func TestWorkerTemplatePlannerRejectsMissingPinnedReference(t *testing.T) {
	planner := workerTemplatePlannerForTest(t)
	spec := workerTemplateSpecForTest()
	resolved := resolvedWorkerTemplateReferences(t, planner, spec)

	_, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope:              workerTemplateScope(),
		TypedSpec:          &spec,
		ResolvedReferences: resolved[1:],
	})

	assert.ErrorIs(t, err, control.ErrCorrupt)
}

func TestPinnedReferenceIndexResolvesDistinctHistoricalRevisions(t *testing.T) {
	scope := workerTemplateScope()
	first := resolvedBindingReference(
		scope,
		resource.KindToolBinding,
		"video-tool",
	)
	second := first
	second.Revision = 2
	second.Digest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	index, err := newPinnedReferenceIndex(
		scope,
		[]control.ResolvedReference{first, second},
	)

	require.NoError(t, err)
	resolvedFirst, err := index.resolve(resource.Reference{
		Kind: resource.KindToolBinding, Name: first.Name, Revision: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), resolvedFirst.Revision)
	resolvedSecond, err := index.resolve(resource.Reference{
		Kind: resource.KindToolBinding, Name: second.Name, Revision: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), resolvedSecond.Revision)
	_, err = index.resolve(resource.Reference{
		Kind: resource.KindToolBinding, Name: first.Name,
	})
	assert.ErrorIs(t, err, control.ErrCorrupt)
}

func workerTemplateSpecForTest() resource.WorkerTemplateSpec {
	model := workerReference(resource.KindModelBinding, "coding-primary")
	profile := workerReference(resource.KindResourceProfile, "balanced-profile")
	repository := workerReference(resource.KindRepository, "agents-mesh")
	return resource.WorkerTemplateSpec{
		OptionsRevision: "runtime-catalog-7",
		WorkerType:      slugkit.MustNewForTest("codex"),
		ModelRef:        &model,
		ToolRefs: map[string]resource.Reference{
			"video-generation": workerReference(resource.KindToolBinding, "video-tool"),
		},
		Runtime: resource.WorkerTemplateRuntimeSpec{
			RuntimeImageID: 31, PlacementPolicy: workerspec.PlacementPolicyExplicit,
			ComputeTargetRef: workerReference(resource.KindComputeTarget, "primary-pool"),
			DeploymentMode:   workerspec.DeploymentModeDedicated, ResourceProfileRef: &profile,
		},
		TypeConfig: resource.WorkerTemplateTypeConfigSpec{
			SchemaVersion: 1, Values: map[string]any{"approval-mode": "never"},
			SecretRefs: map[string]resource.Reference{
				"api-token": workerReference(resource.KindEnvironmentBundle, "secret-bundle"),
			},
			InteractionMode: workerspec.InteractionModeACP,
			AutomationLevel: workerspec.AutomationLevelAutonomous,
		},
		Workspace: resource.WorkerTemplateWorkspaceSpec{
			RepositoryRef: &repository, Branch: "main",
			SkillRefs: []resource.Reference{workerReference(resource.KindSkill, "review-skill")},
			KnowledgeMounts: []resource.WorkerTemplateKnowledgeMount{{
				Ref:  workerReference(resource.KindKnowledgeBase, "engineering-docs"),
				Mode: workerspec.KnowledgeMountReadOnly,
			}},
			EnvironmentBundleRefs: []resource.Reference{
				workerReference(resource.KindEnvironmentBundle, "runtime-environment"),
			},
			ConfigBundleRefs: []resource.Reference{
				workerReference(resource.KindEnvironmentBundle, "config-environment"),
			},
			Instructions: "Review before editing.",
		},
		Lifecycle: resource.WorkerTemplateLifecycleSpec{
			TerminationPolicy: workerspec.TerminationPolicyOnIdle, IdleTimeoutMinutes: 30,
		},
		Metadata: resource.WorkerTemplateMetadataSpec{Alias: "Reviewer"},
	}
}

func workerReference(kind, name string) resource.Reference {
	return resource.Reference{Kind: kind, Name: slugkit.MustNewForTest(name)}
}

func workerTemplateScope() control.Scope {
	return control.Scope{
		OrganizationID: 42, OrganizationSlug: slugkit.MustNewForTest("team-alpha"),
		ActorID: 7,
	}
}

var _ = workercreation.Draft{}
