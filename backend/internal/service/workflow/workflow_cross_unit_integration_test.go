package workflow

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	ticketSvc "github.com/anthropics/agentsmesh/backend/internal/service/ticket"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	eventsv1 "github.com/anthropics/agentsmesh/proto/gen/go/events/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// crossUnitEnv holds objects for cross-unit integration tests that exercise
// the full EventBus -> WorkflowOrchestrator event-driven chain.
type crossUnitEnv struct {
	orchestrator *WorkflowOrchestrator
	workflowSvc  *WorkflowService
	runSvc       *WorkflowRunService
	eventBus     *eventbus.EventBus
	workflow     *workflowDomain.Workflow
	ctx          context.Context
}

// setupCrossUnit creates a real EventBus wired to a real DB-backed WorkflowOrchestrator,
// replicating the production subscription pattern from setupLoopEventSubscriptions.
func setupCrossUnit(t *testing.T, opts ...func(*workflowDomain.Workflow)) crossUnitEnv {
	t.Helper()
	db := testkit.SetupTestDB(t)

	// SQLite `:memory:` creates a separate DB per connection. Force single connection
	// so async EventBus goroutines share the same in-memory DB.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	workflowRepo := infra.NewWorkflowRepository(db)
	runRepo := infra.NewWorkflowRunRepository(db)
	workflowSvc := NewWorkflowService(workflowRepo)
	runSvc := NewWorkflowRunService(runRepo)
	ctx := context.Background()

	eb := eventbus.NewEventBus(nil, slog.Default())
	t.Cleanup(func() { eb.Close() })

	orchestrator := NewWorkflowOrchestrator(workflowSvc, runSvc, eb, slog.Default())
	orchestrator.SetPodDependencies(nil, nil, &mockPodTerminatorForWorkflow{}, nil, nil)

	// Wire subscriptions exactly like production (cmd/server/eventbus_loop.go)
	wireLoopSubscriptions(eb, orchestrator)

	slug := fmt.Sprintf("cu-test-%d", time.Now().UnixNano()%100000)
	workflow, err := workflowSvc.Create(ctx, &CreateWorkflowRequest{
		OrganizationID: 1,
		CreatedByID:    1,
		Name:           "Cross Unit Test",
		Slug:           slug,
		AgentSlug:      "claude",
		PromptTemplate: "do the task",
		ExecutionMode:  workflowDomain.ExecutionModeDirect,
		TimeoutMinutes: 30,
	})
	require.NoError(t, err)

	resourceID := workflow.ID + 3000
	resourceRevision := int64(1)
	snapshotID := workflow.ID + 4000
	workflow.OrchestrationResourceID = &resourceID
	workflow.OrchestrationResourceRevision = &resourceRevision
	workflow.WorkerSpecSnapshotID = &snapshotID
	for _, opt := range opts {
		opt(workflow)
	}
	updates := map[string]interface{}{
		"sandbox_strategy":                workflow.SandboxStrategy,
		"session_persistence":             workflow.SessionPersistence,
		"execution_mode":                  workflow.ExecutionMode,
		"timeout_minutes":                 workflow.TimeoutMinutes,
		"ticket_id":                       workflow.TicketID,
		"orchestration_resource_id":       resourceID,
		"orchestration_resource_revision": resourceRevision,
		"worker_spec_snapshot_id":         snapshotID,
	}
	require.NoError(t, workflowRepo.Update(ctx, workflow.ID, updates))
	workflow, err = workflowSvc.GetByID(ctx, workflow.ID)
	require.NoError(t, err)

	return crossUnitEnv{
		orchestrator: orchestrator,
		workflowSvc:  workflowSvc,
		runSvc:       runSvc,
		eventBus:     eb,
		workflow:     workflow,
		ctx:          ctx,
	}
}

