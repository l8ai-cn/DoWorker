package workerdependencyartifact

import (
	"encoding/json"
	"testing"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanArtifactRoundTripPreservesBothImmutableDocuments(t *testing.T) {
	artifact, err := Build(validInput(t))
	require.NoError(t, err)

	decoded, err := DecodePlanArtifact(artifact.PlanJSON())
	require.NoError(t, err)
	assert.Equal(t, artifact.WorkerSpecJSON(), decoded.WorkerSpecJSON)
	assert.Equal(t, artifact.JSON(), decoded.ResolvedDependenciesJSON)
	assert.Equal(t, artifact.Digest(), decoded.ResolvedDependenciesDigest)

	expectedDigest, err := control.DigestCanonicalJSON(artifact.PlanJSON())
	require.NoError(t, err)
	assert.Equal(t, expectedDigest, artifact.PlanDigest())

	first := artifact.PlanJSON()
	first[0] = 'x'
	assert.NotEqual(t, first, artifact.PlanJSON())
}

func TestPlanArtifactPassesControlPlaneSecretAndDigestValidation(t *testing.T) {
	input := validInput(t)
	addWorkspaceDependencies(t, &input)
	artifact, err := Build(input)
	require.NoError(t, err)
	plan := controlPlanForArtifact(t, input, artifact)

	require.NoError(t, plan.Validate())
}

func TestDecodePlanArtifactRejectsEnvelopeAndDigestTampering(t *testing.T) {
	artifact, err := Build(validInput(t))
	require.NoError(t, err)
	tests := []struct {
		name   string
		mutate func(map[string]json.RawMessage)
		match  string
	}{
		{
			name: "unknown field",
			mutate: func(root map[string]json.RawMessage) {
				root["fallbackWorkerSpec"] = json.RawMessage(`{}`)
			},
			match: "unknown field",
		},
		{
			name: "unsupported version",
			mutate: func(root map[string]json.RawMessage) {
				root["version"] = json.RawMessage(`2`)
			},
			match: "version 2 is unsupported",
		},
		{
			name: "dependency digest",
			mutate: func(root map[string]json.RawMessage) {
				root["resolvedDependenciesDigest"] = json.RawMessage(
					`"` + workerdependency.TextDigest("substituted") + `"`,
				)
			},
			match: "digest binding is invalid",
		},
		{
			name: "WorkerSpec substitution",
			mutate: func(root map[string]json.RawMessage) {
				other := validInput(t)
				other.WorkerSpec.Metadata.Alias = "other-worker"
				otherArtifact, buildErr := Build(other)
				require.NoError(t, buildErr)
				var otherRoot map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(otherArtifact.PlanJSON(), &otherRoot))
				root["workerSpec"] = otherRoot["workerSpec"]
			},
			match: "digest binding is invalid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tampered := mutatePlanArtifact(t, artifact.PlanJSON(), test.mutate)

			_, err := DecodePlanArtifact(tampered)

			require.ErrorContains(t, err, test.match)
		})
	}
}

func TestDecodePlanArtifactRejectsNonCanonicalOrTrailingJSON(t *testing.T) {
	artifact, err := Build(validInput(t))
	require.NoError(t, err)

	_, err = DecodePlanArtifact(append([]byte(" "), artifact.PlanJSON()...))
	require.ErrorContains(t, err, "must be canonical JSON")

	_, err = DecodePlanArtifact(append(artifact.PlanJSON(), []byte(`{}`)...))
	require.ErrorContains(t, err, "must be canonical JSON")
}

func TestDecodePlanArtifactRevalidatesWorkerSpecConsistency(t *testing.T) {
	artifact, err := Build(validInput(t))
	require.NoError(t, err)
	document, err := workerdependency.Decode(artifact.JSON())
	require.NoError(t, err)
	document.Worker.WorkerType = slugkit.MustNewForTest("cursor-cli")
	dependencies, digest, err := workerdependency.EncodeAndDigest(document)
	require.NoError(t, err)
	tampered := mutatePlanArtifact(
		t,
		artifact.PlanJSON(),
		func(root map[string]json.RawMessage) {
			root["resolvedDependencies"] = dependencies
			root["resolvedDependenciesDigest"] = json.RawMessage(`"` + digest + `"`)
		},
	)

	_, err = DecodePlanArtifact(tampered)

	require.ErrorContains(t, err, "worker identity does not match WorkerSpec")
}

func TestWorkerSpecChangeProducesANewBoundArtifact(t *testing.T) {
	firstInput := validInput(t)
	secondInput := validInput(t)
	secondInput.WorkerSpec.Metadata.Alias = "renamed-worker"

	first, err := Build(firstInput)
	require.NoError(t, err)
	second, err := Build(secondInput)
	require.NoError(t, err)

	assert.NotEqual(t, first.WorkerSpecDigest(), second.WorkerSpecDigest())
	assert.NotEqual(t, first.Digest(), second.Digest())
	assert.NotEqual(t, first.PlanDigest(), second.PlanDigest())
}

func mutatePlanArtifact(
	t *testing.T,
	source []byte,
	mutate func(map[string]json.RawMessage),
) []byte {
	t.Helper()
	var root map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(source, &root))
	mutate(root)
	canonical, err := control.CanonicalJSONObject(root)
	require.NoError(t, err)
	return canonical
}

func controlPlanForArtifact(
	t *testing.T,
	input Input,
	artifact Artifact,
) control.Plan {
	t.Helper()
	target := control.ResourceTarget{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindWorkerTemplate,
		},
		Namespace: input.Scope.OrganizationSlug,
		Name:      slugkit.MustNewForTest("video-worker"),
	}
	manifest, err := control.CanonicalJSONObject(resource.Manifest{
		TypeMeta: target.TypeMeta,
		Metadata: resource.Metadata{
			Name: target.Name, Namespace: target.Namespace,
		},
		Spec: json.RawMessage(`{"workerType":"codex-cli"}`),
	})
	require.NoError(t, err)
	draftDigest, err := control.DigestCanonicalJSON(manifest)
	require.NoError(t, err)
	now := time.Now().UTC()
	plan := control.Plan{
		ID: uuid.NewString(), Scope: input.Scope, ActorID: input.Scope.ActorID,
		Operation: control.PlanOperationCreate, Target: target,
		DraftHash: draftDigest, CanonicalManifest: manifest,
		ResolvedReferences: input.PlanReferences,
		SemanticChanges:    []control.SemanticChange{},
		Issues:             []control.PlanIssue{},
		ArtifactKind:       PlanArtifactKind,
		ArtifactJSON:       artifact.PlanJSON(),
		ArtifactDigest:     artifact.PlanDigest(),
		OptionsRevision:    "catalog-v1",
		CreatedAt:          now,
		ExpiresAt:          now.Add(5 * time.Minute),
		Status:             control.PlanStatusPending,
	}
	plan.PlanHash, err = control.ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	return plan
}
