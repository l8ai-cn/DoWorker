package runner

import (
	"context"
	"errors"
	"sync"
	"testing"

	runnerDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/runner"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func TestPodCoordinatorCapacityClaimPreventsConcurrentOversell(t *testing.T) {
	db, cm, tr, hb, podStore, runnerRepo := setupPodCoordinatorDeps(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	candidate := &runnerDomain.Runner{
		OrganizationID:    1,
		NodeID:            "capacity-runner",
		Status:            runnerDomain.RunnerStatusOnline,
		IsEnabled:         true,
		MaxConcurrentPods: 1,
	}
	if err := db.Create(candidate).Error; err != nil {
		t.Fatalf("create runner: %v", err)
	}

	coordinator := NewPodCoordinator(podStore, runnerRepo, cm, tr, hb, newTestLogger())
	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			results <- coordinator.IncrementPods(context.Background(), candidate.ID)
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	successes := 0
	capacityFailures := 0
	for result := range results {
		switch {
		case result == nil:
			successes++
		case errors.Is(result, runnerDomain.ErrRunnerCapacityUnavailable):
			capacityFailures++
		default:
			t.Fatalf("unexpected capacity claim error: %v", result)
		}
	}
	if successes != 1 || capacityFailures != 1 {
		t.Fatalf("claims: success=%d capacity_failure=%d", successes, capacityFailures)
	}
}

func TestPodCoordinatorDoesNotDispatchWithoutCapacityClaim(t *testing.T) {
	db, cm, tr, hb, podStore, runnerRepo := setupPodCoordinatorDeps(t)
	candidate := &runnerDomain.Runner{
		OrganizationID:    1,
		NodeID:            "full-runner",
		Status:            runnerDomain.RunnerStatusOnline,
		IsEnabled:         true,
		CurrentPods:       1,
		MaxConcurrentPods: 1,
	}
	if err := db.Create(candidate).Error; err != nil {
		t.Fatalf("create runner: %v", err)
	}

	coordinator := NewPodCoordinator(podStore, runnerRepo, cm, tr, hb, newTestLogger())
	sender := &MockCommandSender{}
	coordinator.SetCommandSender(sender)
	err := coordinator.CreatePod(context.Background(), candidate.ID, &runnerv1.CreatePodCommand{
		PodKey: "capacity-rejected",
	})
	if !errors.Is(err, runnerDomain.ErrRunnerCapacityUnavailable) {
		t.Fatalf("CreatePod() error = %v", err)
	}
	if sender.CreatePodCalls != 0 {
		t.Fatalf("create commands = %d, want 0", sender.CreatePodCalls)
	}
}
