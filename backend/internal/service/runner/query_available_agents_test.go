package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/stretchr/testify/assert"
)

type failingGrantQuerier struct{}

func (failingGrantQuerier) GetGrantedResourceIDs(context.Context, string, int64, int64) ([]string, error) {
	return nil, errors.New("grant lookup unavailable")
}

func TestListAvailableAgentSlugsUsesActiveEligibleRunners(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	now := time.Now()

	runners := []struct {
		runner   *runnerDomain.Runner
		lastPing time.Time
		podCount int
	}{
		{runner: &runnerDomain.Runner{
			ID: 1, OrganizationID: 7, NodeID: "agent-list-1", Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 2, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"codex-cli", "claude-code"},
		}, lastPing: now},
		{runner: &runnerDomain.Runner{
			ID: 2, OrganizationID: 7, NodeID: "agent-list-2", Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 1, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"codex-cli"},
		}, lastPing: now},
		{runner: &runnerDomain.Runner{
			ID: 3, OrganizationID: 7, NodeID: "agent-list-3", Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 1, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"aider"},
		}, lastPing: now, podCount: 1},
		{runner: &runnerDomain.Runner{
			ID: 4, OrganizationID: 7, NodeID: "agent-list-4", Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 1, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"gemini-cli"},
		}, lastPing: now.Add(-2 * time.Minute)},
		{runner: &runnerDomain.Runner{
			ID: 5, OrganizationID: 8, NodeID: "agent-list-5", Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 1, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"aider"},
		}, lastPing: now},
	}
	for _, candidate := range runners {
		storeActiveRunner(t, db, service, candidate.runner, candidate.lastPing, candidate.podCount)
	}

	got, err := service.ListAvailableAgentSlugs(context.Background(), 7, 11)

	assert.NoError(t, err)
	assert.Equal(t, []string{"claude-code", "codex-cli"}, got)
}

func TestListAvailableAgentSlugsFailsClosedWhenGrantLookupFails(t *testing.T) {
	service := newTestService(setupTestDB(t))
	service.SetGrantQuerier(failingGrantQuerier{})

	_, err := service.ListAvailableAgentSlugs(context.Background(), 7, 11)

	assert.ErrorContains(t, err, "grant lookup unavailable")
}
