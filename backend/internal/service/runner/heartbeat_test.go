package runner

import (
	"context"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/runner"
)

type heartbeatOrderingRepository struct {
	runner.RunnerRepository
	runner *runner.Runner
}

func (r *heartbeatOrderingRepository) GetByID(context.Context, int64) (*runner.Runner, error) {
	copy := *r.runner
	copy.AvailableAgents = nil
	return &copy, nil
}

func (r *heartbeatOrderingRepository) UpdateFields(_ context.Context, _ int64, updates map[string]interface{}) error {
	if currentPods, ok := updates["current_pods"].(int); ok {
		r.runner.CurrentPods = currentPods
	}
	if status, ok := updates["status"].(string); ok {
		r.runner.Status = status
	}
	if lastHeartbeat, ok := updates["last_heartbeat"].(time.Time); ok {
		r.runner.LastHeartbeat = &lastHeartbeat
	}
	if version, ok := updates["runner_version"].(string); ok {
		r.runner.RunnerVersion = &version
	}
	return nil
}

// --- Heartbeat Tests ---

func TestHeartbeat(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	ctx := context.Background()

	// Create runner directly
	r := &runner.Runner{
		OrganizationID:    1,
		NodeID:            "test-runner",
		Description:       "Test",
		Status:            runner.RunnerStatusOffline,
		MaxConcurrentPods: 5,
		IsEnabled:         true,
	}
	db.Create(r)

	// Send heartbeat
	err := service.Heartbeat(ctx, r.ID, 2)
	if err != nil {
		t.Fatalf("failed to send heartbeat: %v", err)
	}

	// Check runner status was updated
	updated, _ := service.GetRunner(ctx, r.ID)
	if updated.Status != runner.RunnerStatusOnline {
		t.Errorf("expected Status '%s', got %s", runner.RunnerStatusOnline, updated.Status)
	}
	if updated.CurrentPods != 2 {
		t.Errorf("expected CurrentPods 2, got %d", updated.CurrentPods)
	}
	if updated.LastHeartbeat == nil {
		t.Error("expected LastHeartbeat to be set")
	}
}

func TestUpdateHeartbeatWithPods(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	ctx := context.Background()

	// Create runner directly
	r := &runner.Runner{
		OrganizationID:    1,
		NodeID:            "test-runner",
		Description:       "Test",
		Status:            runner.RunnerStatusOffline,
		MaxConcurrentPods: 5,
		IsEnabled:         true,
	}
	db.Create(r)

	pods := []HeartbeatPodInfo{
		{PodKey: "pod-1", Status: "running"},
		{PodKey: "pod-2", Status: "running"},
	}

	err := service.UpdateHeartbeatWithPods(ctx, r.ID, pods, "1.0.0")
	if err != nil {
		t.Fatalf("failed to update heartbeat with pods: %v", err)
	}

	updated, _ := service.GetRunner(ctx, r.ID)
	if updated.CurrentPods != 2 {
		t.Errorf("expected 2 pods, got %d", updated.CurrentPods)
	}
}

func TestUpdateHeartbeatWithPodsNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	ctx := context.Background()

	err := service.UpdateHeartbeatWithPods(ctx, 99999, nil, "1.0.0")
	if err != ErrRunnerNotFound {
		t.Errorf("expected ErrRunnerNotFound, got %v", err)
	}
}

func TestUpdateHeartbeatWithPodsRefreshesActiveRunnerFromUpdatedState(t *testing.T) {
	repository := &heartbeatOrderingRepository{
		runner: &runner.Runner{
			ID: 1,
		},
	}
	service := NewService(repository)
	service.activeRunners.Store(int64(1), &ActiveRunner{
		Runner: &runner.Runner{
			ID:              1,
			AvailableAgents: runner.StringSlice{"e2e-echo"},
		},
	})

	if err := service.UpdateHeartbeatWithPods(context.Background(), 1, []HeartbeatPodInfo{{PodKey: "pod-1"}}, "1.0.0"); err != nil {
		t.Fatalf("failed to update heartbeat: %v", err)
	}

	active, ok := service.activeRunners.Load(int64(1))
	if !ok {
		t.Fatal("expected active runner")
	}
	activeRunner := active.(*ActiveRunner).Runner
	if got := activeRunner.AvailableAgents; len(got) != 1 || got[0] != "e2e-echo" {
		t.Fatalf("expected refreshed available agents, got %v", got)
	}
	if activeRunner.CurrentPods != 1 || activeRunner.Status != runner.RunnerStatusOnline {
		t.Fatalf("expected heartbeat runtime fields, got pods=%d status=%s", activeRunner.CurrentPods, activeRunner.Status)
	}
}

func TestMarkOfflineRunners(t *testing.T) {
	db := setupTestDB(t)
	service := newTestService(db)
	ctx := context.Background()

	// Create runner directly
	r := &runner.Runner{
		OrganizationID:    1,
		NodeID:            "test-runner",
		Description:       "Test",
		Status:            runner.RunnerStatusOffline,
		MaxConcurrentPods: 5,
		IsEnabled:         true,
	}
	db.Create(r)

	// Mark as online with a recent heartbeat
	now := time.Now()
	db.Model(&runner.Runner{}).Where("id = ?", r.ID).Updates(map[string]interface{}{
		"status":         runner.RunnerStatusOnline,
		"last_heartbeat": now,
	})

	// Mark offline with a timeout longer than since heartbeat
	service.MarkOfflineRunners(ctx, time.Hour)

	// Should still be online
	updated, _ := service.GetRunner(ctx, r.ID)
	if updated.Status != runner.RunnerStatusOnline {
		t.Errorf("expected status online, got %s", updated.Status)
	}

	// Set old heartbeat
	oldTime := now.Add(-2 * time.Hour)
	db.Model(&runner.Runner{}).Where("id = ?", r.ID).Update("last_heartbeat", oldTime)

	// Now should be marked offline
	service.MarkOfflineRunners(ctx, time.Hour)

	updated, _ = service.GetRunner(ctx, r.ID)
	if updated.Status != runner.RunnerStatusOffline {
		t.Errorf("expected status offline, got %s", updated.Status)
	}
}
