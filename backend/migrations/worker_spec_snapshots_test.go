package migrations

import (
	"strings"
	"testing"
)

func TestMigration000194WorkerSpecSnapshots(t *testing.T) {
	up, err := FS.ReadFile("000194_worker_spec_snapshots.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upSQL := string(up)
	for _, fragment := range []string{
		"CREATE TABLE worker_spec_snapshots",
		"organization_id BIGINT NOT NULL",
		"version SMALLINT NOT NULL",
		"spec_json JSONB NOT NULL",
		"summary_json JSONB NOT NULL",
		"created_at TIMESTAMPTZ NOT NULL",
		"CHECK (organization_id > 0)",
		"CHECK (version = 1)",
		"jsonb_typeof(spec_json) = 'object'",
		"jsonb_typeof(summary_json) = 'object'",
		"spec_json->>'version' = version::text",
		"summary_json->>'version' = version::text",
		"worker_spec_snapshots_model_resource_id_consistent",
		"spec_json #>> '{runtime,model_resource_id}'",
		"summary_json->>'model_resource_id'",
		"^[1-9][0-9]*$",
		"9223372036854775807",
		"::BIGINT",
		"CREATE INDEX idx_worker_spec_snapshots_organization_created_at",
		"ON worker_spec_snapshots (organization_id, created_at DESC)",
		"CREATE FUNCTION prevent_worker_spec_snapshot_update()",
		"CREATE TRIGGER worker_spec_snapshots_immutable",
		"BEFORE UPDATE ON worker_spec_snapshots",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Errorf("up migration must contain %q", fragment)
		}
	}

	down, err := FS.ReadFile("000194_worker_spec_snapshots.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downSQL := string(down)
	triggerIndex := strings.Index(
		downSQL,
		"DROP TRIGGER IF EXISTS worker_spec_snapshots_immutable",
	)
	functionIndex := strings.Index(
		downSQL,
		"DROP FUNCTION IF EXISTS prevent_worker_spec_snapshot_update",
	)
	tableIndex := strings.Index(downSQL, "DROP TABLE IF EXISTS worker_spec_snapshots")
	if triggerIndex < 0 || functionIndex < 0 || tableIndex < 0 {
		t.Fatal("down migration must drop immutable trigger, function, and table")
	}
	if triggerIndex > functionIndex || functionIndex > tableIndex {
		t.Error("down migration must drop trigger, then function, then table")
	}
}
