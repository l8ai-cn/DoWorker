package orchestrationworker

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefinitionPlannerReferencesEveryDeclaredDependency(t *testing.T) {
	planner, err := NewDefinitionPlanner(
		resource.KindExpert,
		&definitionResolverStub{},
		&goalLoopSlugCheckerStub{},
	)
	require.NoError(t, err)

	references, err := planner.References(&resource.ExpertResourceSpec{
		WorkerTemplateRef: reference(resource.KindWorkerTemplate, "reviewer"),
		PromptRef:         pointer(reference(resource.KindPrompt, "review-system")),
	})
	require.NoError(t, err)
	require.Len(t, references, 2)
	assert.Equal(t, "/spec/workerTemplateRef", references[0].Path)
	assert.Equal(t, "/spec/promptRef", references[1].Path)
}

func TestDefinitionPlannerPinsWorkerTemplateSnapshot(t *testing.T) {
	resolver := &definitionResolverStub{snapshotID: 91}
	planner, err := NewDefinitionPlanner(
		resource.KindGoalLoop,
		resolver,
		&goalLoopSlugCheckerStub{},
	)
	require.NoError(t, err)
	scope := definitionScope()
	workerRef := reference(resource.KindWorkerTemplate, "reviewer")

	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope: scope,
		Manifest: resource.Manifest{
			TypeMeta: planner.TypeMeta(),
		},
		TypedSpec: &resource.GoalLoopResourceSpec{
			WorkerTemplateRef:   workerRef,
			Objective:           "Fix checkout",
			AcceptanceCriteria:  []string{"Tests pass"},
			VerificationCommand: "go test ./...",
			MaxIterations:       10,
			TimeoutMinutes:      60,
			NoProgressLimit:     3,
			SameErrorLimit:      2,
			EscalationPolicy:    "pause",
		},
		ResolvedReferences: []control.ResolvedReference{
			resolvedReference(scope, workerRef, 4),
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "GoalLoopApply", output.ArtifactKind)
	assert.Equal(t, DefinitionSchemaRevision, output.OptionsRevision)
	assert.Empty(t, output.Issues)

	var artifact DefinitionApplyArtifact
	require.NoError(t, json.Unmarshal(output.ArtifactJSON, &artifact))
	assert.Equal(t, int64(91), artifact.WorkerSpecSnapshotID)
	assert.Equal(t, int64(4), resolver.workerReference.Revision)
}

func TestDefinitionPlannerBlocksWorkerUpdate(t *testing.T) {
	planner, err := NewDefinitionPlanner(
		resource.KindWorker,
		&definitionResolverStub{snapshotID: 91},
		&goalLoopSlugCheckerStub{},
	)
	require.NoError(t, err)
	scope := definitionScope()
	workerRef := reference(resource.KindWorkerTemplate, "reviewer")

	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope:     scope,
		Operation: control.PlanOperationUpdate,
		Manifest: resource.Manifest{
			TypeMeta: planner.TypeMeta(),
		},
		TypedSpec: &resource.WorkerInvocationSpec{
			WorkerTemplateRef: workerRef,
			Inputs:            map[string]string{},
		},
		ResolvedReferences: []control.ResolvedReference{
			resolvedReference(scope, workerRef, 4),
		},
	})

	require.NoError(t, err)
	require.Len(t, output.Issues, 1)
	assert.Equal(t, control.PlanIssueBlocking, output.Issues[0].Severity)
	assert.Equal(t, "/", output.Issues[0].Path)
	assert.Equal(t, "worker-is-create-only", output.Issues[0].Code)
}

func TestDefinitionPlannerBlocksGoalLoopUpdate(t *testing.T) {
	planner, err := NewDefinitionPlanner(
		resource.KindGoalLoop,
		&definitionResolverStub{},
		&goalLoopSlugCheckerStub{},
	)
	require.NoError(t, err)

	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope:     definitionScope(),
		Operation: control.PlanOperationUpdate,
		Manifest: resource.Manifest{
			TypeMeta: planner.TypeMeta(),
		},
		TypedSpec: &resource.GoalLoopResourceSpec{},
	})

	require.NoError(t, err)
	require.Len(t, output.Issues, 1)
	assert.Equal(t, control.PlanIssueBlocking, output.Issues[0].Severity)
	assert.Equal(t, "/", output.Issues[0].Path)
	assert.Equal(t, "goal-loop-is-create-only", output.Issues[0].Code)
}

func TestDefinitionPlannerBlocksExistingGoalLoopName(t *testing.T) {
	checker := &goalLoopSlugCheckerStub{exists: true}
	planner, err := NewDefinitionPlanner(
		resource.KindGoalLoop,
		&definitionResolverStub{},
		checker,
	)
	require.NoError(t, err)
	scope := definitionScope()

	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope: scope,
		Manifest: resource.Manifest{
			TypeMeta: planner.TypeMeta(),
			Metadata: resource.Metadata{
				Name: slugkit.MustNewForTest("existing-loop"),
			},
		},
		TypedSpec: &resource.GoalLoopResourceSpec{},
	})

	require.NoError(t, err)
	require.Len(t, output.Issues, 1)
	assert.Equal(t, control.PlanIssueBlocking, output.Issues[0].Severity)
	assert.Equal(t, "/metadata/name", output.Issues[0].Path)
	assert.Equal(t, "goal-loop-name-already-exists", output.Issues[0].Code)
	assert.Equal(t, scope.OrganizationID, checker.organizationID)
	assert.Equal(t, "existing-loop", checker.name)
}

