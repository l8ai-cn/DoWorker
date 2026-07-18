package grpc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	poddomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	workflowservice "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mcpPlanID = "11111111-1111-4111-8111-111111111111"

func TestMcpCreatePodUsesResourceControlAndPinnedWorkerApply(t *testing.T) {
	ctx := context.Background()
	db := testkit.SetupTestDB(t)
	pods := agentpod.NewPodService(infra.NewPodRepository(db))
	pod := &poddomain.Pod{
		OrganizationID: 7, CreatedByID: 42, RunnerID: 11, ClusterID: 3,
		PodKey: "42-standalone-resource", AgentSlug: "resource-native",
		Status: poddomain.StatusInitializing, AgentStatus: poddomain.AgentStatusIdle,
		InteractionMode: poddomain.InteractionModeACP,
		AutomationLevel: poddomain.AutomationLevelAutonomous,
	}
	require.NoError(t, db.Create(pod).Error)
	calls := []string{}
	resourceControl := newMCPResourceControlStub(&calls, resource.KindWorker)
	adapter := &GRPCRunnerAdapter{
		resourceControl: resourceControl,
		workerPlanApply: &mcpWorkerApplyStub{
			calls: &calls,
			result: workerplanner.AppliedWorker{
				Head:                 mcpAppliedHead(resource.KindWorker, "review-run"),
				WorkerSpecSnapshotID: 91,
				PodKey:               pod.PodKey,
			},
		},
		podService: pods,
	}

	result, mcpErr := adapter.mcpCreatePod(
		ctx,
		mcpTenant(),
		mcpResourcePayload(resource.KindWorker, "review-run", false),
	)

	require.Nil(t, mcpErr)
	response := result.(map[string]interface{})
	assert.Equal(t, pod.PodKey, response["pod"].(map[string]interface{})["pod_key"])
	assert.Equal(t, int64(91), response["resource"].(map[string]interface{})["worker_spec_snapshot_id"])
	assert.Equal(t, []string{"validate", "plan", "authorize", "apply-worker"}, calls)
	assert.Equal(t, mcpTenantScope(), resourceControl.validateRequest.Scope)
	assert.JSONEq(t, string(resourceControl.validateRequest.Source.Content), string(resourceControl.planRequest.Source.Content))
}

func TestMcpCreatePodRejectsNonWorkerResource(t *testing.T) {
	calls := []string{}
	adapter := &GRPCRunnerAdapter{
		resourceControl: newMCPResourceControlStub(&calls, resource.KindWorkflow),
		workerPlanApply: &mcpWorkerApplyStub{calls: &calls},
		podService:      agentpod.NewPodService(nil),
	}

	_, mcpErr := adapter.mcpCreatePod(
		context.Background(),
		mcpTenant(),
		mcpResourcePayload(resource.KindWorkflow, "nightly-review", false),
	)

	require.NotNil(t, mcpErr)
	assert.Equal(t, int32(400), mcpErr.code)
	assert.Equal(t, []string{"validate"}, calls)
}

func TestMcpCreatePodRejectsLegacyRuntimePayload(t *testing.T) {
	calls := []string{}
	adapter := &GRPCRunnerAdapter{
		resourceControl: newMCPResourceControlStub(&calls, resource.KindWorker),
		workerPlanApply: &mcpWorkerApplyStub{calls: &calls},
		podService:      agentpod.NewPodService(nil),
	}

	_, mcpErr := adapter.mcpCreatePod(
		context.Background(),
		mcpTenant(),
		[]byte(`{"runner_id":11,"agent_slug":"codex-cli"}`),
	)

	require.NotNil(t, mcpErr)
	assert.Equal(t, int32(400), mcpErr.code)
	assert.Equal(t, "resource is required", mcpErr.message)
	assert.Empty(t, calls)
}

