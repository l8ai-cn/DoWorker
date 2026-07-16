package migrations

import (
	"strings"
	"testing"
)

func TestMigration000208UpgradesCursorCLIToAgentACP(t *testing.T) {
	up, err := FS.ReadFile("000208_upgrade_cursor_cli_agent.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upStr := string(up)

	for _, fragment := range []string{
		"UPDATE agents",
		"adapter_id = 'cursor-acp'",
		"launch_command = 'agent'",
		"executable = 'agent'",
		"MODE acp \"acp\"",
		"WHERE slug = 'cursor-cli'",
	} {
		if !strings.Contains(upStr, fragment) {
			t.Errorf("up migration must contain %q", fragment)
		}
	}

	down, err := FS.ReadFile("000208_upgrade_cursor_cli_agent.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	if !strings.Contains(string(down), "cursor-cli") {
		t.Error("down migration must target cursor-cli")
	}
}
