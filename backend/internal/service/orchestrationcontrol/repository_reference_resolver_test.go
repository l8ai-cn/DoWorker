package orchestrationcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryReferenceResolverPinsLatestRevision(t *testing.T) {
	head, revision := referenceResolverResource()
	repository := &orchestrationRepositoryStub{
		resourceSequence: []resourceRead{{head: head}},
		revisions:        map[int64]control.ResourceRevision{revision.Revision: revision},
	}
	authorizer := &orchestrationAuthorizerStub{}
	resolver, err := NewRepositoryReferenceResolver(repository, authorizer)
	require.NoError(t, err)

	resolved, err := resolver.Resolve(
		context.Background(),
		orchestrationServiceScope(),
		DraftReference{
			Path: "/spec/modelRef",
			Reference: orchestrationresource.Reference{
				Kind: "ModelBinding", Name: head.Identity.Name,
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, head.Identity.TypeMeta, resolved.TypeMeta)
	assert.Equal(t, head.Identity.Namespace, resolved.Namespace)
	assert.Equal(t, head.Identity.Name, resolved.Name)
	assert.Equal(t, head.Identity.UID, resolved.UID)
	assert.Equal(t, revision.Revision, resolved.Revision)
	assert.Equal(t, revision.Digest, resolved.Digest)
	assert.Equal(t, 1, authorizer.referenceCalls)
}

func TestRepositoryReferenceResolverPinsExplicitRevision(t *testing.T) {
	head, _ := referenceResolverResource()
	historical := referenceResolverRevision(head, 2, 2, 2)
	repository := &orchestrationRepositoryStub{
		resourceSequence: []resourceRead{{head: head}},
		revisions:        map[int64]control.ResourceRevision{2: historical},
	}
	resolver, err := NewRepositoryReferenceResolver(
		repository,
		&orchestrationAuthorizerStub{},
	)
	require.NoError(t, err)

	resolved, err := resolver.Resolve(
		context.Background(),
		orchestrationServiceScope(),
		DraftReference{
			Path: "/spec/modelRef",
			Reference: orchestrationresource.Reference{
				Kind: "ModelBinding", Name: head.Identity.Name, Revision: 2,
			},
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(2), resolved.Revision)
	assert.Equal(t, historical.Digest, resolved.Digest)
}

func TestRepositoryReferenceResolverEnforcesReferenceAuthorization(t *testing.T) {
	head, _ := referenceResolverResource()
	repository := &orchestrationRepositoryStub{
		resourceSequence: []resourceRead{{head: head}},
		revisions:        map[int64]control.ResourceRevision{},
	}
	authorizer := &orchestrationAuthorizerStub{referenceErr: ErrForbidden}
	resolver, err := NewRepositoryReferenceResolver(repository, authorizer)
	require.NoError(t, err)

	_, err = resolver.Resolve(
		context.Background(),
		orchestrationServiceScope(),
		DraftReference{
			Path: "/spec/modelRef",
			Reference: orchestrationresource.Reference{
				Kind: "ModelBinding", Name: head.Identity.Name,
			},
		},
	)

	assert.ErrorIs(t, err, ErrForbidden)
	assert.Equal(t, 1, authorizer.referenceCalls)
}

func TestRepositoryReferenceResolverRejectsSubstitutedRevision(t *testing.T) {
	head, revision := referenceResolverResource()
	revision.Identity.UID = "99999999-9999-4999-8999-999999999999"
	repository := &orchestrationRepositoryStub{
		resourceSequence: []resourceRead{{head: head}},
		revisions:        map[int64]control.ResourceRevision{revision.Revision: revision},
	}
	resolver, err := NewRepositoryReferenceResolver(
		repository,
		&orchestrationAuthorizerStub{},
	)
	require.NoError(t, err)

	_, err = resolver.Resolve(
		context.Background(),
		orchestrationServiceScope(),
		DraftReference{
			Path: "/spec/modelRef",
			Reference: orchestrationresource.Reference{
				Kind: "ModelBinding", Name: head.Identity.Name,
			},
		},
	)

	assert.ErrorIs(t, err, control.ErrCorrupt)
}

func referenceResolverResource() (
	control.ResourceHead,
	control.ResourceRevision,
) {
	target := control.ResourceTarget{
		TypeMeta: orchestrationresource.TypeMeta{
			APIVersion: orchestrationresource.APIVersionV1Alpha1,
			Kind:       "ModelBinding",
		},
		Namespace: slugkit.MustNewForTest("team-alpha"),
		Name:      slugkit.MustNewForTest("coding-primary"),
	}
	identity := control.ResourceIdentity{
		ResourceTarget: target,
		UID:            "33333333-3333-4333-8333-333333333333",
	}
	at := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	head := control.ResourceHead{
		ID: 201, OrganizationID: 42, Identity: identity,
		Labels: map[string]string{}, Status: []byte(`{}`),
		Revision: 4, Generation: 4, ResourceVersion: 5,
		CreatedByID: 7, UpdatedByID: 7, CreatedAt: at, UpdatedAt: at,
	}
	revision := referenceResolverRevision(head, 4, 4, 5)
	return head, revision
}

func referenceResolverRevision(
	head control.ResourceHead,
	revision, generation, resourceVersion int64,
) control.ResourceRevision {
	manifest, err := control.CanonicalJSONObject(orchestrationresource.Manifest{
		TypeMeta: head.Identity.TypeMeta,
		Metadata: orchestrationresource.Metadata{
			Name: head.Identity.Name, Namespace: head.Identity.Namespace,
			UID: head.Identity.UID, ResourceVersion: fmt.Sprint(resourceVersion),
			Generation: generation,
		},
		Spec: json.RawMessage(`{"resourceId":101}`),
	})
	if err != nil {
		panic(err)
	}
	spec, err := control.CanonicalJSONObject(json.RawMessage(`{"resourceId":101}`))
	if err != nil {
		panic(err)
	}
	digest, err := control.DigestCanonicalJSON(manifest)
	if err != nil {
		panic(err)
	}
	return control.ResourceRevision{
		OrganizationID: head.OrganizationID, ResourceID: head.ID,
		Identity: head.Identity, Revision: revision, Generation: generation,
		ResourceVersion: resourceVersion, CanonicalManifest: manifest,
		CanonicalSpec: spec, ResolvedReferences: []control.ResolvedReference{},
		Digest: digest, ActorID: head.UpdatedByID, CreatedAt: head.UpdatedAt,
	}
}
