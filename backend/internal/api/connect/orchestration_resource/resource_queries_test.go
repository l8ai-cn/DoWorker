package orchestrationresourceconnect

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	resourcev1 "github.com/anthropics/agentsmesh/proto/gen/go/orchestration_resource/v1"
)

func TestGetResourceConvertsCompleteHeadAndUsesResolvedScope(t *testing.T) {
	stub := &serviceStub{getResult: testHead()}
	server := newTestServer(stub, testOrganizations())

	response, err := server.GetResource(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.GetResourceRequest{
			OrgSlug: "acme",
			Target:  protoTarget("acme"),
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), stub.getScope)
	assert.Equal(t, testTarget(), stub.getTarget)
	assert.EqualValues(t, 9, response.Msg.Id)
	assert.Equal(t, testResourceID, response.Msg.Identity.Uid)
	assert.Equal(t, "Builder", response.Msg.DisplayName)
	assert.Equal(t, map[string]string{"team": "platform"}, response.Msg.Labels)
	assert.Equal(t, []byte(`{"phase":"ready"}`), response.Msg.StatusJson)
	assert.EqualValues(t, 3, response.Msg.Revision)
	assert.EqualValues(t, 2, response.Msg.Generation)
	assert.EqualValues(t, 7, response.Msg.ResourceVersion)
	assert.EqualValues(t, 40, response.Msg.CreatedById)
	assert.EqualValues(t, 42, response.Msg.UpdatedById)
	assert.Equal(t, "2026-07-14T00:30:00Z", response.Msg.CreatedAt)
	assert.Equal(t, "2026-07-14T01:30:00Z", response.Msg.UpdatedAt)
}

