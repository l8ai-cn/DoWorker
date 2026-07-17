package orchestrationworker

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

type resourceBindingResolverFixture struct {
	scope      control.Scope
	repository *resourceBindingRepositoryStub
	authorizer *resourceBindingAuthorizerStub
	resolver   *ResourceBindingResolver
	nextID     int64
}

func newResourceBindingResolverFixture(
	t *testing.T,
) *resourceBindingResolverFixture {
	t.Helper()
	registry := resource.NewRegistry()
	require.NoError(t, resource.RegisterWorkerSchemas(registry))
	require.NoError(t, resource.RegisterDefinitionSchemas(registry))
	repository := &resourceBindingRepositoryStub{
		heads:     make(map[string]control.ResourceHead),
		revisions: make(map[string]control.ResourceRevision),
	}
	authorizer := &resourceBindingAuthorizerStub{}
	resolver, err := NewResourceBindingResolver(
		registry,
		repository,
		authorizer,
	)
	require.NoError(t, err)
	return &resourceBindingResolverFixture{
		scope:      workerTemplateScope(),
		repository: repository,
		authorizer: authorizer,
		resolver:   resolver,
		nextID:     200,
	}
}

func (fixture *resourceBindingResolverFixture) setSnapshotID(
	pinned control.ResolvedReference,
	snapshotID int64,
) {
	target := control.ResourceTarget{
		TypeMeta:  pinned.TypeMeta,
		Namespace: pinned.Namespace,
		Name:      pinned.Name,
	}
	head := fixture.repository.heads[resourceBindingTargetKey(target)]
	key := resourceBindingRevisionKey(head.ID, pinned.Revision)
	revision := fixture.repository.revisions[key]
	revision.WorkerSpecSnapshotID = snapshotID
	fixture.repository.revisions[key] = revision
}

func (fixture *resourceBindingResolverFixture) addBinding(
	t *testing.T,
	kind string,
	name string,
	spec any,
	references []control.ResolvedReference,
) control.ResolvedReference {
	t.Helper()
	fixture.nextID++
	target := control.ResourceTarget{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
		},
		Namespace: fixture.scope.OrganizationSlug,
		Name:      resourceBindingName(name),
	}
	identity := control.ResourceIdentity{
		ResourceTarget: target,
		UID: fmt.Sprintf(
			"00000000-0000-4000-8000-%012d",
			fixture.nextID,
		),
	}
	at := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	head := control.ResourceHead{
		ID: fixture.nextID, OrganizationID: fixture.scope.OrganizationID,
		Identity: identity, Labels: map[string]string{}, Status: json.RawMessage(`{}`),
		Revision: 2, Generation: 2, ResourceVersion: 2,
		CreatedByID: fixture.scope.ActorID, UpdatedByID: fixture.scope.ActorID,
		CreatedAt: at, UpdatedAt: at,
	}
	specJSON, err := control.CanonicalJSONObject(spec)
	require.NoError(t, err)
	manifestJSON, err := control.CanonicalJSONObject(resource.Manifest{
		TypeMeta: target.TypeMeta,
		Metadata: resource.Metadata{
			Name: target.Name, Namespace: target.Namespace,
			UID: identity.UID, ResourceVersion: "1", Generation: 1,
		},
		Spec: specJSON,
	})
	require.NoError(t, err)
	digest, err := control.DigestCanonicalJSON(manifestJSON)
	require.NoError(t, err)
	revision := control.ResourceRevision{
		OrganizationID: fixture.scope.OrganizationID,
		ResourceID:     head.ID, Identity: identity,
		Revision: 1, Generation: 1, ResourceVersion: 1,
		CanonicalManifest: manifestJSON, CanonicalSpec: specJSON,
		ResolvedReferences: append([]control.ResolvedReference{}, references...),
		Digest:             digest, ActorID: fixture.scope.ActorID, CreatedAt: at,
	}
	fixture.repository.heads[resourceBindingTargetKey(target)] = head
	fixture.repository.revisions[resourceBindingRevisionKey(head.ID, 1)] = revision
	return control.ResolvedReference{
		TypeMeta: target.TypeMeta, Namespace: target.Namespace, Name: target.Name,
		UID: identity.UID, Revision: revision.Revision, Digest: digest,
	}
}

