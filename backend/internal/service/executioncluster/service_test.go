package executioncluster

import (
	"context"
	"testing"
	"time"

	runnerdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestListEnsuresDefaultsAndAggregatesRunnerStatus(t *testing.T) {
	db := testkit.SetupTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO organizations (name, slug) VALUES ('Test Org', 'test-org')`).Error)
	var organizationID int64
	require.NoError(t, db.Raw(`SELECT id FROM organizations WHERE slug = 'test-org'`).Scan(&organizationID).Error)

	clusters := infra.NewExecutionClusterRepository(db)
	runners := infra.NewRunnerRepository(db)
	service := NewService(clusters, runners, nil, "")
	views, err := service.List(context.Background(), organizationID)
	require.NoError(t, err)
	require.Len(t, views, 2)
	online := findCluster(t, views, "online")
	seenAt := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, runners.Create(context.Background(), &runnerdomain.Runner{
		OrganizationID:    organizationID,
		ClusterID:         online.Cluster.ID,
		NodeID:            "online-runner",
		Status:            runnerdomain.RunnerStatusOnline,
		IsEnabled:         true,
		MaxConcurrentPods: 1,
		TunnelState:       "connected",
		TunnelLastSeenAt:  &seenAt,
	}))

	views, err = service.List(context.Background(), organizationID)
	require.NoError(t, err)
	online = findCluster(t, views, "online")
	require.Equal(t, 1, online.RunnerCount)
	require.Equal(t, 1, online.OnlineRunnerCount)
	require.Equal(t, 1, online.AvailableRunnerCount)
	require.Equal(t, "connected", online.TunnelStatus)
	require.Equal(t, seenAt, *online.TunnelLastSeenAt)
}

func findCluster(t *testing.T, views []View, slug string) View {
	t.Helper()
	for _, view := range views {
		if view.Cluster.Slug.String() == slug {
			return view
		}
	}
	t.Fatalf("cluster %q not found", slug)
	return View{}
}
