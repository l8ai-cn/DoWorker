package migrations

import (
	"strings"
	"testing"
)

func TestMigration000197WorkerSpecModelBinding(t *testing.T) {
	up, err := FS.ReadFile("000197_worker_spec_model_binding.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upSQL := string(up)
	for _, fragment := range []string{
		"worker_spec_snapshots must be empty before model binding migration",
		"DROP CONSTRAINT worker_spec_snapshots_model_resource_id_consistent",
		"CREATE FUNCTION worker_spec_jsonb_is_positive_int64",
		"CREATE FUNCTION worker_spec_model_binding_is_valid",
		"worker_spec_snapshots_spec_model_binding_valid",
		"worker_spec_snapshots_summary_model_binding_valid",
		"worker_spec_snapshots_model_binding_consistent",
		"spec_json #> '{runtime,model_binding}'",
		"summary_json->'model_binding'",
		"resource_id",
		"resource_revision",
		"connection_id",
		"connection_revision",
		"provider_key",
		"model_id",
		"^[a-z0-9]+(-[a-z0-9]+)*$",
		"9223372036854775807",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Errorf("up migration must contain %q", fragment)
		}
	}

	down, err := FS.ReadFile("000197_worker_spec_model_binding.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downSQL := string(down)
	for _, fragment := range []string{
		"worker_spec_snapshots must be empty before model binding rollback",
		"DROP CONSTRAINT worker_spec_snapshots_model_binding_consistent",
		"DROP CONSTRAINT worker_spec_snapshots_summary_model_binding_valid",
		"DROP CONSTRAINT worker_spec_snapshots_spec_model_binding_valid",
		"DROP FUNCTION worker_spec_model_binding_is_valid",
		"DROP FUNCTION worker_spec_jsonb_is_positive_int64",
		"ADD CONSTRAINT worker_spec_snapshots_model_resource_id_consistent",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Errorf("down migration must contain %q", fragment)
		}
	}
}