func TestMcpCreateWorkflowAppliesResourceAndDisablesByDefault(t *testing.T) {
	ctx := context.Background()
	workflowService, workflowID := mcpWorkflowProjection(t, ctx, "daily-review")
	_, err := workflowService.SetStatus(
		ctx,
		7,
		"daily-review",
		workflowDomain.StatusDisabled,
	)
	require.NoError(t, err)
	calls := []string{}
	applier := &mcpWorkflowApplyStub{
		calls: &calls,
		result: workerplanner.AppliedWorkflow{
			Head:                 mcpAppliedHead(resource.KindWorkflow, "daily-review"),
			WorkflowID:           workflowID,
			WorkerSpecSnapshotID: 92,
		},
	}
	adapter := &GRPCRunnerAdapter{
		resourceControl:   newMCPResourceControlStub(&calls, resource.KindWorkflow),
		workflowPlanApply: applier,
		workflowService:   workflowService,
	}

	result, mcpErr := adapter.mcpCreateWorkflow(
		ctx,
		mcpTenant(),
		"7-standalone-caller",
		mcpResourcePayload(resource.KindWorkflow, "daily-review", false),
	)

	require.Nil(t, mcpErr)
	response := result.(map[string]interface{})
	assert.Equal(t, "disabled", response["workflow"].(*mcpWorkflowSummary).Status)
	assert.Equal(t, int64(92), response["resource"].(map[string]interface{})["worker_spec_snapshot_id"])
	assert.Equal(t, []string{"validate", "plan", "authorize", "apply-workflow"}, calls)
	assert.Equal(t, workflowDomain.StatusDisabled, applier.status)
}

func TestMcpCreateWorkflowRejectsLegacyDefinitionPayload(t *testing.T) {
	calls := []string{}
	adapter := &GRPCRunnerAdapter{
		resourceControl: newMCPResourceControlStub(&calls, resource.KindWorkflow),
		workflowPlanApply: &mcpWorkflowApplyStub{
			calls: &calls,
		},
		workflowService: workflowservice.NewWorkflowService(nil),
	}

	_, mcpErr := adapter.mcpCreateWorkflow(
		context.Background(),
		mcpTenant(),
		"",
		[]byte(`{"name":"daily-review","prompt_template":"review"}`),
	)

	require.NotNil(t, mcpErr)
	assert.Equal(t, int32(400), mcpErr.code)
	assert.Equal(t, "resource is required", mcpErr.message)
	assert.Empty(t, calls)
}

func TestMcpCreateWorkflowKeepsAppliedWorkflowEnabledWhenConfirmed(t *testing.T) {
	ctx := context.Background()
	workflowService, workflowID := mcpWorkflowProjection(t, ctx, "nightly-sync")
	calls := []string{}
	applier := &mcpWorkflowApplyStub{
		calls: &calls,
		result: workerplanner.AppliedWorkflow{
			Head:       mcpAppliedHead(resource.KindWorkflow, "nightly-sync"),
			WorkflowID: workflowID, WorkerSpecSnapshotID: 93,
		},
	}
	adapter := &GRPCRunnerAdapter{
		resourceControl:   newMCPResourceControlStub(&calls, resource.KindWorkflow),
		workflowPlanApply: applier,
		workflowService:   workflowService,
	}

	result, mcpErr := adapter.mcpCreateWorkflow(
		ctx,
		mcpTenant(),
		"",
		mcpResourcePayload(resource.KindWorkflow, "nightly-sync", true),
	)

	require.Nil(t, mcpErr)
	assert.Equal(
		t,
		"enabled",
		result.(map[string]interface{})["workflow"].(*mcpWorkflowSummary).Status,
	)
	assert.Equal(t, workflowDomain.StatusEnabled, applier.status)
}

func TestPlanMCPResourceStopsOnBlockingValidationIssue(t *testing.T) {
	calls := []string{}
	stub := newMCPResourceControlStub(&calls, resource.KindWorker)
	stub.validateResult.Issues = []control.PlanIssue{{
		Severity: control.PlanIssueBlocking,
		Path:     "/spec/workerTemplateRef",
		Code:     "reference.not-found",
		Message:  "Referenced WorkerTemplate was not found.",
	}}
	adapter := &GRPCRunnerAdapter{resourceControl: stub}

	_, _, mcpErr := adapter.planMCPResource(
		context.Background(),
		mcpTenant(),
		json.RawMessage(`{"kind":"Worker"}`),
		resource.KindWorker,
	)

	require.NotNil(t, mcpErr)
	assert.Equal(t, int32(400), mcpErr.code)
	assert.Contains(t, mcpErr.message, "reference.not-found")
	assert.Equal(t, []string{"validate"}, calls)
}

type mcpResourceControlStub struct {
	calls           *[]string
	validateResult  controlservice.ValidationResult
	planResult      controlservice.PlanResult
	validateRequest controlservice.ValidateRequest
	planRequest     controlservice.PlanRequest
}

