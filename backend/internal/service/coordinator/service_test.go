package coordinator

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
	ticketDomain "github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	ticketSvc "github.com/anthropics/agentsmesh/backend/internal/service/ticket"
)

func newTestService(t *testing.T, platform TaskPlatform) (*Service, *fakeStore, *fakeTickets) {
	t.Helper()
	store := newFakeStore()
	tickets := &fakeTickets{}
	svc := NewService(Deps{
		Store:         store,
		Tickets:       tickets,
		Dispatch:      &fakeDispatch{},
		PodTerminator: &fakePodTerminator{},
		Platform:      staticFactory{platform: platform, repo: "org/repo"},
	})
	return svc, store, tickets
}

func TestRunProjectClaimsAndDispatches(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{
			{ExternalID: "issue:1", Number: 1, Title: "fix bug", State: "open", Labels: []string{"bug"}},
			{ExternalID: "issue:2", Number: 2, Title: "skip me", State: "open", Labels: []string{"wontfix"}},
		},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	svc, store, tickets := newTestService(t, platform)

	project := &coordinatordom.Project{
		ID: 7, OrganizationID: 1, RepositoryID: 3, PlatformType: coordinatordom.PlatformTypeCNB,
		LabelFilter: []string{"bug"}, MaxConcurrent: 5, AgentSlug: "do-agent",
	}

	res, err := svc.RunProject(context.Background(), project)
	if err != nil {
		t.Fatalf("RunProject: %v", err)
	}
	if res.Scanned != 2 {
		t.Fatalf("scanned = %d, want 2", res.Scanned)
	}
	if res.Dispatched != 1 {
		t.Fatalf("dispatched = %d, want 1 (only the bug-labelled task)", res.Dispatched)
	}
	if len(tickets.created) != 1 {
		t.Fatalf("tickets created = %d, want 1", len(tickets.created))
	}
	if len(store.executions) != 1 {
		t.Fatalf("executions = %d, want 1", len(store.executions))
	}
}

func TestRunProjectIsIdempotent(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{{ExternalID: "issue:1", Number: 1, Title: "t", State: "open"}},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	svc, store, _ := newTestService(t, platform)
	project := &coordinatordom.Project{ID: 1, OrganizationID: 1, RepositoryID: 1, PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 5}

	if _, err := svc.RunProject(context.Background(), project); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if _, err := svc.RunProject(context.Background(), project); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if len(store.executions) != 1 {
		t.Fatalf("executions = %d, want 1 (external link dedupes second run)", len(store.executions))
	}
}

func TestRunProjectRespectsBudget(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{
			{ExternalID: "issue:1", Number: 1, Title: "a", State: "open"},
			{ExternalID: "issue:2", Number: 2, Title: "b", State: "open"},
		},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	svc, store, _ := newTestService(t, platform)
	project := &coordinatordom.Project{ID: 1, OrganizationID: 1, RepositoryID: 1, PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 1}

	res, err := svc.RunProject(context.Background(), project)
	if err != nil {
		t.Fatalf("RunProject: %v", err)
	}
	if res.Dispatched != 1 {
		t.Fatalf("dispatched = %d, want 1 (max_concurrent=1)", res.Dispatched)
	}
	if len(store.executions) != 1 {
		t.Fatalf("executions = %d, want 1", len(store.executions))
	}
}

func TestRunProjectDoesNotDispatchWhenClaimedExecutionPersistenceFails(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{{ExternalID: "issue:1", Number: 1, Title: "t", State: "open"}},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	store := newFakeStore()
	store.createExecutionErr = errors.New("execution persistence failed")
	dispatch := &fakeDispatch{}
	terminator := &fakePodTerminator{}
	tickets := &fakeTickets{}
	svc := NewService(Deps{
		Store:         store,
		Tickets:       tickets,
		Dispatch:      dispatch,
		Platform:      staticFactory{platform: platform, repo: "org/repo"},
		PodTerminator: terminator,
	})
	project := &coordinatordom.Project{
		ID: 1, OrganizationID: 1, RepositoryID: 1,
		PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 5,
	}

	result, err := svc.RunProject(context.Background(), project)

	if err != nil {
		t.Fatalf("RunProject: %v", err)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("errors = %v, want execution persistence failure", result.Errors)
	}
	if dispatch.n != 0 {
		t.Fatalf("dispatches = %d, want 0", dispatch.n)
	}
	if len(terminator.podKeys) != 0 {
		t.Fatalf("terminated pods = %v, want none", terminator.podKeys)
	}
	if got := tickets.createdCount(); got != 0 {
		t.Fatalf("created tickets = %d, want 0 after compensation", got)
	}
	if got := store.linkCount(); got != 0 {
		t.Fatalf("external links = %d, want 0 after rollback", got)
	}
}

