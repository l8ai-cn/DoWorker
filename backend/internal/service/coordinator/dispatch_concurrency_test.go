package coordinator

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

func TestRunProjectSerializesConcurrentDispatchForTask(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{{ExternalID: "issue:1", Number: 1, Title: "t", State: "open"}},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	store := newFakeStore()
	linkMissed := make(chan struct{}, 2)
	releaseLinkMiss := make(chan struct{})
	store.afterMissingLinkLookup = func() {
		linkMissed <- struct{}{}
		<-releaseLinkMiss
	}
	tickets := &fakeTickets{}
	dispatch := &countingDispatch{}
	svc := NewService(Deps{
		Store:         store,
		Tickets:       tickets,
		Dispatch:      dispatch,
		PodTerminator: &fakePodTerminator{},
		Platform:      staticFactory{platform: platform, repo: "org/repo"},
	})
	project := &coordinatordom.Project{
		ID: 1, OrganizationID: 1, RepositoryID: 1,
		PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 5,
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	run := func() {
		defer wg.Done()
		_, err := svc.RunProject(context.Background(), project)
		errs <- err
	}

	wg.Add(1)
	go run()
	<-linkMissed

	wg.Add(1)
	go run()

	select {
	case <-linkMissed:
		close(releaseLinkMiss)
	case <-time.After(200 * time.Millisecond):
		close(releaseLinkMiss)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("RunProject: %v", err)
		}
	}
	if got := dispatch.calls.Load(); got != 1 {
		t.Fatalf("dispatch calls = %d, want 1", got)
	}
	if got := tickets.createdCount(); got != 1 {
		t.Fatalf("created tickets = %d, want 1", got)
	}
}

func TestRunProjectCompensatesAttachmentAfterRequestCancellation(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{{ExternalID: "issue:1", Number: 1, Title: "t", State: "open"}},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	store := newFakeStore()
	store.failCancelledUpdate = true
	ctx, cancel := context.WithCancel(context.Background())
	terminator := &contextCheckingTerminator{}
	svc := NewService(Deps{
		Store:         store,
		Tickets:       &fakeTickets{},
		Dispatch:      cancellingDispatch{cancel: cancel},
		PodTerminator: terminator,
		Platform:      staticFactory{platform: platform, repo: "org/repo"},
	})
	project := &coordinatordom.Project{
		ID: 1, OrganizationID: 1, RepositoryID: 1,
		PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 5,
	}

	result, err := svc.RunProject(ctx, project)

	if err != nil {
		t.Fatalf("RunProject: %v", err)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("errors = %v, want cancelled attachment failure", result.Errors)
	}
	if terminator.contextErr != nil {
		t.Fatalf("termination context = %v, want active compensation context", terminator.contextErr)
	}
	if got := store.executions[0].Status; got != coordinatordom.ExecutionStatusFailed {
		t.Fatalf("execution status = %q, want failed", got)
	}
}

func TestClaimAndDispatchEnforcesProjectBudgetAcrossTasks(t *testing.T) {
	platform := &fakePlatform{claim: ClaimResult{Claimed: true, Marker: "m"}}
	store := newFakeStore()
	dispatch := &countingDispatch{}
	svc := NewService(Deps{
		Store:         store,
		Tickets:       &fakeTickets{},
		Dispatch:      dispatch,
		PodTerminator: &fakePodTerminator{},
		Platform:      staticFactory{platform: platform, repo: "org/repo"},
	})
	project := &coordinatordom.Project{
		ID: 1, OrganizationID: 1, RepositoryID: 1,
		PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 1,
	}
	tasks := []ExternalTask{
		{ExternalID: "issue:1", Number: 1, Title: "first"},
		{ExternalID: "issue:2", Number: 2, Title: "second"},
	}

	var wg sync.WaitGroup
	results := make(chan bool, len(tasks))
	errs := make(chan error, len(tasks))
	for i := range tasks {
		task := tasks[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			dispatched, err := svc.claimAndDispatch(
				context.Background(),
				project,
				platform,
				"org/repo",
				task,
			)
			results <- dispatched
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)

	dispatched := 0
	for result := range results {
		if result {
			dispatched++
		}
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("claimAndDispatch: %v", err)
		}
	}
	if dispatched != 1 {
		t.Fatalf("dispatched = %d, want 1", dispatched)
	}
	if got := dispatch.calls.Load(); got != 1 {
		t.Fatalf("dispatch calls = %d, want 1", got)
	}
}

type countingDispatch struct {
	calls atomic.Int64
}

type cancellingDispatch struct {
	cancel context.CancelFunc
}

func (d cancellingDispatch) CreatePod(
	context.Context,
	*agentpodSvc.OrchestrateCreatePodRequest,
) (*agentpodSvc.OrchestrateCreatePodResult, error) {
	d.cancel()
	now := time.Now()
	return &agentpodSvc.OrchestrateCreatePodResult{
		Pod: &podDomain.Pod{ID: 1, PodKey: "pod-key", CreatedAt: now},
	}, nil
}

type contextCheckingTerminator struct {
	contextErr error
}

func (t *contextCheckingTerminator) TerminatePod(ctx context.Context, _ string) error {
	t.contextErr = ctx.Err()
	return nil
}

func (d *countingDispatch) CreatePod(
	context.Context,
	*agentpodSvc.OrchestrateCreatePodRequest,
) (*agentpodSvc.OrchestrateCreatePodResult, error) {
	id := d.calls.Add(1)
	now := time.Now()
	return &agentpodSvc.OrchestrateCreatePodResult{
		Pod: &podDomain.Pod{ID: id, PodKey: "pod-key", CreatedAt: now},
	}, nil
}