func newMCPResourceControlStub(
	calls *[]string,
	kind string,
) *mcpResourceControlStub {
	target := control.ResourceTarget{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       kind,
		},
		Namespace: "test-org",
		Name:      "resource-name",
	}
	plan := control.Plan{ID: mcpPlanID}
	return &mcpResourceControlStub{
		calls: calls,
		validateResult: controlservice.ValidationResult{
			Target: target, Operation: control.PlanOperationCreate,
			Issues: []control.PlanIssue{},
		},
		planResult: controlservice.PlanResult{
			ValidationResult: controlservice.ValidationResult{
				Target: target, Operation: control.PlanOperationCreate,
				Issues: []control.PlanIssue{},
			},
			Plan: &plan,
		},
	}
}

func (stub *mcpResourceControlStub) Validate(
	_ context.Context,
	request controlservice.ValidateRequest,
) (controlservice.ValidationResult, error) {
	*stub.calls = append(*stub.calls, "validate")
	stub.validateRequest = request
	return stub.validateResult, nil
}

func (stub *mcpResourceControlStub) Plan(
	_ context.Context,
	request controlservice.PlanRequest,
) (controlservice.PlanResult, error) {
	*stub.calls = append(*stub.calls, "plan")
	stub.planRequest = request
	return stub.planResult, nil
}

func (stub *mcpResourceControlStub) AuthorizeApply(
	_ context.Context,
	_ control.Scope,
	_ string,
) error {
	*stub.calls = append(*stub.calls, "authorize")
	return nil
}

type mcpWorkerApplyStub struct {
	calls  *[]string
	result workerplanner.AppliedWorker
}

func (stub *mcpWorkerApplyStub) Apply(
	_ context.Context,
	_ control.Scope,
	_ string,
) (workerplanner.AppliedWorker, error) {
	*stub.calls = append(*stub.calls, "apply-worker")
	return stub.result, nil
}

type mcpWorkflowApplyStub struct {
	calls  *[]string
	result workerplanner.AppliedWorkflow
	status string
}

func (stub *mcpWorkflowApplyStub) ApplyWithStatus(
	_ context.Context,
	_ control.Scope,
	_ string,
	status string,
) (workerplanner.AppliedWorkflow, error) {
	*stub.calls = append(*stub.calls, "apply-workflow")
	stub.status = status
	return stub.result, nil
}

func mcpTenant() *middleware.TenantContext {
	return &middleware.TenantContext{
		OrganizationID: 7, OrganizationSlug: "test-org", UserID: 42,
	}
}

func mcpTenantScope() control.Scope {
	return control.Scope{
		OrganizationID: 7, OrganizationSlug: "test-org", ActorID: 42,
	}
}

func mcpAppliedHead(kind string, name slugkit.Slug) control.ResourceHead {
	return control.ResourceHead{
		Identity: control.ResourceIdentity{
			ResourceTarget: control.ResourceTarget{
				TypeMeta: resource.TypeMeta{
					APIVersion: resource.APIVersionV1Alpha1,
					Kind:       kind,
				},
				Namespace: "test-org", Name: name,
			},
			UID: "22222222-2222-4222-8222-222222222222",
		},
		Revision: 1, CreatedAt: time.Now(),
	}
}

func mcpResourcePayload(kind, name string, enabled bool) []byte {
	payload, _ := json.Marshal(map[string]interface{}{
		"resource": map[string]interface{}{
			"apiVersion": resource.APIVersionV1Alpha1,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name": name, "namespace": "test-org",
			},
			"spec": map[string]interface{}{"fixture": true},
		},
		"enabled": enabled,
	})
	return payload
}

func mcpWorkflowProjection(
	t *testing.T,
	ctx context.Context,
	name string,
) (*workflowservice.WorkflowService, int64) {
	t.Helper()
	db := testkit.SetupTestDB(t)
	service := workflowservice.NewWorkflowService(infra.NewWorkflowRepository(db))
	workflow, err := service.Create(ctx, &workflowservice.CreateWorkflowRequest{
		OrganizationID: 7, CreatedByID: 42, Name: name,
		AgentSlug: "resource-native", PromptTemplate: "review",
		ExecutionMode: "direct", SandboxStrategy: "fresh",
		ConcurrencyPolicy: "skip", TimeoutMinutes: 60,
		MaxConcurrentRuns: 1, MaxRetainedRuns: 30,
	})
	require.NoError(t, err)
	return service, workflow.ID
}
