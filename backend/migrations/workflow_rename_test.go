package migrations

import (
	"strings"
	"testing"
)

func TestMigration000200RenamesScheduledLoopStorageToWorkflow(t *testing.T) {
	up, err := FS.ReadFile("000200_rename_loops_to_workflows.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	for _, statement := range []string{
		"ALTER TABLE loops RENAME TO workflows",
		"ALTER TABLE loop_runs RENAME TO workflow_runs",
		"ALTER TABLE workflow_runs RENAME COLUMN loop_id TO workflow_id",
		"ALTER SEQUENCE loops_id_seq RENAME TO workflows_id_seq",
		"ALTER SEQUENCE loop_runs_id_seq RENAME TO workflow_runs_id_seq",
		"IF EXISTS",
		"to_regclass('public.idx_loops_model_resource_id')",
		"ALTER INDEX idx_loops_org_slug RENAME TO idx_workflows_org_slug",
		"ALTER INDEX idx_loop_runs_loop_id RENAME TO idx_workflow_runs_workflow_id",
	} {
		if !strings.Contains(string(up), statement) {
			t.Errorf("up migration must contain %q", statement)
		}
	}

	down, err := FS.ReadFile("000200_rename_loops_to_workflows.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	for _, statement := range []string{
		"ALTER TABLE workflows RENAME TO loops",
		"ALTER TABLE workflow_runs RENAME TO loop_runs",
		"ALTER TABLE workflow_runs RENAME COLUMN workflow_id TO loop_id",
		"ALTER SEQUENCE workflows_id_seq RENAME TO loops_id_seq",
		"ALTER SEQUENCE workflow_runs_id_seq RENAME TO loop_runs_id_seq",
	} {
		if !strings.Contains(string(down), statement) {
			t.Errorf("down migration must contain %q", statement)
		}
	}
}
