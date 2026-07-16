package migrations

import (
	"strings"
	"testing"
)

func TestMigration000218FailsClosedOnIrreversibleRollback(t *testing.T) {
	down, err := FS.ReadFile("000218_normalize_agent_capability_heading.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downSQL := string(down)
	for _, fragment := range []string{
		"RAISE EXCEPTION",
		"restore from the pre-migration backup",
	} {
		if !strings.Contains(downSQL, fragment) {
			t.Errorf("down migration must contain %q", fragment)
		}
	}
}
