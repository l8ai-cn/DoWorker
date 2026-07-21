package orchestrationresourceconnect

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	service "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	resourcev1 "github.com/l8ai-cn/agentcloud/proto/gen/go/orchestration_resource/v1"
)

func TestMapServiceErrorUsesStableCodesWithoutSensitiveDetails(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code connect.Code
	}{
		{name: "invalid", err: errors.Join(control.ErrInvalid, errors.New(`Secret canonical manifest {"token":"x"}`)), code: connect.CodeInvalidArgument},
		{name: "forbidden", err: errors.Join(service.ErrForbidden, errors.New("SELECT * FROM plans")), code: connect.CodePermissionDenied},
		{name: "not found", err: control.ErrNotFound, code: connect.CodeNotFound},
		{name: "conflict", err: control.ErrConflict, code: connect.CodeAborted},
		{name: "stale", err: control.ErrStale, code: connect.CodeAborted},
		{name: "expired", err: control.ErrExpired, code: connect.CodeAborted},
		{name: "consumed", err: control.ErrConsumed, code: connect.CodeAborted},
		{name: "stale options", err: service.ErrStaleOptions, code: connect.CodeAborted},
		{name: "unavailable", err: service.ErrUnavailable, code: connect.CodeUnavailable},
		{name: "corrupt", err: control.ErrCorrupt, code: connect.CodeInternal},
		{name: "unknown", err: errors.New(`artifact JSON {"private":"value"} SQLSTATE 23505`), code: connect.CodeInternal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mapped := mapServiceError(test.err)
			require.Error(t, mapped)
			assert.Equal(t, test.code, connect.CodeOf(mapped))
			assert.NotContains(t, mapped.Error(), "Secret")
			assert.NotContains(t, mapped.Error(), "canonical")
			assert.NotContains(t, mapped.Error(), "artifact")
			assert.NotContains(t, mapped.Error(), "SELECT")
			assert.NotContains(t, mapped.Error(), "SQLSTATE")
			assert.NotContains(t, mapped.Error(), "private")
		})
	}
}

func TestNewServerRejectsNilDependencies(t *testing.T) {
	type dependencies struct {
		service             Service
		bindingApply        BindingPlanApplier
		workerTemplateApply WorkerTemplatePlanApplier
		workerApply         WorkerPlanCreator
		promptApply         PromptPlanApplier
		expertApply         ExpertPlanApplier
		workflowApply       WorkflowPlanApplier
		goalLoopApply       GoalLoopPlanCreator
		orgs                middleware.OrganizationService
	}
	valid := func() dependencies {
		return dependencies{
			service:             &serviceStub{},
			bindingApply:        &bindingApplyStub{},
			workerTemplateApply: &workerTemplateApplyStub{},
			workerApply:         &workerApplyStub{},
			promptApply:         &promptApplyStub{},
			expertApply:         &expertApplyStub{},
			workflowApply:       &workflowApplyStub{},
			goalLoopApply:       &goalLoopApplyStub{},
			orgs:                testOrganizations(),
		}
	}
	cases := []func(*dependencies){
		func(value *dependencies) { value.service = nil },
		func(value *dependencies) { value.bindingApply = nil },
		func(value *dependencies) { value.workerTemplateApply = nil },
		func(value *dependencies) { value.workerApply = nil },
		func(value *dependencies) { value.promptApply = nil },
		func(value *dependencies) { value.expertApply = nil },
		func(value *dependencies) { value.workflowApply = nil },
		func(value *dependencies) { value.goalLoopApply = nil },
		func(value *dependencies) { value.orgs = nil },
	}
	for _, invalidate := range cases {
		value := valid()
		invalidate(&value)
		assert.Panics(t, func() {
			NewServer(
				value.service,
				value.bindingApply,
				value.workerTemplateApply,
				value.workerApply,
				value.promptApply,
				value.expertApply,
				value.workflowApply,
				value.goalLoopApply,
				value.orgs,
			)
		})
	}
	var typedNilService *serviceStub
	assert.Panics(t, func() {
		NewServer(
			typedNilService,
			&bindingApplyStub{},
			&workerTemplateApplyStub{},
			&workerApplyStub{},
			&promptApplyStub{},
			&expertApplyStub{},
			&workflowApplyStub{},
			&goalLoopApplyStub{},
			testOrganizations(),
		)
	})
	var typedNilGoalLoop *goalLoopApplyStub
	assert.Panics(t, func() {
		NewServer(
			&serviceStub{},
			&bindingApplyStub{},
			&workerTemplateApplyStub{},
			&workerApplyStub{},
			&promptApplyStub{},
			&expertApplyStub{},
			&workflowApplyStub{},
			typedNilGoalLoop,
			testOrganizations(),
		)
	})
	var typedNilOrganizations *organizationStub
	assert.Panics(t, func() {
		NewServer(
			&serviceStub{},
			&bindingApplyStub{},
			&workerTemplateApplyStub{},
			&workerApplyStub{},
			&promptApplyStub{},
			&expertApplyStub{},
			&workflowApplyStub{},
			&goalLoopApplyStub{},
			typedNilOrganizations,
		)
	})
}

func TestResolveOrgScopeInternalErrorDoesNotLeakSQL(t *testing.T) {
	stub := &serviceStub{}
	organizations := testOrganizations()
	organizations.roleErr = errors.New("role lookup failed")
	organizations.memberErr = errors.New("SELECT secret FROM organization_members")
	server := newTestServer(stub, organizations)

	_, err := server.ValidateResource(
		authenticatedContext(42),
		connect.NewRequest(&resourcev1.ValidateResourceRequest{
			OrgSlug: "acme",
			Source: &resourcev1.ResourceSource{
				Format:  resourcev1.SourceFormat_SOURCE_FORMAT_JSON,
				Content: []byte(`{}`),
			},
		}),
	)

	require.Error(t, err)
	assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	assert.NotContains(t, err.Error(), "SELECT")
	assert.NotContains(t, err.Error(), "secret")
}

func TestMountRegistersExactlyFourteenResourceProcedures(t *testing.T) {
	server := newTestServer(&serviceStub{}, testOrganizations())
	mux := http.NewServeMux()
	Mount(mux, server)

	expected := []string{
		"/proto.orchestration_resource.v1.OrchestrationResourceService/ValidateResource",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/PlanResource",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/GetResource",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/GetResourceCapabilities",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/ListResources",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/ExportResource",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/GetResourcePlan",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyBindingResourcePlan",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyWorkerTemplatePlan",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/CreateWorkerFromPlan",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyPromptPlan",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyExpertPlan",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/ApplyWorkflowPlan",
		"/proto.orchestration_resource.v1.OrchestrationResourceService/CreateGoalLoopFromPlan",
	}
	for _, procedure := range expected {
		_, pattern := mux.Handler(httptest.NewRequest(http.MethodPost, procedure, nil))
		assert.Equal(t, procedure, pattern)
	}
}
