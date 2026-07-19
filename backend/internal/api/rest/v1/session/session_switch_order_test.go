package sessionapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRebuildSessionPodKeepsOldPodUntilNewBindingPersists(t *testing.T) {
	deps, row, oldPod, lifecycle, orchestrator := setupSessionSwitchOrderTest(t)
	orchestrator.beforeCreate = func() {
		assert.Empty(t, lifecycle.terminated)
	}
	orchestrator.beforeDispatch = func() {
		assert.Empty(t, lifecycle.terminated)
		persisted, err := deps.Sessions.Get(context.Background(), row.ID)
		require.NoError(t, err)
		assert.Equal(t, "new-pod", persisted.PodKey)
	}

	created, err := deps.rebuildSessionPod(sessionSwitchTestContext(), row, oldPod, "codex-cli", "dev-org", sessionSwitchPlan())

	require.NoError(t, err)
	assert.Equal(t, "new-pod", created.PodKey)
	assert.Equal(t, []string{"old-pod"}, lifecycle.terminated)
	require.NotNil(t, orchestrator.request)
	assert.True(t, orchestrator.request.DeferRunnerDispatch)
	assert.Empty(t, orchestrator.request.ResumeExternalSessionID)
	assert.Equal(t, 1, orchestrator.dispatches)
	persisted, err := deps.Sessions.Get(context.Background(), row.ID)
	require.NoError(t, err)
	assert.Equal(t, "new-pod", persisted.PodKey)
}

func TestRebuildSessionPodRestoresOldBindingWhenDeferredDispatchFails(t *testing.T) {
	deps, row, oldPod, lifecycle, orchestrator := setupSessionSwitchOrderTest(t)
	dispatchErr := errors.New("runner dispatch failed")
	orchestrator.dispatchErr = dispatchErr

	created, err := deps.rebuildSessionPod(sessionSwitchTestContext(), row, oldPod, "codex-cli", "dev-org", sessionSwitchPlan())

	assert.Nil(t, created)
	assert.ErrorIs(t, err, dispatchErr)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	persisted, getErr := deps.Sessions.Get(context.Background(), row.ID)
	require.NoError(t, getErr)
	assert.Equal(t, "old-pod", persisted.PodKey)
	assert.Equal(t, "claude-code", persisted.AgentSlug)
}

func TestRebuildSessionPodTerminatesNewPodWhenBindingUpdateFails(t *testing.T) {
	deps, row, oldPod, lifecycle, _ := setupSessionSwitchOrderTest(t)
	require.NoError(t, deps.SessionsDB.Exec(`
		CREATE TRIGGER fail_session_binding_update
		BEFORE UPDATE ON agent_sessions
		BEGIN
			SELECT RAISE(FAIL, 'session update blocked');
		END
	`).Error)

	created, err := deps.rebuildSessionPod(sessionSwitchTestContext(), row, oldPod, "codex-cli", "dev-org", sessionSwitchPlan())

	require.Error(t, err)
	assert.Nil(t, created)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
}

func TestRebuildSessionPodReturnsCompensationFailure(t *testing.T) {
	deps, row, oldPod, lifecycle, _ := setupSessionSwitchOrderTest(t)
	require.NoError(t, deps.SessionsDB.Exec(`
		CREATE TRIGGER fail_session_binding_update
		BEFORE UPDATE ON agent_sessions
		BEGIN
			SELECT RAISE(FAIL, 'session update blocked');
		END
	`).Error)
	terminationErr := errors.New("termination unavailable")
	lifecycle.terminateErr = terminationErr

	created, err := deps.rebuildSessionPod(sessionSwitchTestContext(), row, oldPod, "codex-cli", "dev-org", sessionSwitchPlan())

	assert.Nil(t, created)
	assert.ErrorIs(t, err, terminationErr)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
}

func TestRebuildSessionPodReportsOldPodTerminationFailureAfterBinding(t *testing.T) {
	deps, row, oldPod, lifecycle, _ := setupSessionSwitchOrderTest(t)
	terminationErr := errors.New("termination unavailable")
	lifecycle.terminateErrors = map[string]error{"old-pod": terminationErr}

	created, err := deps.rebuildSessionPod(sessionSwitchTestContext(), row, oldPod, "codex-cli", "dev-org", sessionSwitchPlan())

	assert.Nil(t, created)
	assert.ErrorIs(t, err, terminationErr)
	assert.Equal(t, []string{"old-pod"}, lifecycle.terminated)
	persisted, getErr := deps.Sessions.Get(context.Background(), row.ID)
	require.NoError(t, getErr)
	assert.Equal(t, "new-pod", persisted.PodKey)
	assert.Equal(t, "codex-cli", persisted.AgentSlug)
}

func TestWriteSessionPodErrorPrioritizesCompensationFailure(t *testing.T) {
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest("POST", "/v1/sessions/conv_switch/switch-agent", nil)

	writeSessionPodError(
		ctx,
		errors.Join(errSessionCompensationFailed, runnerservice.ErrRunnerOffline),
	)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.JSONEq(t, `{
		"error":"session cleanup failed",
		"code":"session_compensation_failed"
	}`, response.Body.String())
}

