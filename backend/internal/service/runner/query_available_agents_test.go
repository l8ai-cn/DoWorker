package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	runnerDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingGrantQuerier struct{}

func (failingGrantQuerier) GetGrantedResourceIDs(context.Context, string, int64, int64) ([]string, error) {
	return nil, errors.New("grant lookup unavailable")
}

func TestListAvailableAgentSlugsUsesActiveEligibleRunners(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	now := time.Now()
	clusterSeven := createTestExecutionCluster(t, db, 7, "local")
	clusterEight := createTestExecutionCluster(t, db, 8, "local")

	runners := []*runnerDomain.Runner{
		{
			ID: 1, OrganizationID: 7, Status: runnerDomain.RunnerStatusOnline,
			ClusterID: clusterSeven, NodeID: "runner-1",
			IsEnabled: true, MaxConcurrentPods: 2,
			Visibility:      runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"codex-cli", "claude-code"},
		},
		{
			ID: 2, OrganizationID: 7, Status: runnerDomain.RunnerStatusOnline,
			ClusterID: clusterSeven, NodeID: "runner-2",
			IsEnabled: true, MaxConcurrentPods: 1,
			Visibility:      runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"codex-cli"},
		},
		{
			ID: 3, OrganizationID: 7, Status: runnerDomain.RunnerStatusOnline,
			ClusterID: clusterSeven, NodeID: "runner-3",
			IsEnabled: true, MaxConcurrentPods: 1, CurrentPods: 1,
			Visibility:      runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"aider"},
		},
		{
			ID: 4, OrganizationID: 7, Status: runnerDomain.RunnerStatusOnline,
			ClusterID: clusterSeven, NodeID: "runner-4",
			IsEnabled: true, MaxConcurrentPods: 1,
			Visibility:      runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"gemini-cli"},
		},
		{
			ID: 5, OrganizationID: 8, Status: runnerDomain.RunnerStatusOnline,
			ClusterID: clusterEight, NodeID: "runner-5",
			IsEnabled: true, MaxConcurrentPods: 1,
			Visibility:      runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"aider"},
		},
	}
	for _, runner := range runners {
		require.NoError(t, service.repo.Create(context.Background(), runner))
	}
	service.activeRunners.Store(int64(1), &ActiveRunner{Runner: runners[0], LastPing: now})
	service.activeRunners.Store(int64(2), &ActiveRunner{Runner: runners[1], LastPing: now})
	service.activeRunners.Store(int64(3), &ActiveRunner{Runner: runners[2], LastPing: now})
	service.activeRunners.Store(int64(4), &ActiveRunner{
		Runner: runners[3], LastPing: now.Add(-2 * time.Minute),
	})
	service.activeRunners.Store(int64(5), &ActiveRunner{Runner: runners[4], LastPing: now})

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
