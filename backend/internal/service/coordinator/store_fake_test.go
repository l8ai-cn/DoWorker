package coordinator

import (
	"context"
	"fmt"
	"sync"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
)

// fakeStore is an in-memory coordinatordom.Repository for service tests. Only
// the methods exercised by RunProject / HandlePodTerminated carry real logic;
// the rest are minimal.
type fakeStore struct {
	mu                     sync.Mutex
	taskLocksMu            sync.Mutex
	taskLocks              map[string]*sync.Mutex
	projects               []*coordinatordom.Project
	links                  []*coordinatordom.TicketExternalLink
	executions             []*coordinatordom.Execution
	nextExecID             int64
	createExecutionErr     error
	nextUpdateError        error
	failCancelledUpdate    bool
	afterMissingLinkLookup func()
}

func newFakeStore() *fakeStore {
	return &fakeStore{taskLocks: make(map[string]*sync.Mutex)}
}

func (s *fakeStore) CreateProject(_ context.Context, p *coordinatordom.Project) error {
	s.projects = append(s.projects, p)
	return nil
}

func (s *fakeStore) GetProject(_ context.Context, orgID, id int64) (*coordinatordom.Project, error) {
	for _, p := range s.projects {
		if p.ID == id && p.OrganizationID == orgID {
			return p, nil
		}
	}
	return nil, coordinatordom.ErrNotFound
}

func (s *fakeStore) GetProjectBySlug(_ context.Context, orgID int64, slug string) (*coordinatordom.Project, error) {
	for _, p := range s.projects {
		if p.Slug == slug && p.OrganizationID == orgID {
			return p, nil
		}
	}
	return nil, coordinatordom.ErrNotFound
}

func (s *fakeStore) ListProjects(_ context.Context, _ *coordinatordom.ProjectListFilter) ([]*coordinatordom.Project, error) {
	return s.projects, nil
}

func (s *fakeStore) ListEnabledProjects(_ context.Context) ([]*coordinatordom.Project, error) {
	return s.projects, nil
}

func (s *fakeStore) UpdateProject(_ context.Context, orgID, id int64, updates map[string]any) error {
	for _, project := range s.projects {
		if project.ID != id || project.OrganizationID != orgID {
			continue
		}
		if workerType, ok := updates["agent_slug"].(string); ok {
			project.AgentSlug = workerType
		}
		if snapshotID, ok := updates["worker_spec_snapshot_id"].(int64); ok {
			project.WorkerSpecSnapshotID = &snapshotID
		}
		return nil
	}
	return coordinatordom.ErrNotFound
}
func (s *fakeStore) DeleteProject(context.Context, int64, int64) error { return nil }

func (s *fakeStore) GetLinkByExternalID(_ context.Context, orgID int64, platformType, externalID string) (*coordinatordom.TicketExternalLink, error) {
	s.mu.Lock()
	for _, l := range s.links {
		if l.OrganizationID == orgID && l.PlatformType == platformType && l.ExternalID == externalID {
			s.mu.Unlock()
			return l, nil
		}
	}
	s.mu.Unlock()
	if s.afterMissingLinkLookup != nil {
		s.afterMissingLinkLookup()
	}
	return nil, coordinatordom.ErrNotFound
}

func (s *fakeStore) CreateLink(_ context.Context, l *coordinatordom.TicketExternalLink) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.links = append(s.links, l)
	return nil
}

func (s *fakeStore) CreateExecution(_ context.Context, e *coordinatordom.Execution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.createExecutionErr != nil {
		return s.createExecutionErr
	}
	s.nextExecID++
	e.ID = s.nextExecID
	s.executions = append(s.executions, e)
	return nil
}

func (s *fakeStore) WithinProjectDispatch(
	_ context.Context,
	projectID int64,
	fn func(coordinatordom.Repository) error,
) error {
	key := fmt.Sprintf("%d", projectID)
	s.taskLocksMu.Lock()
	lock := s.taskLocks[key]
	if lock == nil {
		lock = &sync.Mutex{}
		s.taskLocks[key] = lock
	}
	s.taskLocksMu.Unlock()

	lock.Lock()
	defer lock.Unlock()
	s.mu.Lock()
	links := append([]*coordinatordom.TicketExternalLink(nil), s.links...)
	executions := append([]*coordinatordom.Execution(nil), s.executions...)
	nextExecID := s.nextExecID
	s.mu.Unlock()
	if err := fn(s); err != nil {
		s.mu.Lock()
		s.links = links
		s.executions = executions
		s.nextExecID = nextExecID
		s.mu.Unlock()
		return err
	}
	return nil
}

func (s *fakeStore) GetExecution(_ context.Context, id int64) (*coordinatordom.Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.executions {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, coordinatordom.ErrNotFound
}

func (s *fakeStore) GetExecutionByPodKey(_ context.Context, podKey string) (*coordinatordom.Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.executions {
		if e.PodKey != nil && *e.PodKey == podKey {
			return e, nil
		}
	}
	return nil, coordinatordom.ErrNotFound
}

func (s *fakeStore) GetActiveExecutionByProjectAndExternalID(
	_ context.Context,
	projectID int64,
	externalID string,
) (*coordinatordom.Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, execution := range s.executions {
		if execution.ProjectID == projectID &&
			execution.ExternalID == externalID &&
			!coordinatordom.IsTerminalStatus(execution.Status) {
			return execution, nil
		}
	}
	return nil, coordinatordom.ErrNotFound
}

func (s *fakeStore) ListExecutions(_ context.Context, projectID int64, _ int) ([]*coordinatordom.Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*coordinatordom.Execution
	for _, e := range s.executions {
		if e.ProjectID == projectID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (s *fakeStore) UpdateExecution(ctx context.Context, id int64, updates map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failCancelledUpdate && ctx.Err() != nil {
		return ctx.Err()
	}
	if s.nextUpdateError != nil {
		err := s.nextUpdateError
		s.nextUpdateError = nil
		return err
	}
	for _, e := range s.executions {
		if e.ID != id {
			continue
		}
		if v, ok := updates["status"].(string); ok {
			e.Status = v
		}
		if v, ok := updates["stage"].(string); ok {
			e.Stage = v
		}
		if v, ok := updates["feedback_status"].(string); ok {
			e.FeedbackStatus = v
		}
		if v, ok := updates["error"].(string); ok {
			e.Error = v
		}
		if v, ok := updates["pod_id"].(int64); ok {
			e.PodID = &v
		}
		if v, ok := updates["pod_key"].(string); ok {
			e.PodKey = &v
		}
	}
	return nil
}

func (s *fakeStore) CountActiveExecutions(_ context.Context, projectID int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n int64
	for _, e := range s.executions {
		if e.ProjectID == projectID && !coordinatordom.IsTerminalStatus(e.Status) {
			n++
		}
	}
	return n, nil
}

func (s *fakeStore) linkCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.links)
}
