package orchestrationworker

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdependencyartifact"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var applyTestTime = time.Date(2026, 7, 14, 9, 30, 0, 0, time.UTC)

func TestBuildWorkerTemplateApplyMutationCreatesSnapshotBoundRevision(t *testing.T) {
	registry := workerApplyRegistry(t)
	state := workerApplyCreateState(t)
	_, spec, err := plannedApplyManifest(registry, state)
	require.NoError(t, err)
	_, _, _, _, err = nextApplyState(state, spec)
	require.NoError(t, err)

	mutation, err := buildWorkerTemplateApplyMutation(
		registry,
		state,
		901,
	)

	require.NoError(t, err)
	assert.Equal(t, state.ResultResourceID, mutation.Head.ID)
	assert.Equal(t, state.ResultIdentity, mutation.Head.Identity)
	assert.Equal(t, int64(1), mutation.Head.Revision)
	assert.Equal(t, int64(1), mutation.Head.Generation)
	assert.Equal(t, int64(1), mutation.Head.ResourceVersion)
	assert.Equal(t, int64(901), mutation.Revision.WorkerSpecSnapshotID)
	assert.Equal(t, state.Plan.ArtifactDigest, mutation.ArtifactDigest)
	assert.Equal(t, json.RawMessage(`{}`), mutation.Head.Status)

	var stored resource.Manifest
	require.NoError(t, json.Unmarshal(mutation.Revision.CanonicalManifest, &stored))
	assert.Equal(t, state.ResultIdentity.UID, stored.Metadata.UID)
	assert.Equal(t, "1", stored.Metadata.ResourceVersion)
	assert.Equal(t, int64(1), stored.Metadata.Generation)
	assert.JSONEq(t, `{}`, string(stored.Status))
}

func TestBuildWorkerTemplateApplyMutationPreservesGenerationForMetadataUpdate(
	t *testing.T,
) {
	registry := workerApplyRegistry(t)
	state := workerApplyUpdateState(t, false)

	mutation, err := buildWorkerTemplateApplyMutation(
		registry,
		state,
		902,
	)

	require.NoError(t, err)
	assert.Equal(t, int64(5), mutation.Head.Revision)
	assert.Equal(t, int64(3), mutation.Head.Generation)
	assert.Equal(t, int64(10), mutation.Head.ResourceVersion)
	assert.JSONEq(t, `{"phase":"ready"}`, string(mutation.Head.Status))
	assert.Equal(t, state.Head.CreatedAt, mutation.Head.CreatedAt)
	assert.Equal(t, state.Head.CreatedByID, mutation.Head.CreatedByID)
}

func TestBuildWorkerTemplateApplyMutationIncrementsGenerationForSpecUpdate(
	t *testing.T,
) {
	registry := workerApplyRegistry(t)
	state := workerApplyUpdateState(t, true)

	mutation, err := buildWorkerTemplateApplyMutation(
		registry,
		state,
		903,
	)

	require.NoError(t, err)
	assert.Equal(t, int64(4), mutation.Head.Generation)
	assert.Equal(t, int64(4), mutation.Revision.Generation)
}

func TestBuildWorkerTemplateApplyMutationRejectsWrongTargetAndSnapshot(t *testing.T) {
	registry := workerApplyRegistry(t)
	state := workerApplyCreateState(t)

	_, err := buildWorkerTemplateApplyMutation(registry, state, 0)
	assert.ErrorIs(t, err, control.ErrInvalid)
	state.Plan.Target.Kind = resource.KindModelBinding
	_, err = buildWorkerTemplateApplyMutation(registry, state, 901)
	assert.ErrorIs(t, err, control.ErrInvalid)
}

func TestBuildBindingApplyMutationRejectsSnapshotAndUnknownKind(t *testing.T) {
	registry := workerApplyRegistry(t)
	state := bindingApplyCreateState(t)

	mutation, err := buildBindingApplyMutation(registry, state)

	require.NoError(t, err)
	assert.Zero(t, mutation.Revision.WorkerSpecSnapshotID)
	state.Plan.Target.Kind = resource.KindWorkerTemplate
	_, err = buildBindingApplyMutation(registry, state)
	assert.ErrorIs(t, err, control.ErrInvalid)
}