func TestRunProjectTerminatesPodWhenExecutionAttachmentFails(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{{ExternalID: "issue:1", Number: 1, Title: "t", State: "open"}},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	store := newFakeStore()
	store.nextUpdateError = errors.New("execution attachment failed")
	terminator := &fakePodTerminator{}
	svc := NewService(Deps{
		Store:         store,
		Tickets:       &fakeTickets{},
		Dispatch:      &fakeDispatch{},
		Platform:      staticFactory{platform: platform, repo: "org/repo"},
		PodTerminator: terminator,
	})
	project := &coordinatordom.Project{
		ID: 1, OrganizationID: 1, RepositoryID: 1,
		PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 5,
	}

	result, err := svc.RunProject(context.Background(), project)
	if err != nil {
		t.Fatalf("RunProject: %v", err)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("errors = %v, want execution attachment failure", result.Errors)
	}
	if len(terminator.podKeys) != 1 || terminator.podKeys[0] != "pod-key" {
		t.Fatalf("terminated pods = %v, want [pod-key]", terminator.podKeys)
	}
	if got := store.executions[0].Status; got != coordinatordom.ExecutionStatusFailed {
		t.Fatalf("execution status = %q, want failed", got)
	}

	retry, err := svc.RunProject(context.Background(), project)
	if err != nil {
		t.Fatalf("retry RunProject: %v", err)
	}
	if retry.Dispatched != 1 {
		t.Fatalf("retry dispatched = %d, want 1", retry.Dispatched)
	}
	if len(store.executions) != 2 {
		t.Fatalf("retry executions = %d, want 2", len(store.executions))
	}
}

func TestRunProjectReportsTicketCompensationFailure(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{{ExternalID: "issue:1", Number: 1, Title: "t", State: "open"}},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	store := newFakeStore()
	store.createExecutionErr = errors.New("execution persistence failed")
	tickets := &fakeTickets{deleteErr: errors.New("ticket cleanup failed")}
	svc := NewService(Deps{
		Store:         store,
		Tickets:       tickets,
		Dispatch:      &fakeDispatch{},
		Platform:      staticFactory{platform: platform, repo: "org/repo"},
		PodTerminator: &fakePodTerminator{},
	})
	project := &coordinatordom.Project{
		ID: 1, OrganizationID: 1, RepositoryID: 1,
		PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 5,
	}

	result, err := svc.RunProject(context.Background(), project)

	if err != nil {
		t.Fatalf("RunProject: %v", err)
	}
	if len(result.Errors) != 1 ||
		!strings.Contains(result.Errors[0], "ticket cleanup failed") {
		t.Fatalf("errors = %v, want ticket cleanup failure", result.Errors)
	}
}

func TestHandlePodTerminatedPostsFeedback(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{{ExternalID: "issue:1", Number: 1, Title: "t", State: "open"}},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	svc, store, tickets := newTestService(t, platform)
	project := &coordinatordom.Project{ID: 1, OrganizationID: 1, RepositoryID: 1, PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 5}
	_ = store.CreateProject(context.Background(), project)
	if _, err := svc.RunProject(context.Background(), project); err != nil {
		t.Fatalf("RunProject: %v", err)
	}

	exec := store.executions[0]
	svc.HandlePodTerminated(context.Background(), *exec.PodKey, "completed")

	if platform.feedback == 0 {
		t.Fatalf("expected feedback comment to be posted")
	}
	if got := store.executions[0].Status; got != coordinatordom.ExecutionStatusSucceeded {
		t.Fatalf("execution status = %q, want succeeded", got)
	}
	if tickets.lastStatus != ticketDomain.TicketStatusInReview {
		t.Fatalf("ticket status = %q, want in_review", tickets.lastStatus)
	}
}

