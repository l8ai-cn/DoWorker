package coordinator

import (
	"context"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
)

// fakeStore is an in-memory coordinatordom.Repository for service tests. Only
// the methods exercised by RunProject / HandlePodTerminated carry real logic;
// the rest are minimal.
type fakeStore struct {
	projects   []*coordinatordom.Project
	links      []*coordinatordom.TicketExternalLink
	executions []*coordinatordom.Execution
	nextExecID int64
}

func newFakeStore() *fakeStore { return &fakeStore{} }

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

func (s *fakeStore) UpdateProject(context.Context, int64, int64, map[string]any) error { return nil }
func (s *fakeStore) DeleteProject(context.Context, int64, int64) error                 { return nil }

func (s *fakeStore) GetLinkByExternalID(_ context.Context, orgID int64, platformType, externalID string) (*coordinatordom.TicketExternalLink, error) {
	for _, l := range s.links {
		if l.OrganizationID == orgID && l.PlatformType == platformType && l.ExternalID == externalID {
			return l, nil
		}
	}
	return nil, coordinatordom.ErrNotFound
}

func (s *fakeStore) CreateLink(_ context.Context, l *coordinatordom.TicketExternalLink) error {
	s.links = append(s.links, l)
	return nil
}

func (s *fakeStore) CreateExecution(_ context.Context, e *coordinatordom.Execution) error {
	s.nextExecID++
	e.ID = s.nextExecID
	s.executions = append(s.executions, e)
	return nil
}

func (s *fakeStore) GetExecution(_ context.Context, id int64) (*coordinatordom.Execution, error) {
	for _, e := range s.executions {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, coordinatordom.ErrNotFound
}

func (s *fakeStore) GetExecutionByPodKey(_ context.Context, podKey string) (*coordinatordom.Execution, error) {
	for _, e := range s.executions {
		if e.PodKey != nil && *e.PodKey == podKey {
			return e, nil
		}
	}
	return nil, coordinatordom.ErrNotFound
}

func (s *fakeStore) ListExecutions(_ context.Context, projectID int64, _ int) ([]*coordinatordom.Execution, error) {
	var out []*coordinatordom.Execution
	for _, e := range s.executions {
		if e.ProjectID == projectID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (s *fakeStore) UpdateExecution(_ context.Context, id int64, updates map[string]any) error {
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
	}
	return nil
}

func (s *fakeStore) CountActiveExecutions(_ context.Context, projectID int64) (int64, error) {
	var n int64
	for _, e := range s.executions {
		if e.ProjectID == projectID && !coordinatordom.IsTerminalStatus(e.Status) {
			n++
		}
	}
	return n, nil
}
