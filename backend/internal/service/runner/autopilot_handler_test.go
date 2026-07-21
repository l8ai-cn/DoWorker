package runner

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newAutopilotTestCoordinator builds a minimal PodCoordinator wired only with
// the autopilot repo + logger — the autopilot callback handlers depend on
// nothing else (no connection manager / router / heartbeat).
func newAutopilotTestCoordinator(db *gorm.DB) *PodCoordinator {
	return &PodCoordinator{
		autopilotRepo: infra.NewAutopilotRepository(db),
		logger:        newTestLogger(),
	}
}

func seedAutopilotController(t *testing.T, db *gorm.DB, key string) *agentpod.AutopilotController {
	t.Helper()
	c := &agentpod.AutopilotController{
		AutopilotControllerKey: key,
		PodKey:                 "pod-target",
		PodID:                  10,
		RunnerID:               7,
		OrganizationID:         1,
		Phase:                  agentpod.AutopilotPhaseInitializing,
		MaxIterations:          10,
		CircuitBreakerState:    agentpod.CircuitBreakerClosed,
	}
	require.NoError(t, infra.NewAutopilotRepository(db).Create(context.Background(), c))
	return c
}

func fetchController(t *testing.T, db *gorm.DB, key string) *agentpod.AutopilotController {
	t.Helper()
	c, err := infra.NewAutopilotRepository(db).GetByKey(context.Background(), key)
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

// ==================== Iteration handler ====================

// TestHandleAutopilotIteration_PersistsAndUpdates also guards BUG-1: a wrong
// autopilot_iterations column name would make CreateIteration fail here.
func TestHandleAutopilotIteration_PersistsAndUpdates(t *testing.T) {
	db := setupTestDB(t)
	pc := newAutopilotTestCoordinator(db)
	ctrl := seedAutopilotController(t, db, "autopilot-it-1")

	var iterCb, statusCb bool
	pc.onAutopilotIterationChange = func(key string, iter int32, phase, summary string, files []string, dur int64) {
		iterCb = true
		assert.Equal(t, "autopilot-it-1", key)
		assert.Equal(t, int32(3), iter)
		assert.Equal(t, []string{"a.go", "b.go"}, files)
	}
	pc.onAutopilotStatusChange = func(key, podKey, phase string, iter, max int32, cbState, cbReason string, takeover bool) {
		statusCb = true
	}

	pc.handleAutopilotIteration(7, &runnerv1.AutopilotIterationEvent{
		AutopilotKey: "autopilot-it-1",
		Iteration:    3,
		Phase:        "action_sent",
		Summary:      "did something",
		FilesChanged: []string{"a.go", "b.go"},
		DurationMs:   1500,
	})

	iters, err := infra.NewAutopilotRepository(db).ListIterations(context.Background(), ctrl.ID)
	require.NoError(t, err)
	require.Len(t, iters, 1)
	assert.Equal(t, int32(3), iters[0].Iteration)
	assert.Equal(t, "action_sent", iters[0].Phase)
	require.NotNil(t, iters[0].Summary)
	assert.Equal(t, "did something", *iters[0].Summary)

	assert.Equal(t, int32(3), fetchController(t, db, "autopilot-it-1").CurrentIteration)
	assert.True(t, iterCb, "iteration callback should fire")
	assert.True(t, statusCb, "status callback should fire")
}

// TestHandleAutopilotIteration_NilController_NoPanic guards BUG-2: GetByKey
// returns (nil, nil) for an absent record; the handler must not deref rp.
func TestHandleAutopilotIteration_NilController_NoPanic(t *testing.T) {
	db := setupTestDB(t)
	pc := newAutopilotTestCoordinator(db)

	require.NotPanics(t, func() {
		pc.handleAutopilotIteration(7, &runnerv1.AutopilotIterationEvent{
			AutopilotKey: "does-not-exist",
			Iteration:    1,
		})
	})
}

// ==================== Status handler ====================

func TestHandleAutopilotControllerStatus_UpdatesPhase(t *testing.T) {
	db := setupTestDB(t)
	pc := newAutopilotTestCoordinator(db)
	seedAutopilotController(t, db, "autopilot-st-1")

	var cbPhase string
	pc.onAutopilotStatusChange = func(key, podKey, phase string, iter, max int32, cbState, cbReason string, takeover bool) {
		cbPhase = phase
	}

	pc.handleAutopilotControllerStatus(7, &runnerv1.AutopilotStatusEvent{
		AutopilotKey: "autopilot-st-1",
		PodKey:       "pod-target",
		Status: &runnerv1.AutopilotStatus{
			Phase:               agentpod.AutopilotPhaseRunning,
			CurrentIteration:    2,
			MaxIterations:       10,
			CircuitBreakerState: agentpod.CircuitBreakerClosed,
		},
	})

	updated := fetchController(t, db, "autopilot-st-1")
	assert.Equal(t, agentpod.AutopilotPhaseRunning, updated.Phase)
	assert.Equal(t, int32(2), updated.CurrentIteration)
	assert.Equal(t, agentpod.AutopilotPhaseRunning, cbPhase)
}

func TestHandleAutopilotControllerStatus_WaitingApprovalSetsTimestamp(t *testing.T) {
	db := setupTestDB(t)
	pc := newAutopilotTestCoordinator(db)
	seedAutopilotController(t, db, "autopilot-st-2")

	pc.handleAutopilotControllerStatus(7, &runnerv1.AutopilotStatusEvent{
		AutopilotKey: "autopilot-st-2",
		Status: &runnerv1.AutopilotStatus{
			Phase:        agentpod.AutopilotPhaseWaitingApproval,
			MaxIterations: 10,
		},
	})

	updated := fetchController(t, db, "autopilot-st-2")
	assert.Equal(t, agentpod.AutopilotPhaseWaitingApproval, updated.Phase)
	assert.NotNil(t, updated.ApprovalRequestAt, "approval_request_at should be set")
}

func TestHandleAutopilotControllerStatus_NilStatus_NoUpdate(t *testing.T) {
	db := setupTestDB(t)
	pc := newAutopilotTestCoordinator(db)
	seedAutopilotController(t, db, "autopilot-st-3")

	called := false
	pc.onAutopilotStatusChange = func(key, podKey, phase string, iter, max int32, cbState, cbReason string, takeover bool) {
		called = true
	}

	pc.handleAutopilotControllerStatus(7, &runnerv1.AutopilotStatusEvent{
		AutopilotKey: "autopilot-st-3",
		Status:       nil,
	})

	// Phase untouched, callback not fired.
	assert.Equal(t, agentpod.AutopilotPhaseInitializing, fetchController(t, db, "autopilot-st-3").Phase)
	assert.False(t, called)
}

// ==================== Created handler ====================

func TestHandleAutopilotControllerCreated_SetsRunning(t *testing.T) {
	db := setupTestDB(t)
	pc := newAutopilotTestCoordinator(db)
	seedAutopilotController(t, db, "autopilot-cr-1")

	var cbPhase string
	pc.onAutopilotStatusChange = func(key, podKey, phase string, iter, max int32, cbState, cbReason string, takeover bool) {
		cbPhase = phase
	}

	pc.handleAutopilotControllerCreated(7, &runnerv1.AutopilotCreatedEvent{
		AutopilotKey: "autopilot-cr-1",
		PodKey:       "pod-target",
	})

	updated := fetchController(t, db, "autopilot-cr-1")
	assert.Equal(t, agentpod.AutopilotPhaseRunning, updated.Phase)
	assert.NotNil(t, updated.StartedAt)
	assert.Equal(t, agentpod.AutopilotPhaseRunning, cbPhase)
}

// TestHandleAutopilotControllerCreated_NilController_NoPanic guards BUG-2 in
// the created handler: the onAutopilotStatusChange block dereferences rp.
func TestHandleAutopilotControllerCreated_NilController_NoPanic(t *testing.T) {
	db := setupTestDB(t)
	pc := newAutopilotTestCoordinator(db)

	called := false
	pc.onAutopilotStatusChange = func(key, podKey, phase string, iter, max int32, cbState, cbReason string, takeover bool) {
		called = true
	}

	require.NotPanics(t, func() {
		pc.handleAutopilotControllerCreated(7, &runnerv1.AutopilotCreatedEvent{
			AutopilotKey: "does-not-exist",
			PodKey:       "pod-x",
		})
	})
	assert.False(t, called, "callback must not fire when controller is absent")
}

// ==================== Terminated handler ====================

func TestHandleAutopilotControllerTerminated_ReasonMapping(t *testing.T) {
	cases := []struct {
		reason    string
		wantPhase string
	}{
		{"completed", agentpod.AutopilotPhaseCompleted},
		{"failed", agentpod.AutopilotPhaseFailed},
		{"max_iterations", agentpod.AutopilotPhaseMaxIterations},
		{"", agentpod.AutopilotPhaseStopped},
		{"unrecognized", agentpod.AutopilotPhaseStopped},
	}
	for _, tc := range cases {
		t.Run(tc.reason, func(t *testing.T) {
			db := setupTestDB(t)
			pc := newAutopilotTestCoordinator(db)
			key := "autopilot-tm-" + tc.wantPhase + tc.reason
			seedAutopilotController(t, db, key)

			pc.handleAutopilotControllerTerminated(7, &runnerv1.AutopilotTerminatedEvent{
				AutopilotKey: key,
				Reason:       tc.reason,
			})

			updated := fetchController(t, db, key)
			assert.Equal(t, tc.wantPhase, updated.Phase)
			assert.NotNil(t, updated.CompletedAt, "completed_at should be set on terminate")
		})
	}
}

// TestHandleAutopilotControllerTerminated_NilController_NoPanic guards BUG-2 in
// the terminated handler: the onAutopilotStatusChange block derefs rp.PodKey.
func TestHandleAutopilotControllerTerminated_NilController_NoPanic(t *testing.T) {
	db := setupTestDB(t)
	pc := newAutopilotTestCoordinator(db)

	called := false
	pc.onAutopilotStatusChange = func(key, podKey, phase string, iter, max int32, cbState, cbReason string, takeover bool) {
		called = true
	}

	require.NotPanics(t, func() {
		pc.handleAutopilotControllerTerminated(7, &runnerv1.AutopilotTerminatedEvent{
			AutopilotKey: "does-not-exist",
			Reason:       "completed",
		})
	})
	assert.False(t, called, "callback must not fire when controller is absent")
}