// wireLoopSubscriptions replicates production subscription wiring.
func wireLoopSubscriptions(eb *eventbus.EventBus, orchestrator *WorkflowOrchestrator) {
	eb.Subscribe(eventbus.EventPodTerminated, func(event *eventbus.Event) {
		var data eventsv1.PodStatusChangedEventData
		if err := protojson.Unmarshal(event.Data, &data); err != nil {
			return
		}
		now := time.Now()
		orchestrator.HandlePodTerminated(context.Background(), data.PodKey, data.Status, &now)
	})

	eb.Subscribe(eventbus.EventPodStatusChanged, func(event *eventbus.Event) {
		var data eventsv1.PodStatusChangedEventData
		if err := protojson.Unmarshal(event.Data, &data); err != nil {
			return
		}
		switch data.Status {
		case podDomain.StatusCompleted, podDomain.StatusError:
			now := time.Now()
			orchestrator.HandlePodTerminated(context.Background(), data.PodKey, data.Status, &now)
		}
	})

	eb.Subscribe(eventbus.EventAutopilotStatusChanged, func(event *eventbus.Event) {
		var data eventsv1.AutopilotStatusChangedEventData
		if err := protojson.Unmarshal(event.Data, &data); err != nil {
			return
		}
		switch data.Phase {
		case podDomain.AutopilotPhaseCompleted, podDomain.AutopilotPhaseFailed, podDomain.AutopilotPhaseStopped:
			orchestrator.HandleAutopilotTerminated(context.Background(), data.AutopilotControllerKey, data.Phase)
		}
	})
}

// createRunWithPodKey triggers a run and associates a pod key.
func createRunWithPodKey(t *testing.T, env crossUnitEnv, podKey, autopilotKey string) *workflowDomain.WorkflowRun {
	t.Helper()
	result, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID:  env.workflow.ID,
		TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)
	require.False(t, result.Skipped)
	require.NoError(t, env.orchestrator.SetRunPodKey(env.ctx, result.Run.ID, podKey, autopilotKey))
	return result.Run
}

// publishPodTerminated publishes an EventPodTerminated event.
func publishPodTerminated(t *testing.T, eb *eventbus.EventBus, podKey, status string, orgID int64) {
	t.Helper()
	event, err := eventbus.NewEntityEvent(eventbus.EventPodTerminated, orgID, "pod", podKey,
		&eventsv1.PodStatusChangedEventData{PodKey: podKey, Status: status})
	require.NoError(t, err)
	require.NoError(t, eb.Publish(context.Background(), event))
}

// publishAutopilotStatusChanged publishes an EventAutopilotStatusChanged event.
func publishAutopilotStatusChanged(t *testing.T, eb *eventbus.EventBus, autopilotKey, phase, podKey string, orgID int64) {
	t.Helper()
	event, err := eventbus.NewEntityEvent(eventbus.EventAutopilotStatusChanged, orgID, "autopilot", autopilotKey,
		&eventsv1.AutopilotStatusChangedEventData{AutopilotControllerKey: autopilotKey, Phase: phase, PodKey: podKey})
	require.NoError(t, err)
	require.NoError(t, eb.Publish(context.Background(), event))
}

// waitForRunStatus polls the DB until the run reaches the expected status with FinishedAt set.
// We check FinishedAt because the SSOT resolver can derive a status before FinishRun persists it.
func waitForRunStatus(t *testing.T, runSvc *WorkflowRunService, runID int64, expected string) *workflowDomain.WorkflowRun {
	t.Helper()
	var run *workflowDomain.WorkflowRun
	require.Eventually(t, func() bool {
		r, err := runSvc.GetByID(context.Background(), runID)
		if err != nil {
			return false
		}
		run = r
		return r.Status == expected && r.FinishedAt != nil
	}, 3*time.Second, 50*time.Millisecond, "run %d did not reach status %q with FinishedAt", runID, expected)
	return run
}

// --- Test 1: EventBus PodTerminated -> Workflow Completion ---

func TestCrossUnit_EventBusPodTerminated_TriggersLoopCompletion(t *testing.T) {
	env := setupCrossUnit(t)
	podKey := fmt.Sprintf("cu-pod-%d", time.Now().UnixNano())

	run := createRunWithPodKey(t, env, podKey, "")

	// Publish EventPodTerminated with status=completed (like production)
	publishPodTerminated(t, env.eventBus, podKey, podDomain.StatusCompleted, 1)

	// Wait for async handler to process and update DB
	completed := waitForRunStatus(t, env.runSvc, run.ID, workflowDomain.RunStatusCompleted)
	assert.NotNil(t, completed.FinishedAt)

	// Verify workflow stats: total_runs and successful_runs incremented
	workflow, err := env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, workflow.TotalRuns)
	assert.Equal(t, 1, workflow.SuccessfulRuns)
	assert.Equal(t, 0, workflow.FailedRuns)
}