func TestDefinitionPlannerBlocksMissingPromptInputs(t *testing.T) {
	resolver := &definitionResolverStub{
		snapshotID: 91,
		prompt: resource.PromptSpec{
			Content: "Review {{change}}",
			Variables: map[string]resource.PromptVariableSpec{
				"change": {Required: true},
			},
		},
	}
	planner, err := NewDefinitionPlanner(
		resource.KindWorkflow,
		resolver,
		&goalLoopSlugCheckerStub{},
	)
	require.NoError(t, err)
	scope := definitionScope()
	workerRef := reference(resource.KindWorkerTemplate, "reviewer")
	promptRef := reference(resource.KindPrompt, "nightly-review")

	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope:    scope,
		Manifest: resource.Manifest{TypeMeta: planner.TypeMeta()},
		TypedSpec: &resource.WorkflowResourceSpec{
			WorkerTemplateRef: workerRef,
			PromptRef:         promptRef,
			Inputs:            map[string]string{},
		},
		ResolvedReferences: []control.ResolvedReference{
			resolvedReference(scope, workerRef, 2),
			resolvedReference(scope, promptRef, 3),
		},
	})
	require.NoError(t, err)
	require.Len(t, output.Issues, 1)
	assert.Equal(t, control.PlanIssueBlocking, output.Issues[0].Severity)
	assert.Equal(t, "/spec/inputs/change", output.Issues[0].Path)
	assert.Equal(t, "missing-prompt-input", output.Issues[0].Code)
}

func TestDefinitionPlannerBlocksExpertPromptWithoutRequiredDefaults(t *testing.T) {
	resolver := &definitionResolverStub{
		snapshotID: 91,
		prompt: resource.PromptSpec{
			Content: "Review {{change}}",
			Variables: map[string]resource.PromptVariableSpec{
				"change": {Required: true},
			},
		},
	}
	planner, err := NewDefinitionPlanner(
		resource.KindExpert,
		resolver,
		&goalLoopSlugCheckerStub{},
	)
	require.NoError(t, err)
	scope := definitionScope()
	workerRef := reference(resource.KindWorkerTemplate, "reviewer")
	promptRef := reference(resource.KindPrompt, "review-system")

	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope:    scope,
		Manifest: resource.Manifest{TypeMeta: planner.TypeMeta()},
		TypedSpec: &resource.ExpertResourceSpec{
			WorkerTemplateRef: workerRef,
			PromptRef:         &promptRef,
		},
		ResolvedReferences: []control.ResolvedReference{
			resolvedReference(scope, workerRef, 2),
			resolvedReference(scope, promptRef, 3),
		},
	})

	require.NoError(t, err)
	require.Len(t, output.Issues, 1)
	assert.Equal(t, "missing-prompt-input", output.Issues[0].Code)
}

func TestDefinitionPlannerPlansPromptWithoutSnapshot(t *testing.T) {
	planner, err := NewDefinitionPlanner(
		resource.KindPrompt,
		&definitionResolverStub{},
		&goalLoopSlugCheckerStub{},
	)
	require.NoError(t, err)
	spec := &resource.PromptSpec{
		Content:   "Review",
		Variables: map[string]resource.PromptVariableSpec{},
	}
	output, err := planner.Plan(context.Background(), controlservice.TargetPlanInput{
		Scope: definitionScope(),
		Manifest: resource.Manifest{
			TypeMeta: planner.TypeMeta(),
			Spec:     json.RawMessage(`{"content":"Review","variables":{}}`),
		},
		TypedSpec: spec,
	})
	require.NoError(t, err)
	assert.Equal(t, "PromptSpec", output.ArtifactKind)
	assert.JSONEq(t, `{"content":"Review","variables":{}}`, string(output.ArtifactJSON))
}

type definitionResolverStub struct {
	snapshotID      int64
	prompt          resource.PromptSpec
	workerReference control.ResolvedReference
}

func (stub *definitionResolverStub) ResolveWorkerSpecSnapshotID(
	_ context.Context,
	_ control.Scope,
	reference control.ResolvedReference,
) (int64, error) {
	stub.workerReference = reference
	return stub.snapshotID, nil
}

func (stub *definitionResolverStub) ResolvePromptSpec(
	_ context.Context,
	_ control.Scope,
	_ control.ResolvedReference,
) (resource.PromptSpec, error) {
	return stub.prompt, nil
}

type goalLoopSlugCheckerStub struct {
	exists         bool
	organizationID int64
	name           string
}

func (stub *goalLoopSlugCheckerStub) ExistsSlug(
	_ context.Context,
	organizationID int64,
	name string,
) (bool, error) {
	stub.organizationID = organizationID
	stub.name = name
	return stub.exists, nil
}

func definitionScope() control.Scope {
	return control.Scope{
		OrganizationID:   7,
		OrganizationSlug: "acme",
		ActorID:          11,
	}
}

func reference(kind string, name string) resource.Reference {
	return resource.Reference{Kind: kind, Name: slugkit.Slug(name)}
}

func pointer(value resource.Reference) *resource.Reference {
	return &value
}

func resolvedReference(
	scope control.Scope,
	reference resource.Reference,
	revision int64,
) control.ResolvedReference {
	return control.ResolvedReference{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       reference.Kind,
		},
		Namespace: scope.OrganizationSlug,
		Name:      reference.Name,
		UID:       "11111111-1111-4111-8111-111111111111",
		Revision:  revision,
		Digest:    "sha256:" + strings.Repeat("a", 64),
	}
}