func TestBuildPromptApplyMutationAcceptsOnlyPromptSpecArtifact(t *testing.T) {
	registry := workerApplyRegistry(t)
	state := promptApplyCreateState(t)

	mutation, err := buildPromptApplyMutation(registry, state)

	require.NoError(t, err)
	assert.Zero(t, mutation.Revision.WorkerSpecSnapshotID)
	assert.Equal(t, resource.KindPrompt, mutation.Head.Identity.Kind)

	state.Plan.ArtifactKind = resource.KindPrompt + "Apply"
	_, err = buildPromptApplyMutation(registry, state)
	assert.ErrorIs(t, err, control.ErrInvalid)

	state = promptApplyCreateState(t)
	state.Plan.Target.Kind = resource.KindExpert
	_, err = buildPromptApplyMutation(registry, state)
	assert.ErrorIs(t, err, control.ErrInvalid)
}

func workerApplyRegistry(t *testing.T) *resource.Registry {
	t.Helper()
	registry := resource.NewRegistry()
	require.NoError(t, resource.RegisterWorkerSchemas(registry))
	require.NoError(t, resource.RegisterDefinitionSchemas(registry))
	return registry
}

func workerApplyCreateState(t *testing.T) controlservice.LockedApplyState {
	t.Helper()
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindWorkerTemplate,
		},
		Metadata: resource.Metadata{
			Name:        "review-worker",
			Namespace:   "team-alpha",
			DisplayName: "Review Worker",
			Labels:      map[string]string{"team": "platform"},
		},
	}
	spec := workerTemplateSpecForTest()
	manifest.Spec = canonicalApplyJSON(t, spec)
	canonicalManifest := canonicalApplyJSON(t, manifest)
	artifact := canonicalApplyJSON(t, map[string]any{"version": 1})
	return controlservice.LockedApplyState{
		Plan: control.Plan{
			Scope: workerTemplateScope(), ActorID: 7,
			Operation: control.PlanOperationCreate,
			Target: control.ResourceTarget{
				TypeMeta:  manifest.TypeMeta,
				Namespace: manifest.Metadata.Namespace,
				Name:      manifest.Metadata.Name,
			},
			CanonicalManifest:  canonicalManifest,
			ResolvedReferences: []control.ResolvedReference{},
			ArtifactKind:       workerdependencyartifact.PlanArtifactKind,
			ArtifactJSON:       artifact,
			ArtifactDigest:     digestApplyJSON(t, artifact),
		},
		ResultResourceID: 301,
		ResultIdentity: control.ResourceIdentity{
			ResourceTarget: control.ResourceTarget{
				TypeMeta:  manifest.TypeMeta,
				Namespace: manifest.Metadata.Namespace,
				Name:      manifest.Metadata.Name,
			},
			UID: "11111111-1111-4111-8111-111111111111",
		},
		AppliedAt: applyTestTime,
	}
}

func workerApplyUpdateState(
	t *testing.T,
	changeSpec bool,
) controlservice.LockedApplyState {
	t.Helper()
	state := workerApplyCreateState(t)
	state.Plan.Operation = control.PlanOperationUpdate
	state.Plan.TargetResourceID = 301
	state.Plan.BaseUID = state.ResultIdentity.UID
	state.Plan.BaseResourceVersion = 9
	state.Head = &control.ResourceHead{
		ID: 301, OrganizationID: 42, Identity: state.ResultIdentity,
		DisplayName: "Old name", Labels: map[string]string{"old": "label"},
		Status:   json.RawMessage(`{"phase":"ready"}`),
		Revision: 4, Generation: 3, ResourceVersion: 9,
		CreatedByID: 5, UpdatedByID: 6,
		CreatedAt: applyTestTime.Add(-time.Hour),
		UpdatedAt: applyTestTime.Add(-time.Minute),
	}
	var manifest resource.Manifest
	require.NoError(t, json.Unmarshal(state.Plan.CanonicalManifest, &manifest))
	currentSpec := manifest.Spec
	if changeSpec {
		spec := workerTemplateSpecForTest()
		spec.Metadata.Alias = "Changed alias"
		manifest.Spec = canonicalApplyJSON(t, spec)
		state.Plan.CanonicalManifest = canonicalApplyJSON(t, manifest)
	}
	state.CurrentRevision = &control.ResourceRevision{
		OrganizationID: 42, ResourceID: 301, Identity: state.ResultIdentity,
		Revision: 4, Generation: 3, ResourceVersion: 9,
		CanonicalSpec: currentSpec,
	}
	return state
}

