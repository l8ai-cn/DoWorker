package grpc

import (
	"context"
	"errors"
	"testing"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	agentsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationworker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMcpCreatePodAuthorizesAndAppliesWorkerPlan(t *testing.T) {
	authorizer := &recordingMCPWorkerPlanAuthorizer{}
	applier := &recordingMCPWorkerPlanApplier{
		result: workerplanner.AppliedWorker{
			PodKey:   "7-standalone-12345678",
			RunnerID: 11,
		},
	}
	reader := &mcpWorkerPodReaderStub{pod: &podDomain.Pod{
		PodKey: applier.result.PodKey,
		Status: podDomain.StatusRunning,
	}}
	adapter := &GRPCRunnerAdapter{
		workerPlanAuthorizer: authorizer,
		workerPlanApplier:    applier,
		workerPodReader:      reader,
	}
	tenant := &middleware.TenantContext{
		OrganizationID: 42, OrganizationSlug: "team-alpha", UserID: 7,
	}

	result, mcpErr := adapter.mcpCreatePod(
		context.Background(),
		tenant,
		[]byte(`{"plan_id":"11111111-1111-4111-8111-111111111111"}`),
	)

	require.Nil(t, mcpErr)
	expectedScope := control.Scope{
		OrganizationID: 42, OrganizationSlug: "team-alpha", ActorID: 7,
	}
	assert.Equal(t, expectedScope, authorizer.scope)
	assert.Equal(t, expectedScope, applier.scope)
	assert.Equal(t, authorizer.planID, applier.planID)
	response := result.(map[string]interface{})
	assert.Equal(t, map[string]interface{}{
		"pod_key": applier.result.PodKey,
		"status":  podDomain.StatusRunning,
	}, response["pod"])
}

func TestMcpCreatePodAuthorizationFailureDoesNotApply(t *testing.T) {
	applier := &recordingMCPWorkerPlanApplier{}
	adapter := &GRPCRunnerAdapter{
		workerPlanAuthorizer: &recordingMCPWorkerPlanAuthorizer{
			err: controlservice.ErrForbidden,
		},
		workerPlanApplier: applier,
		workerPodReader:   &mcpWorkerPodReaderStub{},
	}
	tenant := &middleware.TenantContext{
		OrganizationID: 42, OrganizationSlug: "team-alpha", UserID: 7,
	}

	_, mcpErr := adapter.mcpCreatePod(
		context.Background(),
		tenant,
		[]byte(`{"plan_id":"11111111-1111-4111-8111-111111111111"}`),
	)

	require.NotNil(t, mcpErr)
	assert.Equal(t, int32(403), mcpErr.code)
	assert.Empty(t, applier.planID)
}

func TestMapWorkerPlanErrorToMCPUsesStableRedactedErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		code    int32
		message string
	}{
		{name: "invalid", err: control.ErrInvalid, code: 400, message: "Worker plan is invalid"},
		{name: "forbidden", err: controlservice.ErrForbidden, code: 403, message: "Worker plan access is forbidden"},
		{name: "not found", err: control.ErrNotFound, code: 404, message: "Worker plan was not found"},
		{name: "conflict", err: control.ErrConflict, code: 409, message: "Worker plan state changed; create a new plan"},
		{name: "stale", err: control.ErrStale, code: 409, message: "Worker plan state changed; create a new plan"},
		{name: "expired", err: control.ErrExpired, code: 409, message: "Worker plan state changed; create a new plan"},
		{name: "consumed", err: control.ErrConsumed, code: 409, message: "Worker plan state changed; create a new plan"},
		{name: "queue full", err: podDomain.ErrQueueFull, code: 429, message: "Runner pending queue is full"},
		{name: "no runner", err: agentsvc.ErrNoAvailableRunner, code: 422, message: "No runner is available for the Worker snapshot"},
		{name: "unavailable", err: controlservice.ErrUnavailable, code: 503, message: "Worker apply service is unavailable"},
		{name: "unknown", err: errors.New("database password=super-secret"), code: 500, message: "failed to apply Worker plan"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := mapWorkerPlanErrorToMCP(test.err)
			assert.Equal(t, test.code, got.code)
			assert.Equal(t, test.message, got.message)
			assert.NotContains(t, got.message, "super-secret")
		})
	}
}

func TestDecodeMCPWorkerPlanPayloadUsesStableRedactedErrors(t *testing.T) {
	payloads := [][]byte{
		[]byte(`{"plan_id":"11111111-1111-4111-8111-111111111111","prompt":"super-secret"}`),
		[]byte(`{"plan_id":{"secret":"super-secret"}}`),
		[]byte(`{"plan_id":"11111111-1111-4111-8111-111111111111"`),
	}

	for _, payload := range payloads {
		_, got := decodeMCPWorkerPlanPayload(payload)
		require.NotNil(t, got)
		assert.Equal(t, int32(400), got.code)
		assert.Equal(t, "invalid request payload", got.message)
		assert.NotContains(t, got.message, "super-secret")
		assert.NotContains(t, got.message, "prompt")
	}
}

type recordingMCPWorkerPlanAuthorizer struct {
	scope  control.Scope
	planID string
	err    error
}

func (stub *recordingMCPWorkerPlanAuthorizer) AuthorizeApply(
	_ context.Context,
	scope control.Scope,
	planID string,
) error {
	stub.scope = scope
	stub.planID = planID
	return stub.err
}

type recordingMCPWorkerPlanApplier struct {
	scope  control.Scope
	planID string
	result workerplanner.AppliedWorker
	err    error
}

func (stub *recordingMCPWorkerPlanApplier) Apply(
	_ context.Context,
	scope control.Scope,
	planID string,
) (workerplanner.AppliedWorker, error) {
	stub.scope = scope
	stub.planID = planID
	return stub.result, stub.err
}

type mcpWorkerPodReaderStub struct {
	pod *podDomain.Pod
	err error
}

func (stub *mcpWorkerPodReaderStub) GetPod(
	_ context.Context,
	_ string,
) (*podDomain.Pod, error) {
	return stub.pod, stub.err
}
