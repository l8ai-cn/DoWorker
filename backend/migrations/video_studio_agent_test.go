package migrations

import (
	"strings"
	"testing"
)

func TestMigration000213VideoStudioAgent(t *testing.T) {
	up, err := FS.ReadFile("000213_add_video_studio_agent.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upSQL := string(up)
	for _, expected := range []string{
		"'video-studio'",
		"'Video Studio'",
		"'video-studio-codex'",
		"FROM agents",
		"WHERE slug = 'codex-cli'",
		"RAISE EXCEPTION",
	} {
		if !strings.Contains(upSQL, expected) {
			t.Errorf("up migration must contain %q", expected)
		}
	}

	down, err := FS.ReadFile("000213_add_video_studio_agent.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downSQL := string(down)
	agentIndex := strings.Index(downSQL, "DELETE FROM agents WHERE")
	if agentIndex < 0 {
		t.Fatal("down migration must remove video-studio")
	}
	for _, table := range []string{
		"organization_agent_configs",
		"organization_agents",
		"user_agent_configs",
	} {
		index := strings.Index(downSQL, "DELETE FROM "+table)
		if index < 0 || index > agentIndex {
			t.Errorf("%s must be cleared before agents", table)
		}
	}
}
