package migrations

import (
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func TestMigration000188MiniMaxCLIAgent(t *testing.T) {
	up, err := FS.ReadFile("000188_add_minimax_cli_agent.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upSQL := string(up)

	for _, expected := range []string{
		"'minimax-cli'",
		"'mmx'",
		"AGENT mmx",
		"EXECUTABLE mmx",
		"MODE pty",
		`arg "text"`,
		`arg "repl"`,
	} {
		if !strings.Contains(upSQL, expected) {
			t.Errorf("up migration must contain %q", expected)
		}
	}

	down, err := FS.ReadFile("000188_add_minimax_cli_agent.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downSQL := string(down)

	configsIdx := strings.Index(downSQL, "DELETE FROM organization_agent_configs")
	agentsIdx := strings.Index(downSQL, "DELETE FROM agents WHERE slug")
	if configsIdx < 0 || agentsIdx < 0 {
		t.Fatal("down migration must remove dependent organization agent rows and the agent")
	}
	if configsIdx > agentsIdx {
		t.Error("down migration must remove dependent rows before deleting the agent")
	}
}

func TestMigrationSourceReads188AndContinuesAt190(t *testing.T) {
	source, err := iofs.New(FS, ".")
	if err != nil {
		t.Fatalf("create migration source: %v", err)
	}
	t.Cleanup(func() { _ = source.Close() })

	down, _, err := source.ReadDown(188)
	if err != nil {
		t.Fatalf("read migration 188 down: %v", err)
	}
	t.Cleanup(func() { _ = down.Close() })

	next, err := source.Next(188)
	if err != nil {
		t.Fatalf("find migration after 188: %v", err)
	}
	if next != 190 {
		t.Fatalf("migration after 188 = %d, want 190", next)
	}
}
