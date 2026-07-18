package orchestrationcontrol

import (
	"context"
	"encoding/json"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceGetResourceAuthorizesExactHead(t *testing.T) {
	head := orchestrationServiceHead()
	repository := &resourceQueryRepositoryStub{head: head}
	authorizer := &orchestrationAuthorizerStub{}
	service := &Service{repository: repository, authorizer: authorizer}

	result, err := service.GetResource(
		context.Background(),
		orchestrationServiceScope(),
		head.Identity.ResourceTarget,
	)

	require.NoError(t, err)
	assert.Equal(t, head, result)
	assert.Equal(t, 1, authorizer.referenceCalls)
}

func TestServiceListResourcesAuthorizesEmptyTenantList(t *testing.T) {
	repository := &resourceQueryRepositoryStub{
		heads: []control.ResourceHead{},
		total: 17,
	}
	authorizer := &orchestrationAuthorizerStub{}
	service := &Service{
		repository: repository, authorizer: authorizer,
		workerDefinitions: workerDefinitionPolicyStub{},
	}

	result, err := service.ListResources(
		context.Background(),
		orchestrationServiceScope(),
		ResourceListFilter{Kind: "WorkerTemplate", Limit: 50},
	)

	require.NoError(t, err)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(17), result.Total)
	assert.Equal(t, 1, authorizer.listCalls)
}

func TestServiceExportResourceUsesRequestedHistoricalRevision(t *testing.T) {
	head, _ := referenceResolverResource()
	revision := referenceResolverRevision(head, 2, 2, 2)
	repository := &resourceQueryRepositoryStub{
		head: head,
		revisions: map[int64]control.ResourceRevision{
			2: revision,
		},
	}
	service := &Service{
		repository: repository,
		authorizer: &orchestrationAuthorizerStub{},
	}

	result, err := service.ExportResource(
		context.Background(),
		ExportResourceRequest{
			Scope:    orchestrationServiceScope(),
			Target:   head.Identity.ResourceTarget,
			Revision: 2,
			Format:   SourceFormatYAML,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, SourceFormatYAML, result.Format)
	assert.Contains(t, string(result.Content), "kind: ModelBinding")
	assert.Equal(t, int64(2), repository.requestedRevision)
}

func TestServiceExportResourceRejectsUnknownStoredManifestFields(t *testing.T) {
	head, revision := referenceResolverResource()
	var document map[string]any
	require.NoError(t, json.Unmarshal(revision.CanonicalManifest, &document))
	document["unknown"] = true
	manifest, err := control.CanonicalJSONObject(document)
	require.NoError(t, err)
	revision.CanonicalManifest = manifest
	revision.Digest, err = control.DigestCanonicalJSON(manifest)
	require.NoError(t, err)
	repository := &resourceQueryRepositoryStub{
		head: head,
		revisions: map[int64]control.ResourceRevision{
			revision.Revision: revision,
		},
	}
	service := &Service{
		repository: repository,
		authorizer: &orchestrationAuthorizerStub{},
	}

	_, err = service.ExportResource(
		context.Background(),
		ExportResourceRequest{
			Scope:  orchestrationServiceScope(),
			Target: head.Identity.ResourceTarget,
			Format: SourceFormatJSON,
		},
	)

	assert.ErrorIs(t, err, control.ErrCorrupt)
}

func TestServiceGetResourcePlanPreservesActorScope(t *testing.T) {
	scope := orchestrationServiceScope()
	fixture := newOrchestrationServiceFixture(t)
	planned, err := fixture.service(t).Plan(context.Background(), PlanRequest{
		Scope: scope,
		Source: ResourceSource{
			Format: SourceFormatJSON, Content: []byte(testResourceJSON),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, planned.Plan)
	plan := *planned.Plan
	repository := &resourceQueryRepositoryStub{plan: plan}
	service := &Service{repository: repository}

	result, err := service.GetResourcePlan(
		context.Background(),
		scope,
		plan.ID,
	)

	require.NoError(t, err)
	assert.Equal(t, plan, result)
	assert.Equal(t, scope, repository.planScope)
}

type resourceQueryRepositoryStub struct {
	Repository
	head              control.ResourceHead
	heads             []control.ResourceHead
	total             int64
	revisions         map[int64]control.ResourceRevision
	plan              control.Plan
	requestedRevision int64
	planScope         control.Scope
	listFilter        ResourceListFilter
}

func (stub *resourceQueryRepositoryStub) GetResource(
	context.Context,
	control.Scope,
	control.ResourceTarget,
) (control.ResourceHead, error) {
	return stub.head, nil
}

func (stub *resourceQueryRepositoryStub) ListResources(
	_ context.Context,
	_ control.Scope,
	filter ResourceListFilter,
) (ResourceListPage, error) {
	stub.listFilter = filter
	return ResourceListPage{
		Items: append([]control.ResourceHead{}, stub.heads...),
		Total: stub.total,
	}, nil
}

func (stub *resourceQueryRepositoryStub) GetRevision(
	_ context.Context,
	_ control.Scope,
	_ int64,
	revision int64,
) (control.ResourceRevision, error) {
	stub.requestedRevision = revision
	return stub.revisions[revision], nil
}

func (stub *resourceQueryRepositoryStub) GetPlan(
	_ context.Context,
	scope control.Scope,
	_ string,
) (control.Plan, error) {
	stub.planScope = scope
	return stub.plan, nil
}