func resolvedBindingReference(
	scope control.Scope,
	kind string,
	name string,
) control.ResolvedReference {
	return control.ResolvedReference{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
		},
		Namespace: scope.OrganizationSlug,
		Name:      resourceBindingName(name),
		UID:       "99999999-9999-4999-8999-999999999999",
		Revision:  1,
		Digest:    "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
}

func resourceBindingName(value string) slugkit.Slug {
	return slugkit.MustNewForTest(value)
}

type resourceBindingRepositoryStub struct {
	heads         map[string]control.ResourceHead
	revisions     map[string]control.ResourceRevision
	revisionCalls int
	lastRevision  int64
}

func (stub *resourceBindingRepositoryStub) GetResource(
	_ context.Context,
	_ control.Scope,
	target control.ResourceTarget,
) (control.ResourceHead, error) {
	head, exists := stub.heads[resourceBindingTargetKey(target)]
	if !exists {
		return control.ResourceHead{}, control.ErrNotFound
	}
	return head, nil
}

func (*resourceBindingRepositoryStub) ListResources(
	context.Context,
	control.Scope,
	controlservice.ResourceListFilter,
) (controlservice.ResourceListPage, error) {
	return controlservice.ResourceListPage{}, nil
}

func (stub *resourceBindingRepositoryStub) GetRevision(
	_ context.Context,
	_ control.Scope,
	resourceID int64,
	revision int64,
) (control.ResourceRevision, error) {
	stub.revisionCalls++
	stub.lastRevision = revision
	value, exists := stub.revisions[resourceBindingRevisionKey(resourceID, revision)]
	if !exists {
		return control.ResourceRevision{}, control.ErrNotFound
	}
	return value, nil
}

func (*resourceBindingRepositoryStub) ListRevisions(
	context.Context,
	control.Scope,
	int64,
	int,
	int,
) ([]control.ResourceRevision, error) {
	return nil, nil
}

func (*resourceBindingRepositoryStub) CreatePlan(
	context.Context,
	control.Plan,
) error {
	return nil
}

func (*resourceBindingRepositoryStub) GetPlan(
	context.Context,
	control.Scope,
	string,
) (control.Plan, error) {
	return control.Plan{}, control.ErrNotFound
}

func (*resourceBindingRepositoryStub) RunApplyTransaction(
	context.Context,
	control.Scope,
	string,
	controlservice.ApplyBuilder,
) (control.ResourceHead, error) {
	return control.ResourceHead{}, nil
}

type resourceBindingAuthorizerStub struct {
	err            error
	referenceCalls int
}

func (*resourceBindingAuthorizerStub) AuthorizeList(
	context.Context,
	control.Scope,
) error {
	return nil
}

func (*resourceBindingAuthorizerStub) AuthorizeCreate(
	context.Context,
	control.Scope,
	control.ResourceTarget,
) error {
	return nil
}

func (*resourceBindingAuthorizerStub) AuthorizeUpdate(
	context.Context,
	control.Scope,
	control.ResourceHead,
) error {
	return nil
}

func (stub *resourceBindingAuthorizerStub) AuthorizeReference(
	context.Context,
	control.Scope,
	control.ResourceHead,
) error {
	stub.referenceCalls++
	return stub.err
}

func resourceBindingTargetKey(target control.ResourceTarget) string {
	return target.APIVersion + "/" + target.Kind + "/" +
		target.Namespace.String() + "/" + target.Name.String()
}

func resourceBindingRevisionKey(resourceID int64, revision int64) string {
	return fmt.Sprintf("%d/%d", resourceID, revision)
}
