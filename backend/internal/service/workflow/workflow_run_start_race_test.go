package workflow

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	workflowDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/stretchr/testify/require"
)

func TestSetRunPodKeyRejectsCancelledRun(t *testing.T) {
	env := setupWorkflowTest(t)
	result, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID:  env.workflow.ID,
		TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)
	require.NoError(t, env.orchestrator.MarkRunCancelled(
		env.ctx,
		result.Run.ID,
		"cancelled during pod creation",
	))

	err = env.orchestrator.SetRunPodKey(env.ctx, result.Run.ID, "late-pod", "")

	require.Error(t, err)
	run, getErr := env.runSvc.GetByID(env.ctx, result.Run.ID)
	require.NoError(t, getErr)
	require.Nil(t, run.PodKey)
	require.Equal(t, workflowDomain.RunStatusCancelled, run.Status)
}

func TestStartRunTerminatesPodWhenCancellationWinsBinding(t *testing.T) {
	env := setupWorkflowTest(t, func(workflow *workflowDomain.Workflow) {
		workflow.ExecutionMode = workflowDomain.ExecutionModeAutopilot
	})
	creator := newBlockingWorkflowPodCreator(1)
	autopilot := &countingWorkflowAutopilotStarter{}
	env.orchestrator.SetPodDependencies(
		creator,
		autopilot,
		env.podTerm,
		nil,
		nil,
	)
	result, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID:  env.workflow.ID,
		TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)

	done := make(chan struct{})
	go func() {
		defer close(done)
		env.orchestrator.StartRun(
			env.ctx,
			result.Workflow,
			result.Run,
			1,
		)
	}()
	<-creator.allEntered
	require.NoError(t, env.orchestrator.MarkRunCancelled(
		env.ctx,
		result.Run.ID,
		"cancelled during pod creation",
	))
	close(creator.release)
	<-done

	require.Equal(t, []string{"workflow-pod-1"}, env.podTerm.getTerminatedKeys())
	require.Zero(t, autopilot.createCalls)
	run, getErr := env.runSvc.GetByID(env.ctx, result.Run.ID)
	require.NoError(t, getErr)
	require.Nil(t, run.PodKey)
	require.Equal(t, workflowDomain.RunStatusCancelled, run.Status)
}

func TestConcurrentStartRunTerminatesBindingLoser(t *testing.T) {
	env := setupWorkflowTest(t)
	creator := newBlockingWorkflowPodCreator(2)
	env.orchestrator.SetPodDependencies(
		creator,
		nil,
		env.podTerm,
		nil,
		nil,
	)
	result, err := env.orchestrator.TriggerRun(env.ctx, &TriggerRunRequest{
		WorkflowID:  env.workflow.ID,
		TriggerType: workflowDomain.RunTriggerManual,
	})
	require.NoError(t, err)

	var starts sync.WaitGroup
	starts.Add(2)
	for range 2 {
		go func() {
			defer starts.Done()
			env.orchestrator.StartRun(
				env.ctx,
				result.Workflow,
				result.Run,
				1,
			)
		}()
	}
	<-creator.allEntered
	close(creator.release)
	starts.Wait()

	run, getErr := env.runSvc.GetByID(env.ctx, result.Run.ID)
	require.NoError(t, getErr)
	require.NotNil(t, run.PodKey)
	require.Contains(t, []string{"workflow-pod-1", "workflow-pod-2"}, *run.PodKey)
	terminated := env.podTerm.getTerminatedKeys()
	require.Len(t, terminated, 1)
	require.NotEqual(t, *run.PodKey, terminated[0])
}

type blockingWorkflowPodCreator struct {
	mu         sync.Mutex
	calls      int
	wantCalls  int
	allEntered chan struct{}
	release    chan struct{}
}

func newBlockingWorkflowPodCreator(wantCalls int) *blockingWorkflowPodCreator {
	return &blockingWorkflowPodCreator{
		wantCalls:  wantCalls,
		allEntered: make(chan struct{}),
		release:    make(chan struct{}),
	}
}

func (c *blockingWorkflowPodCreator) CreatePod(
	context.Context,
	*agentpodSvc.OrchestrateCreatePodRequest,
) (*agentpodSvc.OrchestrateCreatePodResult, error) {
	c.mu.Lock()
	c.calls++
	call := c.calls
	if c.calls == c.wantCalls {
		close(c.allEntered)
	}
	c.mu.Unlock()
	<-c.release
	return &agentpodSvc.OrchestrateCreatePodResult{Pod: &agentpod.Pod{
		ID:             int64(call),
		OrganizationID: 1,
		RunnerID:       1,
		PodKey:         fmt.Sprintf("workflow-pod-%d", call),
	}}, nil
}

type countingWorkflowAutopilotStarter struct {
	createCalls int
}

func (s *countingWorkflowAutopilotStarter) CreateAndStart(
	context.Context,
	*agentpodSvc.CreateAndStartRequest,
) (*agentpod.AutopilotController, error) {
	s.createCalls++
	return &agentpod.AutopilotController{
		AutopilotControllerKey: "workflow-autopilot-1",
	}, nil
}

func (*countingWorkflowAutopilotStarter) GetApprovalTimedOut(
	context.Context,
	[]int64,
) ([]*agentpod.AutopilotController, error) {
	return nil, nil
}

func (*countingWorkflowAutopilotStarter) UpdateAutopilotControllerStatus(
	context.Context,
	string,
	map[string]interface{},
) error {
	return nil
}
