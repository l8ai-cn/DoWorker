package orchestrationcontrol

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

const (
	testPlanID   = "11111111-1111-4111-8111-111111111111"
	testTargetID = "22222222-2222-4222-8222-222222222222"
	testRefID    = "33333333-3333-4333-8333-333333333333"
)

var testCreatedAt = time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)

func validScope() Scope {
	return Scope{
		OrganizationID:   42,
		OrganizationSlug: slugkit.MustNewForTest("team-alpha"),
		ActorID:          7,
	}
}

func validTarget() ResourceTarget {
	return ResourceTarget{
		TypeMeta: orchestrationresource.TypeMeta{
			APIVersion: orchestrationresource.APIVersionV1Alpha1,
			Kind:       "WorkerTemplate",
		},
		Namespace: slugkit.MustNewForTest("team-alpha"),
		Name:      slugkit.MustNewForTest("worker-one"),
	}
}

func validIdentity() ResourceIdentity {
	return ResourceIdentity{ResourceTarget: validTarget(), UID: testTargetID}
}

func validResolvedReferenceForControl() ResolvedReference {
	return ResolvedReference{
		TypeMeta: orchestrationresource.TypeMeta{
			APIVersion: orchestrationresource.APIVersionV1Alpha1,
			Kind:       "ModelBinding",
		},
		Namespace: slugkit.MustNewForTest("team-alpha"),
		Name:      slugkit.MustNewForTest("coding-primary"),
		UID:       testRefID,
		Revision:  3,
		Digest:    "sha256:" + strings.Repeat("a", 64),
	}
}

func canonicalStoredManifest(t *testing.T) json.RawMessage {
	t.Helper()
	raw, err := CanonicalJSONObject(orchestrationresource.Manifest{
		TypeMeta: validTarget().TypeMeta,
		Metadata: orchestrationresource.Metadata{
			Name:            validTarget().Name,
			Namespace:       validTarget().Namespace,
			UID:             testTargetID,
			ResourceVersion: "3",
			Generation:      2,
		},
		Spec:   json.RawMessage(`{"model":"coding-primary"}`),
		Status: json.RawMessage(`{"ready":true}`),
	})
	require.NoError(t, err)
	return raw
}

func TestScopeAndTargetUseAuthenticatedTenant(t *testing.T) {
	require.NoError(t, validScope().Validate())
	require.NoError(t, validTarget().Validate(validScope()))

	tests := []struct {
		name   string
		mutate func(*Scope, *ResourceTarget)
	}{
		{"organization id", func(scope *Scope, _ *ResourceTarget) {
			scope.OrganizationID = 0
		}},
		{"actor id", func(scope *Scope, _ *ResourceTarget) {
			scope.ActorID = -1
		}},
		{"organization slug", func(scope *Scope, _ *ResourceTarget) {
			scope.OrganizationSlug = "Team_Alpha"
		}},
		{"manifest namespace", func(_ *Scope, target *ResourceTarget) {
			target.Namespace = slugkit.MustNewForTest("team-beta")
		}},
		{"type meta", func(_ *Scope, target *ResourceTarget) {
			target.Kind = "worker-template"
		}},
		{"name", func(_ *Scope, target *ResourceTarget) {
			target.Name = "Worker_One"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := validScope()
			target := validTarget()
			test.mutate(&scope, &target)

			err := target.Validate(scope)
			require.ErrorIs(t, err, ErrInvalid)
		})
	}
}

func TestResourceIdentityRequiresCanonicalUUID(t *testing.T) {
	require.NoError(t, validIdentity().Validate(validScope()))

	for _, uid := range []string{
		"",
		"not-a-uuid",
		"22222222222242228222222222222222",
		"22222222-2222-4222-8222-22222222222A",
	} {
		identity := validIdentity()
		identity.UID = uid

		err := identity.Validate(validScope())
		require.ErrorIs(t, err, ErrInvalid)
		if uid != "" {
			require.NotContains(t, err.Error(), uid)
		}
	}
}

