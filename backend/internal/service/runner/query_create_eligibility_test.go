package runner

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/grant"
	runnerDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveRunnerForCreate(t *testing.T) {
	ctx := context.Background()
	const (
		orgID     = int64(1)
		userID    = int64(10)
		agentSlug = "claude-code"
	)
	otherUserID := int64(20)

	tests := []struct {
		name             string
		configure        func(*runnerDomain.Runner, *ActiveRunner)
		grant            bool
		allowUnavailable bool
		wantErr          bool
	}{
		{name: "eligible"},
		{name: "private non-owner with grant", configure: func(r *runnerDomain.Runner, _ *ActiveRunner) {
			r.Visibility = runnerDomain.VisibilityPrivate
			r.RegisteredByUserID = &otherUserID
		}, grant: true},
		{name: "other organization", configure: func(r *runnerDomain.Runner, _ *ActiveRunner) {
			r.OrganizationID = 2
		}, allowUnavailable: true, wantErr: true},
		{name: "private non-owner without grant", configure: func(r *runnerDomain.Runner, _ *ActiveRunner) {
			r.Visibility = runnerDomain.VisibilityPrivate
			r.RegisteredByUserID = &otherUserID
		}, allowUnavailable: true, wantErr: true},
		{name: "disabled", configure: func(r *runnerDomain.Runner, _ *ActiveRunner) {
			r.IsEnabled = false
		}, allowUnavailable: true, wantErr: true},
		{name: "unsupported agent", configure: func(r *runnerDomain.Runner, _ *ActiveRunner) {
			r.AvailableAgents = runnerDomain.StringSlice{"aider"}
		}, allowUnavailable: true, wantErr: true},
		{name: "at capacity", configure: func(r *runnerDomain.Runner, ar *ActiveRunner) {
			r.CurrentPods = r.MaxConcurrentPods
			ar.PodCount = r.MaxConcurrentPods
		}, wantErr: true},
		{name: "offline", configure: func(r *runnerDomain.Runner, _ *ActiveRunner) {
			r.Status = runnerDomain.RunnerStatusOffline
		}, wantErr: true},
		{name: "stale", configure: func(_ *runnerDomain.Runner, ar *ActiveRunner) {
			ar.LastPing = time.Now().Add(-2 * time.Minute)
		}, wantErr: true},
		{name: "allows unavailable offline runner", configure: func(r *runnerDomain.Runner, _ *ActiveRunner) {
			r.Status = runnerDomain.RunnerStatusOffline
		}, allowUnavailable: true},
		{name: "allows unavailable full runner", configure: func(r *runnerDomain.Runner, ar *ActiveRunner) {
			r.CurrentPods = r.MaxConcurrentPods
			ar.PodCount = r.MaxConcurrentPods
		}, allowUnavailable: true},
		{name: "allows unavailable stale runner", configure: func(_ *runnerDomain.Runner, ar *ActiveRunner) {
			ar.LastPing = time.Now().Add(-2 * time.Minute)
		}, allowUnavailable: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			service := newTestService(db)
			r := &runnerDomain.Runner{
				OrganizationID: orgID, NodeID: "runner-resolve", Status: runnerDomain.RunnerStatusOnline,
				IsEnabled: true, MaxConcurrentPods: 2, AvailableAgents: runnerDomain.StringSlice{agentSlug},
				Visibility: runnerDomain.VisibilityOrganization,
			}
			active := &ActiveRunner{Runner: r, LastPing: time.Now()}
			if tt.configure != nil {
				tt.configure(r, active)
			}
			disabled := !r.IsEnabled
			require.NoError(t, db.Create(r).Error)
			if disabled {
				require.NoError(t, db.Model(r).UpdateColumn("is_enabled", false).Error)
				r.IsEnabled = false
			}
			if tt.grant {
				require.NoError(t, db.Exec(
					"INSERT INTO resource_grants (organization_id, resource_type, resource_id, user_id, granted_by) VALUES (?, ?, ?, ?, ?)",
					orgID, grant.TypeRunner, strconv.FormatInt(r.ID, 10), userID, userID,
				).Error)
			}
			service.activeRunners.Store(r.ID, active)

			resolved, err := service.ResolveRunnerForCreate(ctx, r.ID, orgID, userID, agentSlug, tt.allowUnavailable)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrNoRunnerForAgent)
				assert.Nil(t, resolved)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, r.ID, resolved.ID)
		})
	}
}

func TestUpdateAvailableAgentsRefreshesActiveRunner(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	service := newTestService(db)
	r := &runnerDomain.Runner{
		OrganizationID:    1,
		NodeID:            "runner-agent-refresh",
		Status:            runnerDomain.RunnerStatusOnline,
		IsEnabled:         true,
		MaxConcurrentPods: 2,
		Visibility:        runnerDomain.VisibilityOrganization,
	}
	require.NoError(t, db.Create(r).Error)
	require.NoError(t, service.MarkConnected(ctx, r.ID))

	require.NoError(t, service.UpdateAvailableAgents(ctx, r.ID, []string{"e2e-echo"}))

	resolved, err := service.ResolveRunnerForCreate(ctx, r.ID, 1, 10, "e2e-echo", false)
	require.NoError(t, err)
	assert.Equal(t, r.ID, resolved.ID)
}
