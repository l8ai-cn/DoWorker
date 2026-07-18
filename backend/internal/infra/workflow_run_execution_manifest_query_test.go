package infra

import (
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTimedOutWorkflowRunsQueryUsesPinnedManifest(t *testing.T) {
	db := setupLoopTestDB(t).Session(&gorm.Session{DryRun: true})
	var runs []*workflow.WorkflowRun
	query := timedOutWorkflowRunsQuery(db, []int64{7, 8}).Find(&runs)
	require.NoError(t, query.Error)
	sql := query.Statement.SQL.String()

	assert.Contains(t, sql, "workflow_runs.execution_manifest")
	assert.Contains(t, sql, "timeout_minutes")
	assert.NotContains(t, sql, "JOIN workflows")
}

func TestIdleWorkflowPodsQueryUsesPinnedManifest(t *testing.T) {
	db := setupLoopTestDB(t).Session(&gorm.Session{DryRun: true})
	var runs []*workflow.WorkflowRun
	query := idleWorkflowPodsQuery(db, []int64{7, 8}).Find(&runs)
	require.NoError(t, query.Error)
	sql := query.Statement.SQL.String()

	assert.Contains(t, sql, "workflow_runs.execution_manifest")
	assert.Contains(t, sql, "idle_timeout_seconds")
	assert.NotContains(t, sql, "JOIN workflows")
}