func TestResourceHeadValidatesStateConstraints(t *testing.T) {
	head := ResourceHead{
		ID:              101,
		OrganizationID:  42,
		Identity:        validIdentity(),
		DisplayName:     "Worker One",
		Labels:          map[string]string{"role": "builder"},
		Status:          json.RawMessage(`{"ready":true}`),
		Revision:        2,
		Generation:      2,
		ResourceVersion: 3,
		CreatedByID:     7,
		UpdatedByID:     8,
		CreatedAt:       testCreatedAt,
		UpdatedAt:       testCreatedAt.Add(time.Minute),
	}
	require.NoError(t, head.Validate(validScope()))

	tests := []struct {
		name   string
		want   error
		mutate func(*ResourceHead)
	}{
		{"missing id", ErrInvalid, func(value *ResourceHead) { value.ID = 0 }},
		{"organization mismatch", ErrInvalid, func(value *ResourceHead) { value.OrganizationID = 99 }},
		{"invalid display name", ErrInvalid, func(value *ResourceHead) { value.DisplayName = "bad\nname" }},
		{"invalid label", ErrInvalid, func(value *ResourceHead) {
			value.Labels = map[string]string{"Bad_Key": "builder"}
		}},
		{"non-canonical status", ErrCorrupt, func(value *ResourceHead) {
			value.Status = json.RawMessage(`{ "ready": true }`)
		}},
		{"secret status", ErrCorrupt, func(value *ResourceHead) {
			value.Status = json.RawMessage(`{"apiToken":"sk-do-not-store"}`)
		}},
		{"zero revision", ErrInvalid, func(value *ResourceHead) { value.Revision = 0 }},
		{"generation exceeds revision", ErrInvalid, func(value *ResourceHead) { value.Generation = 3 }},
		{"version behind revision", ErrInvalid, func(value *ResourceHead) { value.ResourceVersion = 1 }},
		{"missing creator", ErrInvalid, func(value *ResourceHead) { value.CreatedByID = 0 }},
		{"missing updater", ErrInvalid, func(value *ResourceHead) { value.UpdatedByID = 0 }},
		{"zero created time", ErrInvalid, func(value *ResourceHead) { value.CreatedAt = time.Time{} }},
		{"update before create", ErrInvalid, func(value *ResourceHead) {
			value.UpdatedAt = value.CreatedAt.Add(-time.Second)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := head
			test.mutate(&value)
			require.ErrorIs(t, value.Validate(validScope()), test.want)
		})
	}
}

func TestResourceRevisionInvokesPhaseOneContract(t *testing.T) {
	manifest := canonicalStoredManifest(t)
	spec, err := CanonicalJSONObject(json.RawMessage(`{"model":"coding-primary"}`))
	require.NoError(t, err)
	digest, err := DigestCanonicalJSON(manifest)
	require.NoError(t, err)

	revision := ResourceRevision{
		OrganizationID:    42,
		ResourceID:        101,
		Identity:          validIdentity(),
		Revision:          2,
		Generation:        2,
		ResourceVersion:   3,
		CanonicalManifest: manifest,
		CanonicalSpec:     spec,
		ResolvedReferences: []ResolvedReference{
			validResolvedReferenceForControl(),
		},
		Digest:               digest,
		WorkerSpecSnapshotID: 9,
		ActorID:              7,
		CreatedAt:            testCreatedAt,
	}
	require.NoError(t, revision.Validate(validScope()))

	t.Run("missing resource id", func(t *testing.T) {
		value := revision
		value.ResourceID = 0
		require.ErrorIs(t, value.Validate(validScope()), ErrInvalid)
	})

	t.Run("manifest identity mismatch", func(t *testing.T) {
		value := revision
		var document map[string]any
		require.NoError(t, json.Unmarshal(value.CanonicalManifest, &document))
		document["metadata"].(map[string]any)["namespace"] = "team-beta"
		value.CanonicalManifest, err = CanonicalJSONObject(document)
		require.NoError(t, err)
		value.Digest, err = DigestCanonicalJSON(value.CanonicalManifest)
		require.NoError(t, err)

		require.ErrorIs(t, value.Validate(validScope()), ErrCorrupt)
	})

	t.Run("spec mismatch", func(t *testing.T) {
		value := revision
		value.CanonicalSpec = json.RawMessage(`{"model":"other"}`)
		require.ErrorIs(t, value.Validate(validScope()), ErrCorrupt)
	})

	t.Run("digest mismatch", func(t *testing.T) {
		value := revision
		value.Digest = "sha256:" + strings.Repeat("b", 64)
		require.ErrorIs(t, value.Validate(validScope()), ErrCorrupt)
	})
}
