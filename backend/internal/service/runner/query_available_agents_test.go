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
	service := newTestService(setupTestDB(t))
	now := time.Now()

	service.activeRunners.Store(int64(1), &ActiveRunner{
		Runner: &runnerDomain.Runner{
			ID: 1, OrganizationID: 7, Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 2, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"codex-cli", "claude-code"},
		},
		LastPing: now,
	})
	service.activeRunners.Store(int64(2), &ActiveRunner{
		Runner: &runnerDomain.Runner{
			ID: 2, OrganizationID: 7, Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 1, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"codex-cli"},
		},
		LastPing: now,
	})
	service.activeRunners.Store(int64(3), &ActiveRunner{
		Runner: &runnerDomain.Runner{
			ID: 3, OrganizationID: 7, Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 1, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"aider"},
		},
		LastPing: now,
		PodCount: 1,
	})
	service.activeRunners.Store(int64(4), &ActiveRunner{
		Runner: &runnerDomain.Runner{
			ID: 4, OrganizationID: 7, Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 1, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"gemini-cli"},
		},
		LastPing: now.Add(-2 * time.Minute),
	})
	service.activeRunners.Store(int64(5), &ActiveRunner{
		Runner: &runnerDomain.Runner{
			ID: 5, OrganizationID: 8, Status: runnerDomain.RunnerStatusOnline,
			IsEnabled: true, MaxConcurrentPods: 1, Visibility: runnerDomain.VisibilityOrganization,
			AvailableAgents: runnerDomain.StringSlice{"aider"},
		},
		LastPing: now,
	})

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
