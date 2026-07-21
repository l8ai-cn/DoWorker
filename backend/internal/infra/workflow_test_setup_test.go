package infra

import (
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"gorm.io/gorm"
)

// setupLoopTestDB creates an in-memory SQLite database for testing.
// Creates workflow-related tables plus minimal pods/autopilot_controllers tables for SSOT queries.
func setupLoopTestDB(t *testing.T) *gorm.DB {
	return testkit.SetupTestDB(t)
}

// Helper functions for creating test data
func workflowStrPtr(s string) *string { return &s }

func bindWorkflowResourceForExecution(row *workflow.Workflow, seed int64) {
	revision := int64(1)
	snapshotID := seed + 1000
	row.OrchestrationResourceID = &seed
	row.OrchestrationResourceRevision = &revision
	row.WorkerSpecSnapshotID = &snapshotID
}