func TestGetResourceRejectsTargetOutsideResolvedOrganization(t *testing.T) {
	stub := &serviceStub{}
	server := newTestServer(stub, testOrganizations())

	_, err := server.GetResource(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.GetResourceRequest{
			OrgSlug: "acme",
			Target:  protoTarget("other-org"),
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	assert.Empty(t, stub.getTarget.Name)
}

func TestListResourcesUsesDefaultsAndServiceTotal(t *testing.T) {
	head := testHead()
	second := testHead()
	second.ID = 10
	second.Identity.Name = "reviewer"
	stub := &serviceStub{}
	stub.listResult.Items = append(stub.listResult.Items, head, second)
	stub.listResult.Total = 19
	server := newTestServer(stub, testOrganizations())

	response, err := server.ListResources(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ListResourcesRequest{OrgSlug: "acme"}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), stub.listScope)
	assert.Equal(t, service.ResourceListFilter{Limit: 50, Offset: 0}, stub.listFilter)
	assert.Len(t, response.Msg.Items, 2)
	assert.EqualValues(t, 19, response.Msg.Total)
	assert.EqualValues(t, 50, response.Msg.Limit)
	assert.Zero(t, response.Msg.Offset)
}

func TestListResourcesConvertsExplicitPaginationAndRejectsOverMaximum(t *testing.T) {
	kind := "WorkerTemplate"
	offset := int32(7)
	limit := int32(100)
	stub := &serviceStub{}
	server := newTestServer(stub, testOrganizations())

	response, err := server.ListResources(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ListResourcesRequest{
			OrgSlug: "acme",
			Kind:    &kind,
			Offset:  &offset,
			Limit:   &limit,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, service.ResourceListFilter{Kind: kind, Limit: 100, Offset: 7}, stub.listFilter)
	assert.EqualValues(t, 100, response.Msg.Limit)
	assert.EqualValues(t, 7, response.Msg.Offset)

	tooLarge := int32(101)
	_, err = server.ListResources(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ListResourcesRequest{
			OrgSlug: "acme",
			Limit:   &tooLarge,
		}),
	)
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestListResourcesConvertsEnvironmentBundleReferenceFilter(t *testing.T) {
	kind := "EnvironmentBundle"
	stub := &serviceStub{}
	server := newTestServer(stub, testOrganizations())

	response, err := server.ListResources(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ListResourcesRequest{
			OrgSlug: "acme",
			Kind:    &kind,
			EnvironmentBundleFilter: &resourcev1.EnvironmentBundleReferenceFilter{
				Purpose:    resourcev1.EnvironmentBundlePurpose_ENVIRONMENT_BUNDLE_PURPOSE_CREDENTIAL,
				WorkerType: "do-agent",
				TargetName: "DO_API_KEY",
			},
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, service.ResourceListFilter{
		Kind: "EnvironmentBundle", Limit: 50,
		EnvironmentBundle: &service.EnvironmentBundleReferenceFilter{
			Purpose:    service.EnvironmentBundlePurposeCredential,
			WorkerType: slugkit.Slug("do-agent"),
			TargetName: "DO_API_KEY",
		},
	}, stub.listFilter)
	assert.Equal(t, requestEnvironmentBundleFilter(
		resourcev1.EnvironmentBundlePurpose_ENVIRONMENT_BUNDLE_PURPOSE_CREDENTIAL,
		"do-agent",
		"DO_API_KEY",
	), response.Msg.AppliedEnvironmentBundleFilter)
}

func TestListResourcesRejectsEnvironmentBundleFilterForAnotherKind(t *testing.T) {
	kind := "Prompt"
	stub := &serviceStub{}
	server := newTestServer(stub, testOrganizations())

	_, err := server.ListResources(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ListResourcesRequest{
			OrgSlug: "acme",
			Kind:    &kind,
			EnvironmentBundleFilter: &resourcev1.EnvironmentBundleReferenceFilter{
				Purpose:    resourcev1.EnvironmentBundlePurpose_ENVIRONMENT_BUNDLE_PURPOSE_RUNTIME,
				WorkerType: "do-agent",
			},
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	assert.Nil(t, stub.listFilter.EnvironmentBundle)
}

func requestEnvironmentBundleFilter(
	purpose resourcev1.EnvironmentBundlePurpose,
	workerType string,
	targetName string,
) *resourcev1.EnvironmentBundleReferenceFilter {
	return &resourcev1.EnvironmentBundleReferenceFilter{
		Purpose: purpose, WorkerType: workerType, TargetName: targetName,
	}
}

func TestExportResourceConvertsTargetRevisionAndFormat(t *testing.T) {
	revision := int64(3)
	stub := &serviceStub{exportResult: service.ResourceExport{
		Format:  service.SourceFormatYAML,
		Content: []byte("kind: WorkerTemplate\n"),
	}}
	server := newTestServer(stub, testOrganizations())

	response, err := server.ExportResource(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ExportResourceRequest{
			OrgSlug:  "acme",
			Target:   protoTarget("acme"),
			Revision: &revision,
			Format:   resourcev1.SourceFormat_SOURCE_FORMAT_YAML,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), stub.exportRequest.Scope)
	assert.Equal(t, testTarget(), stub.exportRequest.Target)
	assert.EqualValues(t, 3, stub.exportRequest.Revision)
	assert.Equal(t, service.SourceFormatYAML, stub.exportRequest.Format)
	assert.Equal(t, []byte("kind: WorkerTemplate\n"), response.Msg.Content)
}

func TestExportResourceRejectsNegativeRevision(t *testing.T) {
	revision := int64(-1)
	stub := &serviceStub{}
	server := newTestServer(stub, testOrganizations())

	_, err := server.ExportResource(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ExportResourceRequest{
			OrgSlug:  "acme",
			Target:   protoTarget("acme"),
			Revision: &revision,
			Format:   resourcev1.SourceFormat_SOURCE_FORMAT_JSON,
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestGetResourcePlanConvertsPlanAndBindsActor(t *testing.T) {
	stub := &serviceStub{getPlanResult: testPlan()}
	server := newTestServer(stub, testOrganizations())

	response, err := server.GetResourcePlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.GetResourcePlanRequest{
			OrgSlug: "acme",
			PlanId:  testPlanID,
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, testScope(), stub.getPlanScope)
	assert.Equal(t, testPlanID, stub.getPlanID)
	assert.Equal(t, testPlanID, response.Msg.PlanId)
	assert.Equal(t, testResourceID, response.Msg.Base.Uid)
	assert.Equal(t, "catalog-7", response.Msg.OptionsRevision)
}

func TestGetResourcePlanRejectsInvalidPlanID(t *testing.T) {
	stub := &serviceStub{}
	server := newTestServer(stub, testOrganizations())

	_, err := server.GetResourcePlan(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.GetResourcePlanRequest{
			OrgSlug: "acme",
			PlanId:  "not-a-uuid",
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	assert.Empty(t, stub.getPlanID)
}

func protoTarget(namespace string) *resourcev1.ResourceTarget {
	return &resourcev1.ResourceTarget{
		TypeMeta: &resourcev1.TypeMeta{
			ApiVersion: "agentsmesh.io/v1alpha1",
			Kind:       "WorkerTemplate",
		},
		Namespace: namespace,
		Name:      "builder",
	}
}