// --- Test 2: EventBus PodError -> Workflow Failure ---

func TestCrossUnit_EventBusPodError_TriggersLoopFailure(t *testing.T) {
	env := setupCrossUnit(t)
	podKey := fmt.Sprintf("cu-pod-err-%d", time.Now().UnixNano())

	run := createRunWithPodKey(t, env, podKey, "")

	// Publish EventPodTerminated with status=error
	publishPodTerminated(t, env.eventBus, podKey, podDomain.StatusError, 1)

	// Wait for async handler
	failed := waitForRunStatus(t, env.runSvc, run.ID, workflowDomain.RunStatusFailed)
	assert.NotNil(t, failed.FinishedAt)

	// Verify workflow stats: failed_runs incremented
	workflow, err := env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, workflow.TotalRuns)
	assert.Equal(t, 0, workflow.SuccessfulRuns)
	assert.Equal(t, 1, workflow.FailedRuns)
}

// --- Test 3: EventBus AutopilotStatusChanged -> Workflow Completion ---

func TestCrossUnit_EventBusAutopilotTerminated_TriggersLoopCompletion(t *testing.T) {
	env := setupCrossUnit(t, func(l *workflowDomain.Workflow) {
		l.ExecutionMode = workflowDomain.ExecutionModeAutopilot
	})

	podKey := fmt.Sprintf("cu-pod-ap-%d", time.Now().UnixNano())
	autopilotKey := fmt.Sprintf("cu-ap-%d", time.Now().UnixNano())

	run := createRunWithPodKey(t, env, podKey, autopilotKey)

	// Publish EventAutopilotStatusChanged with phase=completed
	publishAutopilotStatusChanged(t, env.eventBus, autopilotKey, podDomain.AutopilotPhaseCompleted, podKey, 1)

	// Wait for async handler
	completed := waitForRunStatus(t, env.runSvc, run.ID, workflowDomain.RunStatusCompleted)
	assert.NotNil(t, completed.FinishedAt)

	// Verify workflow stats
	workflow, err := env.workflowSvc.GetByID(env.ctx, env.workflow.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, workflow.TotalRuns)
	assert.Equal(t, 1, workflow.SuccessfulRuns)
}

// --- Test 4: Workflow Completion -> Posts Ticket Comment ---

func TestCrossUnit_LoopCompletion_PostsTicketComment(t *testing.T) {
	mockRepo := &mockTicketRepoForCrossUnit{}
	tktSvc := ticketSvc.NewService(mockRepo)

	ticketID := int64(42)
	env := setupCrossUnit(t, func(l *workflowDomain.Workflow) {
		l.TicketID = &ticketID
	})
	env.orchestrator.SetPodDependencies(nil, nil, &mockPodTerminatorForWorkflow{}, tktSvc, nil)

	podKey := fmt.Sprintf("cu-pod-tkt-%d", time.Now().UnixNano())
	run := createRunWithPodKey(t, env, podKey, "")

	// Publish EventPodTerminated with status=completed
	publishPodTerminated(t, env.eventBus, podKey, podDomain.StatusCompleted, 1)

	// Wait for run completion
	waitForRunStatus(t, env.runSvc, run.ID, workflowDomain.RunStatusCompleted)

	// postTicketComment is called via goroutine, wait for it
	require.Eventually(t, func() bool {
		return len(mockRepo.getComments()) > 0
	}, 3*time.Second, 50*time.Millisecond, "ticket comment was not posted")

	comments := mockRepo.getComments()
	require.Len(t, comments, 1)
	assert.Equal(t, ticketID, comments[0].TicketID)
	assert.Equal(t, int64(1), comments[0].UserID) // Workflow.CreatedByID
	assert.Contains(t, comments[0].Content, "Workflow Run #1")
	assert.Contains(t, comments[0].Content, "completed")
	// Success emoji
	assert.Contains(t, comments[0].Content, "✅") // checkmark emoji
}

// mockTicketRepoForCrossUnit is in mock_ticket_repo_test.go