func bindingApplyCreateState(t *testing.T) controlservice.LockedApplyState {
	t.Helper()
	state := workerApplyCreateState(t)
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindModelBinding,
		},
		Metadata: resource.Metadata{
			Name: "coding-primary", Namespace: "team-alpha",
			DisplayName: "Coding primary", Labels: map[string]string{},
		},
		Spec: canonicalApplyJSON(t, resource.ModelBindingSpec{ResourceID: 101}),
	}
	state.Plan.Target = control.ResourceTarget{
		TypeMeta:  manifest.TypeMeta,
		Namespace: manifest.Metadata.Namespace,
		Name:      manifest.Metadata.Name,
	}
	state.Plan.CanonicalManifest = canonicalApplyJSON(t, manifest)
	state.Plan.ArtifactKind = resource.KindModelBinding + "Spec"
	state.ResultIdentity.ResourceTarget = state.Plan.Target
	return state
}

func promptApplyCreateState(t *testing.T) controlservice.LockedApplyState {
	t.Helper()
	state := workerApplyCreateState(t)
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindPrompt,
		},
		Metadata: resource.Metadata{
			Name: "review-task", Namespace: "team-alpha",
			DisplayName: "Review task", Labels: map[string]string{},
		},
		Spec: canonicalApplyJSON(t, resource.PromptSpec{
			Content: "Review {{change}}",
			Variables: map[string]resource.PromptVariableSpec{
				"change": {Required: true},
			},
		}),
	}
	state.Plan.Target = control.ResourceTarget{
		TypeMeta:  manifest.TypeMeta,
		Namespace: manifest.Metadata.Namespace,
		Name:      manifest.Metadata.Name,
	}
	state.Plan.CanonicalManifest = canonicalApplyJSON(t, manifest)
	state.Plan.ArtifactKind = "PromptSpec"
	state.ResultIdentity.ResourceTarget = state.Plan.Target
	return state
}

func expertApplyCreateState(t *testing.T) controlservice.LockedApplyState {
	t.Helper()
	state := workerApplyCreateState(t)
	workerRef := resource.Reference{
		Kind: resource.KindWorkerTemplate,
		Name: "review-worker",
	}
	promptRef := resource.Reference{
		Kind: resource.KindPrompt,
		Name: "review-system",
	}
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindExpert,
		},
		Metadata: resource.Metadata{
			Name: "review-expert", Namespace: "team-alpha",
			DisplayName: "Review Expert", Labels: map[string]string{},
		},
		Spec: canonicalApplyJSON(t, resource.ExpertResourceSpec{
			WorkerTemplateRef: workerRef,
			PromptRef:         &promptRef,
			Description:       "Reviews code changes",
			Category:          "engineering",
			ReleaseNotes:      "Initial revision",
		}),
	}
	state.Plan.Target = control.ResourceTarget{
		TypeMeta:  manifest.TypeMeta,
		Namespace: manifest.Metadata.Namespace,
		Name:      manifest.Metadata.Name,
	}
	state.Plan.CanonicalManifest = canonicalApplyJSON(t, manifest)
	state.Plan.ArtifactKind = resource.KindExpert + "Apply"
	state.Plan.ArtifactJSON = canonicalApplyJSON(t, DefinitionApplyArtifact{
		WorkerSpecSnapshotID: 901,
	})
	state.Plan.ArtifactDigest = digestApplyJSON(t, state.Plan.ArtifactJSON)
	state.Plan.ResolvedReferences = []control.ResolvedReference{
		resolvedApplyReference(state.Plan.Scope, workerRef, 2, "a"),
		resolvedApplyReference(state.Plan.Scope, promptRef, 3, "b"),
	}
	state.ResultIdentity.ResourceTarget = state.Plan.Target
	return state
}

func resolvedApplyReference(
	scope control.Scope,
	reference resource.Reference,
	revision int64,
	digestDigit string,
) control.ResolvedReference {
	return control.ResolvedReference{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       reference.Kind,
		},
		Namespace: scope.OrganizationSlug,
		Name:      reference.Name,
		UID:       "22222222-2222-4222-8222-222222222222",
		Revision:  revision,
		Digest:    "sha256:" + strings.Repeat(digestDigit, 64),
	}
}

func canonicalApplyJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	canonical, err := control.CanonicalJSONObject(value)
	require.NoError(t, err)
	return canonical
}

func digestApplyJSON(t *testing.T, value json.RawMessage) string {
	t.Helper()
	digest, err := control.DigestCanonicalJSON(value)
	require.NoError(t, err)
	return digest
}
