package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	agentsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
)

func postQuickTask(t *testing.T, h *PodHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/quick-tasks", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenant", &middleware.TenantContext{
		OrganizationID:   42,
		OrganizationSlug: "team-alpha",
		UserID:           7,
	})
	h.CreateQuickTask(c)
	return w
}

func TestQuickTask_InvalidPlanID_400(t *testing.T) {
	for _, body := range []string{
		`{}`,
		`{"plan_id":"not-a-uuid"}`,
		`{"plan_id":"00000000-0000-0000-0000-000000000000"}`,
	} {
		w := postQuickTask(t, &PodHandler{}, body)
		assert.Equal(t, http.StatusBadRequest, w.Code, "body=%s", body)
	}
}

func TestQuickTask_AppliesWorkerPlanAndReturnsQueueState(t *testing.T) {
	applier := &recordingQuickTaskPlanApplier{
		result: workerplanner.AppliedWorker{
			PodKey:   "7-standalone-12345678",
			RunnerID: 11,
		},
	}
	authorizer := &recordingQuickTaskPlanAuthorizer{}
	queue := &quickTaskQueueStub{
		position:  3,
		expiresAt: time.Date(2026, time.July, 17, 8, 30, 0, 0, time.UTC),
	}
	handler := &PodHandler{
		quickTaskPlanAuthorizer: authorizer,
		quickTaskPlanApplier:    applier,
		quickTaskPodReader: &quickTaskPodReaderStub{pod: &podDomain.Pod{
			PodKey: applier.result.PodKey,
			Status: podDomain.StatusQueued,
		}},
		pendingQueue: queue,
	}

	w := postQuickTask(
		t,
		handler,
		`{"plan_id":"11111111-1111-4111-8111-111111111111"}`,
	)

	require.Equal(t, http.StatusAccepted, w.Code)
	assert.Equal(t, control.Scope{
		OrganizationID:   42,
		OrganizationSlug: "team-alpha",
		ActorID:          7,
	}, applier.scope)
	assert.Equal(t, "11111111-1111-4111-8111-111111111111", applier.planID)
	assert.Equal(t, applier.scope, authorizer.scope)
	assert.Equal(t, applier.planID, authorizer.planID)
	var response QuickTaskResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, applier.result.PodKey, response.PodKey)
	assert.Equal(t, podDomain.StatusQueued, response.Status)
	assert.Equal(t, 3, response.QueuePosition)
	assert.Equal(t, "2026-07-17T08:30:00Z", response.ExpiresAt)
	assert.Equal(t, int64(11), queue.runnerID)
	assert.Equal(t, applier.result.PodKey, queue.podKey)
}

func TestQuickTask_PlanReplayReturnsCurrentPodStatus(t *testing.T) {
	applier := &recordingQuickTaskPlanApplier{
		result: workerplanner.AppliedWorker{
			PodKey:   "7-standalone-12345678",
			RunnerID: 11,
		},
	}
	handler := &PodHandler{
		quickTaskPlanAuthorizer: &recordingQuickTaskPlanAuthorizer{},
		quickTaskPlanApplier:    applier,
		quickTaskPodReader: &quickTaskPodReaderStub{pod: &podDomain.Pod{
			PodKey: applier.result.PodKey,
			Status: podDomain.StatusRunning,
		}},
	}

	w := postQuickTask(
		t,
		handler,
		`{"plan_id":"11111111-1111-4111-8111-111111111111"}`,
	)

	require.Equal(t, http.StatusAccepted, w.Code)
	var response QuickTaskResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, podDomain.StatusRunning, response.Status)
}

func TestQuickTask_AuthorizationFailureDoesNotApplyPlan(t *testing.T) {
	applier := &recordingQuickTaskPlanApplier{}
	handler := &PodHandler{
		quickTaskPlanAuthorizer: &recordingQuickTaskPlanAuthorizer{
			err: controlservice.ErrForbidden,
		},
		quickTaskPlanApplier: applier,
		quickTaskPodReader:   &quickTaskPodReaderStub{},
	}

	w := postQuickTask(
		t,
		handler,
		`{"plan_id":"11111111-1111-4111-8111-111111111111"}`,
	)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, applier.planID)
}

func TestQuickTask_UnconfiguredApplyService_503(t *testing.T) {
	w := postQuickTask(
		t,
		&PodHandler{},
		`{"plan_id":"11111111-1111-4111-8111-111111111111"}`,
	)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestQuickTask_InvalidJSON_400(t *testing.T) {
	w := postQuickTask(t, &PodHandler{}, `{"plan_id":`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMapQuickTaskError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantErr  string
	}{
		{"queue full -> 429", podDomain.ErrQueueFull, http.StatusTooManyRequests, "QUEUE_FULL"},
		{"no runner -> 422", agentsvc.ErrNoAvailableRunner, http.StatusUnprocessableEntity, "NO_RUNNER_FOR_AGENT"},
		{"invalid plan -> 400", control.ErrInvalid, http.StatusBadRequest, "WORKER_PLAN_INVALID"},
		{"missing plan -> 404", control.ErrNotFound, http.StatusNotFound, "WORKER_PLAN_NOT_FOUND"},
		{"stale plan -> 409", control.ErrStale, http.StatusConflict, "WORKER_PLAN_STATE_CHANGED"},
		{"consumed plan -> 409", control.ErrConsumed, http.StatusConflict, "WORKER_PLAN_STATE_CHANGED"},
		{"apply unavailable -> 503", controlservice.ErrUnavailable, http.StatusServiceUnavailable, "WORKER_APPLY_UNAVAILABLE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			mapQuickTaskError(c, tt.err)
			assert.Equal(t, tt.wantCode, w.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, tt.wantErr, resp["code"])
		})
	}
}

type recordingQuickTaskPlanApplier struct {
	scope  control.Scope
	planID string
	result workerplanner.AppliedWorker
	err    error
}

type recordingQuickTaskPlanAuthorizer struct {
	scope  control.Scope
	planID string
	err    error
}

func (stub *recordingQuickTaskPlanAuthorizer) AuthorizeApply(
	_ context.Context,
	scope control.Scope,
	planID string,
) error {
	stub.scope = scope
	stub.planID = planID
	return stub.err
}

func (stub *recordingQuickTaskPlanApplier) Apply(
	_ context.Context,
	scope control.Scope,
	planID string,
) (workerplanner.AppliedWorker, error) {
	stub.scope = scope
	stub.planID = planID
	return stub.result, stub.err
}

type quickTaskQueueStub struct {
	position  int
	expiresAt time.Time
	runnerID  int64
	podKey    string
}

type quickTaskPodReaderStub struct {
	pod *podDomain.Pod
	err error
}

func (stub *quickTaskPodReaderStub) GetPod(
	_ context.Context,
	_ string,
) (*podDomain.Pod, error) {
	return stub.pod, stub.err
}

func (stub *quickTaskQueueStub) QueuePosition(
	_ context.Context,
	runnerID int64,
	podKey string,
) (int, error) {
	stub.runnerID = runnerID
	stub.podKey = podKey
	return stub.position, nil
}

func (stub *quickTaskQueueStub) GetCreatePodExpiry(
	_ context.Context,
	_ string,
) (time.Time, error) {
	return stub.expiresAt, nil
}

var _ pendingQueueReader = (*quickTaskQueueStub)(nil)
var _ quickTaskPodReader = (*quickTaskPodReaderStub)(nil)
var _ QuickTaskPlanAuthorizer = (*recordingQuickTaskPlanAuthorizer)(nil)
