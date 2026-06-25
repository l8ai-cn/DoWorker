package coordinator

import (
	"context"

	coordinatordom "github.com/anthropics/agentsmesh/backend/internal/domain/coordinator"
)

// ExternalTask is the platform-agnostic view of a discovered task. It carries
// everything the coordinator needs to evaluate the claim policy, sync a ticket,
// and post feedback.
type ExternalTask struct {
	ExternalID  string
	Number      int
	Kind        string
	Title       string
	Description string
	State       string
	Type        string
	Priority    string
	Labels      []string
	Assignees   []string
	URL         string
}

func (t ExternalTask) Candidate() coordinatordom.Candidate {
	return coordinatordom.Candidate{
		Title:       t.Title,
		Description: t.Description,
		State:       t.State,
		Type:        t.Type,
		Priority:    t.Priority,
		Labels:      t.Labels,
		Assignees:   t.Assignees,
	}
}

// ClaimResult reports whether the platform-level claim (a marker comment)
// succeeded. Claimed=false with a non-empty Reason means another coordinator
// already holds the task.
type ClaimResult struct {
	Claimed bool
	Reason  string
	Marker  string
}

// TaskPlatform abstracts the external task source (CNB issues, Linear, ...).
// Implementations are constructed per-scan with a repository-scoped credential.
type TaskPlatform interface {
	PlatformType() string
	DiscoverTasks(ctx context.Context, repo string, policy coordinatordom.ClaimPolicy) ([]ExternalTask, error)
	TryClaim(ctx context.Context, repo string, task ExternalTask, claimKey string) (ClaimResult, error)
	PostFeedback(ctx context.Context, repo string, task ExternalTask, body string) error
}
