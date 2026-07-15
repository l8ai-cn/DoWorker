package orchestrationcontrol

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizeApplyRechecksUpdatePermission(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	head := orchestrationServiceHead()
	fixture.repository.resourceSequence = []resourceRead{
		{head: head},
		{head: head},
	}
	fixture.repository.revisions[head.Revision] = orchestrationServiceRevision(
		t,
		head,
	)
	service := fixture.service(t)
	planned, err := service.Plan(context.Background(), PlanRequest{
		Scope: fixture.scope,
		Source: ResourceSource{
			Format: SourceFormatJSON, Content: []byte(testResourceJSON),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, planned.Plan)
	fixture.repository.plan = *planned.Plan
	fixture.repository.resourceSequence = []resourceRead{{head: head}}
	fixture.authorizer.updateErr = ErrForbidden
	updateCalls := fixture.authorizer.updateCalls

	err = service.AuthorizeApply(
		context.Background(),
		fixture.scope,
		planned.Plan.ID,
	)

	assert.ErrorIs(t, err, ErrForbidden)
	assert.Equal(t, updateCalls+1, fixture.authorizer.updateCalls)
	assert.Zero(t, fixture.authorizer.referenceCalls)
}

func TestAuthorizeApplyRechecksPinnedReferencePermission(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	service := fixture.service(t)
	planned, err := service.Plan(context.Background(), PlanRequest{
		Scope: fixture.scope,
		Source: ResourceSource{
			Format: SourceFormatJSON, Content: []byte(testResourceJSON),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, planned.Plan)
	fixture.repository.plan = *planned.Plan
	reference := planned.Plan.ResolvedReferences[0]
	referenceHead, referenceRevision := referencedResourceForApply(
		t,
		fixture.scope,
		reference,
	)
	fixture.repository.resourceSequence = []resourceRead{{head: referenceHead}}
	fixture.repository.revisions[reference.Revision] = referenceRevision
	fixture.authorizer.referenceErr = ErrForbidden
	createCalls := fixture.authorizer.createCalls

	err = service.AuthorizeApply(
		context.Background(),
		fixture.scope,
		planned.Plan.ID,
	)

	assert.ErrorIs(t, err, ErrForbidden)
	assert.Equal(t, createCalls+1, fixture.authorizer.createCalls)
	assert.Equal(t, 1, fixture.authorizer.referenceCalls)
}

func TestAuthorizeApplyRejectsPinnedReferenceDigestMismatch(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	service := fixture.service(t)
	planned, err := service.Plan(context.Background(), PlanRequest{
		Scope: fixture.scope,
		Source: ResourceSource{
			Format: SourceFormatJSON, Content: []byte(testResourceJSON),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, planned.Plan)
	reference := planned.Plan.ResolvedReferences[0]
	referenceHead, referenceRevision := referencedResourceForApply(
		t,
		fixture.scope,
		reference,
	)
	fixture.repository.plan = *planned.Plan
	fixture.repository.resourceSequence = []resourceRead{{head: referenceHead}}
	fixture.repository.revisions[reference.Revision] = referenceRevision

	err = service.AuthorizeApply(
		context.Background(),
		fixture.scope,
		planned.Plan.ID,
	)

	assert.ErrorIs(t, err, control.ErrCorrupt)
	assert.Equal(t, 1, fixture.authorizer.referenceCalls)
}

func referencedResourceForApply(
	t *testing.T,
	scope control.Scope,
	reference control.ResolvedReference,
) (control.ResourceHead, control.ResourceRevision) {
	t.Helper()
	head := control.ResourceHead{
		ID:             reference.Revision + 1000,
		OrganizationID: scope.OrganizationID,
		Identity: control.ResourceIdentity{
			ResourceTarget: control.ResourceTarget{
				TypeMeta:  reference.TypeMeta,
				Namespace: reference.Namespace,
				Name:      reference.Name,
			},
			UID: reference.UID,
		},
		Labels: map[string]string{}, Status: []byte(`{}`),
		Revision: reference.Revision, Generation: 1,
		ResourceVersion: reference.Revision,
		CreatedByID:     scope.ActorID, UpdatedByID: scope.ActorID,
		CreatedAt: fixtureTimestamp(), UpdatedAt: fixtureTimestamp(),
	}
	manifest, err := json.Marshal(map[string]any{
		"apiVersion": reference.APIVersion,
		"kind":       reference.Kind,
		"metadata": map[string]any{
			"name":            reference.Name,
			"namespace":       reference.Namespace,
			"uid":             reference.UID,
			"generation":      1,
			"resourceVersion": strconv.FormatInt(reference.Revision, 10),
		},
		"spec": map[string]any{},
	})
	require.NoError(t, err)
	canonicalManifest, err := control.CanonicalJSONObject(manifest)
	require.NoError(t, err)
	digest, err := control.DigestCanonicalJSON(canonicalManifest)
	require.NoError(t, err)
	revision := control.ResourceRevision{
		OrganizationID:     scope.OrganizationID,
		ResourceID:         head.ID,
		Identity:           head.Identity,
		Revision:           reference.Revision,
		Generation:         1,
		ResourceVersion:    reference.Revision,
		CanonicalManifest:  canonicalManifest,
		CanonicalSpec:      []byte(`{}`),
		ResolvedReferences: []control.ResolvedReference{},
		Digest:             digest,
		ActorID:            scope.ActorID,
		CreatedAt:          fixtureTimestamp(),
	}
	return head, revision
}

func fixtureTimestamp() time.Time {
	return time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC)
}
