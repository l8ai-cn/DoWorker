package migrations

import (
	"strings"
	"testing"
)

func TestMigration000209AllowsProtocolAdapterInWorkerSpecBinding(t *testing.T) {
	up, err := FS.ReadFile("000209_worker_spec_protocol_adapter.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upSQL := string(up)
	for _, fragment := range []string{
		"CREATE OR REPLACE FUNCTION worker_spec_model_binding_is_valid",
		"worker_spec_snapshots require protocol_adapter backfill before migration",
		"WHEN binding = '{}'::JSONB THEN TRUE",
		"'protocol_adapter'",
		"binding->>'protocol_adapter' ~ '^[a-z0-9]+(-[a-z0-9]+)*$'",
		"worker_spec_tool_model_binding_is_valid",
		"worker_spec_tool_model_bindings_are_valid",
		"worker_spec_snapshots_tool_models_consistent",
	} {
		if !strings.Contains(upSQL, fragment) {
			t.Errorf("up migration must contain %q", fragment)
		}
	}

	down, err := FS.ReadFile("000209_worker_spec_protocol_adapter.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	if !strings.Contains(string(down), "worker_spec_snapshots must be empty before protocol adapter rollback") {
		t.Error("down migration must protect immutable snapshots")
	}
	if !strings.Contains(string(down), "DROP FUNCTION worker_spec_tool_model_bindings_are_valid") {
		t.Error("down migration must remove tool model validation functions")
	}
}
