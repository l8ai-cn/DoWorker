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
		"'protocol_adapter'",
		"binding ? 'protocol_adapter'",
		"binding->>'protocol_adapter' ~ '^[a-z0-9]+(-[a-z0-9]+)*$'",
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
}