func TestSwitchAgentRejectsSameAgentWorkerConfigChange(t *testing.T) {
	deps, _, _, _, orchestrator := setupSessionSwitchOrderTest(t)

	response := sessionSwitchRequest(
		deps.Deps,
		`{"agent_id":"claude-code","worker_spec":{
			"options_revision":"rev-1",
			"runtime_image_id":4,
			"placement_policy":"explicit",
			"compute_target_id":8,
			"deployment_mode":"pooled",
			"resource_profile_id":9
		},"automation_level":"autonomous"}`,
	)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{
		"error":"same-agent operation cannot change worker configuration",
		"code":"validation_failed"
	}`, response.Body.String())
	assert.Nil(t, orchestrator.request)
}

type sessionSwitchTestDeps struct {
	*Deps
	SessionsDB *gorm.DB
}

func setupSessionSwitchOrderTest(
	t *testing.T,
) (*sessionSwitchTestDeps, *sessionDomain.Session, *podDomain.Pod, *recordingSessionPodLifecycle, *switchPodOrchestrator) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/session-switch.db"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&sessionDomain.Session{}))
	sessions := sessionsvc.NewService(db)
	row := &sessionDomain.Session{
		ID:             "conv_switch",
		OrganizationID: 11,
		UserID:         7,
		PodKey:         "old-pod",
		AgentSlug:      "claude-code",
		Status:         "idle",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	require.NoError(t, sessions.Create(context.Background(), row))
	oldPod := &podDomain.Pod{
		PodKey: "old-pod", RunnerID: 3, OrganizationID: 11,
		CreatedByID: 7, AgentSlug: "claude-code", Status: podDomain.StatusCompleted,
		ExternalSessionID: stringPtr("provider-session"),
	}
	oldPod.WorkerSpecSnapshotID = int64Ptr(91)
	lifecycle := &recordingSessionPodLifecycle{}
	orchestrator := &switchPodOrchestrator{
		result: &agentpod.OrchestrateCreatePodResult{
			Pod: &podDomain.Pod{
				PodKey: "new-pod", RunnerID: 3, OrganizationID: 11,
				CreatedByID: 7, AgentSlug: "codex-cli", Status: podDomain.StatusInitializing,
			},
		},
	}
	return &sessionSwitchTestDeps{
		Deps: &Deps{
			Sessions:        sessions,
			PodOrchestrator: orchestrator,
			PodCoordinator:  lifecycle,
			WorkerCreation: &recordingSessionWorkerDraftFactory{
				draft: sessionTestWorkerDraft(t, "codex-cli"),
			},
		},
		SessionsDB: db,
	}, row, oldPod, lifecycle, orchestrator
}

func sessionSwitchPlan() sessionWorkerPlanInput {
	return sessionWorkerPlanInput{
		WorkerSpec:      validSessionWorkerSpecBody(),
		WorkerTypeSlug:  "codex-cli",
		AgentfileLayer:  acpAgentfileLayer(),
		AutomationLevel: "autonomous",
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}

func sessionSwitchTestContext() *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/v1/sessions/conv_switch/switch-agent", nil)
	return ctx
}

func sessionSwitchRequest(deps *Deps, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/v1/sessions/conv_switch/switch-agent",
		bytes.NewBufferString(body),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: "conv_switch"}}
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: 11, UserID: 7})
	deps.handleSwitchAgent(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

type switchPodOrchestrator struct {
	beforeCreate   func()
	beforeDispatch func()
	request        *agentpod.OrchestrateCreatePodRequest
	result         *agentpod.OrchestrateCreatePodResult
	dispatches     int
	dispatchErr    error
}

func (o *switchPodOrchestrator) CreatePod(
	_ context.Context,
	req *agentpod.OrchestrateCreatePodRequest,
) (*agentpod.OrchestrateCreatePodResult, error) {
	if o.beforeCreate != nil {
		o.beforeCreate()
	}
	o.request = req
	return o.result, nil
}

func (o *switchPodOrchestrator) DispatchDeferredPod(
	_ context.Context,
	_ *agentpod.OrchestrateCreatePodRequest,
	result *agentpod.OrchestrateCreatePodResult,
) (*agentpod.OrchestrateCreatePodResult, error) {
	o.dispatches++
	if o.beforeDispatch != nil {
		o.beforeDispatch()
	}
	if o.dispatchErr != nil {
		return nil, o.dispatchErr
	}
	return result, nil
}

type recordingSessionPodLifecycle struct {
	terminated      []string
	contextErrors   []error
	terminateErr    error
	terminateErrors map[string]error
}

func (l *recordingSessionPodLifecycle) TerminatePod(ctx context.Context, podKey string) error {
	l.terminated = append(l.terminated, podKey)
	l.contextErrors = append(l.contextErrors, ctx.Err())
	if err := l.terminateErrors[podKey]; err != nil {
		return err
	}
	return l.terminateErr
}

func (l *recordingSessionPodLifecycle) TerminatePodDeleteBranch(ctx context.Context, podKey string) error {
	l.terminated = append(l.terminated, podKey)
	l.contextErrors = append(l.contextErrors, ctx.Err())
	if err := l.terminateErrors[podKey]; err != nil {
		return err
	}
	return l.terminateErr
}
