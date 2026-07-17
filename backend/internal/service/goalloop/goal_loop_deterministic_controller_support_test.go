package goalloop

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/goalloop"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func verificationFailureCycle(
	t *testing.T,
	service *Service,
	verification *recordingVerificationDispatcher,
	loop *domain.GoalLoop,
) error {
	t.Helper()
	require.NoError(t, handleCurrentAgentStatus(
		service, *loop.PodKey, agentpod.AgentStatusWaiting, time.Now(),
	))
	command := verification.commands[len(verification.commands)-1]
	return service.HandleVerificationResult(context.Background(), 7, &runnerv1.VerificationResultEvent{
		RequestId: command.GetRequestId(),
		PodKey:    *loop.PodKey,
		ExitCode:  1,
		Output:    "checkout test failed",
	})
}

func handleCurrentAgentStatus(
	service *Service,
	podKey, status string,
	eventAt time.Time,
) error {
	service.podLookup.(*goalLoopPodStore).pod.AgentStatus = status
	return service.HandlePodAgentStatus(context.Background(), podKey, status, eventAt)
}

func deterministicLoopService(
	maxIterations, noProgressLimit, sameErrorLimit int,
) (*domain.GoalLoop, *Service, *recordingVerificationDispatcher, *recordingPromptDispatcher) {
	podKey := "goal-loop-pod"
	loop := &domain.GoalLoop{
		ID: 1, OrganizationID: 1, Slug: "repair-checkout", Status: domain.StatusActive,
		PodKey: &podKey, VerificationCommand: "go test ./...", TimeoutMinutes: 60,
		MaxIterations: maxIterations, NoProgressLimit: noProgressLimit,
		SameErrorLimit: sameErrorLimit, EscalationPolicy: domain.EscalationPause,
	}
	verification := &recordingVerificationDispatcher{}
	prompts := &recordingPromptDispatcher{}
	service := NewService(newGoalLoopTestRepo(loop))
	pod := runningPod(podKey)
	pod.AgentStatus = agentpod.AgentStatusWaiting
	service.podLookup = &goalLoopPodStore{pod: pod}
	service.podTerminator = &goalLoopTerminator{}
	service.verificationSender = verification
	service.promptSender = prompts
	return loop, service, verification, prompts
}

type recordingVerificationDispatcher struct {
	commands []*runnerv1.RunVerificationCommand
}

func (d *recordingVerificationDispatcher) SendRunVerification(
	_ context.Context,
	_ int64,
	command *runnerv1.RunVerificationCommand,
) error {
	d.commands = append(d.commands, command)
	return nil
}

type promptCall struct {
	runnerID  int64
	podKey    string
	commandID string
	prompt    string
}

type recordingPromptDispatcher struct {
	mu       sync.Mutex
	calls    []promptCall
	attempts int
	seen     map[string]struct{}
	err      error
}

func (d *recordingPromptDispatcher) Enabled() bool {
	return true
}

func (d *recordingPromptDispatcher) EnqueueSendPrompt(
	_ context.Context,
	_ int64,
	runnerID int64,
	podKey, commandID, prompt string,
	_ time.Duration,
) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.attempts++
	if d.err != nil {
		return d.err
	}
	if d.seen == nil {
		d.seen = make(map[string]struct{})
	}
	if _, duplicate := d.seen[commandID]; duplicate {
		return agentpod.ErrDuplicateCommand
	}
	d.seen[commandID] = struct{}{}
	d.calls = append(d.calls, promptCall{
		runnerID:  runnerID,
		podKey:    podKey,
		commandID: commandID,
		prompt:    prompt,
	})
	return nil
}

type concurrentVerificationRepo struct {
	*goalLoopTestRepo
	readers sync.WaitGroup
	release chan struct{}
	once    sync.Once
	consume sync.Mutex
}

func newConcurrentVerificationRepo(loop *domain.GoalLoop) *concurrentVerificationRepo {
	repo := &concurrentVerificationRepo{
		goalLoopTestRepo: newGoalLoopTestRepo(loop),
		release:          make(chan struct{}),
	}
	repo.readers.Add(2)
	return repo
}

func (r *concurrentVerificationRepo) GetByVerificationRequestID(
	ctx context.Context,
	requestID string,
) (*domain.GoalLoop, error) {
	loop, err := r.goalLoopTestRepo.GetByVerificationRequestID(ctx, requestID)
	var snapshot *domain.GoalLoop
	if loop != nil {
		clone := *loop
		snapshot = &clone
	}
	r.readers.Done()
	<-r.release
	return snapshot, err
}

func (r *concurrentVerificationRepo) waitForReaders() {
	r.readers.Wait()
	r.once.Do(func() { close(r.release) })
}

func (r *concurrentVerificationRepo) ConsumeVerificationResult(
	ctx context.Context,
	id int64,
	requestID string,
	updates map[string]any,
) (bool, error) {
	r.consume.Lock()
	defer r.consume.Unlock()
	return r.goalLoopTestRepo.ConsumeVerificationResult(ctx, id, requestID, updates)
}