func TestHandlePodTerminatedIsIdempotent(t *testing.T) {
	platform := &fakePlatform{
		tasks: []ExternalTask{{ExternalID: "issue:1", Number: 1, Title: "t", State: "open"}},
		claim: ClaimResult{Claimed: true, Marker: "m"},
	}
	svc, store, _ := newTestService(t, platform)
	project := &coordinatordom.Project{ID: 1, OrganizationID: 1, RepositoryID: 1, PlatformType: coordinatordom.PlatformTypeCNB, MaxConcurrent: 5}
	_ = store.CreateProject(context.Background(), project)
	if _, err := svc.RunProject(context.Background(), project); err != nil {
		t.Fatalf("RunProject: %v", err)
	}

	podKey := *store.executions[0].PodKey
	svc.HandlePodTerminated(context.Background(), podKey, "completed")
	svc.HandlePodTerminated(context.Background(), podKey, "completed")

	if platform.feedback != 1 {
		t.Fatalf("feedback posts = %d, want 1", platform.feedback)
	}
}

func TestUpdateProjectReturnsNotFound(t *testing.T) {
	svc, _, _ := newTestService(t, &fakePlatform{})

	err := svc.UpdateProject(context.Background(), 1, 404, map[string]any{"name": "missing"})

	if !errors.Is(err, coordinatordom.ErrNotFound) {
		t.Fatalf("UpdateProject error = %v, want ErrNotFound", err)
	}
}

func TestDeleteProjectReturnsNotFound(t *testing.T) {
	svc, _, _ := newTestService(t, &fakePlatform{})

	err := svc.DeleteProject(context.Background(), 1, 404)

	if !errors.Is(err, coordinatordom.ErrNotFound) {
		t.Fatalf("DeleteProject error = %v, want ErrNotFound", err)
	}
}

// --- fakes ---

type fakePlatform struct {
	tasks    []ExternalTask
	claim    ClaimResult
	feedback int
}

func (f *fakePlatform) PlatformType() string { return coordinatordom.PlatformTypeCNB }
func (f *fakePlatform) DiscoverTasks(context.Context, string, coordinatordom.ClaimPolicy) ([]ExternalTask, error) {
	return f.tasks, nil
}
func (f *fakePlatform) TryClaim(context.Context, string, ExternalTask, string) (ClaimResult, error) {
	return f.claim, nil
}
func (f *fakePlatform) PostFeedback(context.Context, string, ExternalTask, string) error {
	f.feedback++
	return nil
}

type staticFactory struct {
	platform TaskPlatform
	repo     string
}

func (s staticFactory) For(context.Context, *coordinatordom.Project) (TaskPlatform, string, error) {
	return s.platform, s.repo, nil
}

type fakeDispatch struct{ n int64 }

func (f *fakeDispatch) CreatePod(context.Context, *agentpodSvc.OrchestrateCreatePodRequest) (*agentpodSvc.OrchestrateCreatePodResult, error) {
	f.n++
	now := time.Now()
	return &agentpodSvc.OrchestrateCreatePodResult{
		Pod: &podDomain.Pod{ID: f.n, PodKey: "pod-key", CreatedAt: now},
	}, nil
}

type fakePodTerminator struct {
	podKeys []string
	err     error
}

func (f *fakePodTerminator) TerminatePod(_ context.Context, podKey string) error {
	f.podKeys = append(f.podKeys, podKey)
	return f.err
}

type fakeTickets struct {
	mu         sync.Mutex
	created    []*ticketDomain.Ticket
	lastStatus string
	deleteErr  error
}

func (f *fakeTickets) CreateTicket(_ context.Context, req *ticketSvc.CreateTicketRequest) (*ticketDomain.Ticket, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	tk := &ticketDomain.Ticket{ID: int64(len(f.created) + 1), Title: req.Title}
	f.created = append(f.created, tk)
	return tk, nil
}
func (f *fakeTickets) GetTicket(_ context.Context, id int64) (*ticketDomain.Ticket, error) {
	return &ticketDomain.Ticket{ID: id}, nil
}
func (f *fakeTickets) UpdateStatus(_ context.Context, _ int64, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastStatus = status
	return nil
}

func (f *fakeTickets) DeleteTicket(_ context.Context, ticketID int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.deleteErr != nil {
		return f.deleteErr
	}
	for i, ticket := range f.created {
		if ticket.ID == ticketID {
			f.created = append(f.created[:i], f.created[i+1:]...)
			return nil
		}
	}
	return nil
}

func (f *fakeTickets) createdCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.created)
}
