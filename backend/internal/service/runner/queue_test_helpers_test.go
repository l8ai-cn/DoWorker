package runner

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerDomain "github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
	"gorm.io/gorm"
)

func createOnlineRunner(t *testing.T, db *gorm.DB, runnerRepo runnerDomain.RunnerRepository, currentPods, maxPods int) *runnerDomain.Runner {
	t.Helper()
	orgID := testkit.CreateOrg(t, db, fmt.Sprintf("org-%d", time.Now().UnixNano()), 0)
	nodeID := fmt.Sprintf("node-%d", time.Now().UnixNano())
	runnerID := testkit.CreateRunner(t, db, orgID, nodeID)
	if err := db.Exec(
		`UPDATE runners SET current_pods = ?, max_concurrent_pods = ? WHERE id = ?`,
		currentPods, maxPods, runnerID,
	).Error; err != nil {
		t.Fatalf("update runner capacity: %v", err)
	}
	run, err := runnerRepo.GetByID(context.Background(), runnerID)
	if err != nil {
		t.Fatalf("get runner: %v", err)
	}
	return run
}

func seedQueuedPod(t *testing.T, db *gorm.DB, orgID, runnerID int64, podKey string) error {
	t.Helper()
	return db.Exec(
		`INSERT INTO pods (organization_id, pod_key, runner_id, created_by_id, status) VALUES (?, ?, ?, 1, ?)`,
		orgID, podKey, runnerID, agentpod.StatusQueued,
	).Error
}
