package orchestrationcontrol

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/stretchr/testify/require"
)

func validCreatePlan(t *testing.T) Plan {
	t.Helper()
	manifest, err := CanonicalJSONObject(orchestrationresource.Manifest{
		TypeMeta: validTarget().TypeMeta,
		Metadata: orchestrationresource.Metadata{
			Name:      validTarget().Name,
			Namespace: validTarget().Namespace,
		},
		Spec: json.RawMessage(`{"modelBindingRef":{"kind":"ModelBinding","name":"coding-primary"}}`),
	})
	require.NoError(t, err)
	artifact, err := CanonicalJSONObject(map[string]any{
		"workerSpecVersion": 1,
		"modelResource": map[string]any{
			"uid":      testRefID,
			"revision": 3,
		},
	})
	require.NoError(t, err)
	artifactDigest, err := DigestCanonicalJSON(artifact)
	require.NoError(t, err)
	draftHash, err := DigestCanonicalJSON(manifest)
	require.NoError(t, err)

	plan := Plan{
		ID:                testPlanID,
		Scope:             validScope(),
		ActorID:           7,
		Operation:         PlanOperationCreate,
		Target:            validTarget(),
		DraftHash:         draftHash,
		CanonicalManifest: manifest,
		ResolvedReferences: []ResolvedReference{
			validResolvedReferenceForControl(),
		},
		SemanticChanges: []SemanticChange{{
			Operation: SemanticChangeAdd,
			Path:      "/spec/modelBindingRef",
			After: ChangeValue{
				Digest: "sha256:" + strings.Repeat("e", 64),
			},
		}},
		Issues: []PlanIssue{{
			Severity: PlanIssueWarning,
			Path:     "/spec/modelBindingRef",
			Code:     "model-binding-pinned",
			Message:  "The model binding is pinned to an immutable revision.",
		}},
		ArtifactKind:    "WorkerSpec",
		ArtifactJSON:    artifact,
		ArtifactDigest:  artifactDigest,
		OptionsRevision: "runtime-catalog-4",
		CreatedAt:       testCreatedAt,
		ExpiresAt:       testCreatedAt.Add(5 * time.Minute),
		Status:          PlanStatusPending,
	}
	plan.PlanHash, err = ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	return plan
}

func validUpdatePlan(t *testing.T) Plan {
	t.Helper()
	plan := validCreatePlan(t)
	plan.Operation = PlanOperationUpdate
	plan.TargetResourceID = 101
	plan.BaseUID = testTargetID
	plan.BaseResourceVersion = 8
	var err error
	plan.PlanHash, err = ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	return plan
}

func TestPlanValidatesCreateAndUpdateContracts(t *testing.T) {
	require.NoError(t, validCreatePlan(t).Validate())
	require.NoError(t, validUpdatePlan(t).Validate())

	t.Run("create rejects base state", func(t *testing.T) {
		plan := validCreatePlan(t)
		plan.BaseUID = testTargetID
		plan.BaseResourceVersion = 1
		require.ErrorIs(t, plan.Validate(), ErrInvalid)
	})

	t.Run("update requires base state", func(t *testing.T) {
		plan := validUpdatePlan(t)
		plan.BaseUID = ""
		require.ErrorIs(t, plan.Validate(), ErrInvalid)
	})

	t.Run("actor mismatch", func(t *testing.T) {
		plan := validCreatePlan(t)
		plan.ActorID = 8
		require.ErrorIs(t, plan.Validate(), ErrInvalid)
	})

	t.Run("tenant comes from scope", func(t *testing.T) {
		plan := validCreatePlan(t)
		plan.Scope.OrganizationSlug = "team-beta"
		require.ErrorIs(t, plan.Validate(), ErrInvalid)
	})

	t.Run("plan hash mismatch is corrupt", func(t *testing.T) {
		plan := validCreatePlan(t)
		plan.PlanHash = "sha256:" + strings.Repeat("f", 64)
		require.ErrorIs(t, plan.Validate(), ErrCorrupt)
	})
}

func TestPlanValueEnumsPathsAndShapes(t *testing.T) {
	plan := validCreatePlan(t)

	tests := []struct {
		name   string
		mutate func(*Plan)
	}{
		{"operation", func(value *Plan) { value.Operation = "replace" }},
		{"status", func(value *Plan) { value.Status = "running" }},
		{"issue severity", func(value *Plan) { value.Issues[0].Severity = "info" }},
		{"issue path", func(value *Plan) { value.Issues[0].Path = "spec.model" }},
		{"issue path escape", func(value *Plan) { value.Issues[0].Path = "/spec/~2model" }},
		{"issue code", func(value *Plan) { value.Issues[0].Code = "Bad Code" }},
		{"change operation", func(value *Plan) { value.SemanticChanges[0].Operation = "move" }},
		{"change path", func(value *Plan) { value.SemanticChanges[0].Path = "spec/model" }},
		{"add with before", func(value *Plan) {
			value.SemanticChanges[0].Before.Digest = "sha256:" + strings.Repeat("a", 64)
		}},
		{"artifact kind", func(value *Plan) { value.ArtifactKind = "worker-spec" }},
		{"options revision", func(value *Plan) { value.OptionsRevision = "" }},
		{"expiry ordering", func(value *Plan) { value.ExpiresAt = value.CreatedAt }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := plan
			value.Issues = append([]PlanIssue(nil), plan.Issues...)
			value.SemanticChanges = append([]SemanticChange(nil), plan.SemanticChanges...)
			test.mutate(&value)
			require.ErrorIs(t, value.Validate(), ErrInvalid)
		})
	}
}

func TestPlanRejectsRawSecretsWithoutEchoingThem(t *testing.T) {
	const secret = "sk-live-do-not-echo"
	tests := []struct {
		name   string
		mutate func(*Plan)
	}{
		{"manifest", func(plan *Plan) {
			plan.CanonicalManifest = json.RawMessage(`{
				"apiVersion":"agentsmesh.io/v1alpha1",
				"kind":"WorkerTemplate",
				"metadata":{"name":"worker-one","namespace":"team-alpha"},
				"spec":{"apiToken":"` + secret + `"}
			}`)
		}},
		{"artifact", func(plan *Plan) {
			plan.ArtifactJSON = json.RawMessage(`{"apiToken":"` + secret + `"}`)
		}},
		{"issue", func(plan *Plan) {
			plan.Issues[0].Message = "token=" + secret
		}},
		{"diff", func(plan *Plan) {
			plan.SemanticChanges[0].After = ChangeValue{
				RedactedJSON: json.RawMessage(`{"token":"` + secret + `"}`),
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := validCreatePlan(t)
			test.mutate(&plan)

			err := plan.Validate()
			require.Error(t, err)
			require.NotContains(t, err.Error(), secret)
		})
	}
}

func TestPlanRejectsSecretsNestedUnderSensitiveKeys(t *testing.T) {
	plan := validCreatePlan(t)
	plan.ArtifactJSON = json.RawMessage(
		`{"password":{"value":"plaintext-value"}}`,
	)

	err := plan.Validate()
	require.ErrorIs(t, err, ErrInvalid)
	require.NotContains(t, err.Error(), "plaintext-value")
}
