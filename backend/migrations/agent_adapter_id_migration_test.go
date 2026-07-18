package migrations

import (
	"strings"
	"testing"
)

func TestMigration000207AddsValidatedAgentAdapterID(t *testing.T) {
	up, err := FS.ReadFile("000207_add_agent_adapter_id.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	for _, fragment := range []string{
		"ALTER TABLE agents ADD COLUMN adapter_id",
		"WHEN 'cursor-cli' THEN 'cursor-acp'",
		"WHEN 'do-agent' THEN 'do-agent-acp'",
		"ALTER COLUMN adapter_id SET NOT NULL",
		"agents_adapter_id_check",
		"adapter_id ~ '^[a-z0-9]+(-[a-z0-9]+)*$'",
	} {
		if !strings.Contains(string(up), fragment) {
			t.Errorf("up migration must contain %q", fragment)
		}
	}

	down, err := FS.ReadFile("000207_add_agent_adapter_id.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	for _, fragment := range []string{
		"DROP CONSTRAINT IF EXISTS agents_adapter_id_check",
		"DROP COLUMN IF EXISTS adapter_id",
	} {
		if !strings.Contains(string(down), fragment) {
			t.Errorf("down migration must contain %q", fragment)
		}
	}
}
