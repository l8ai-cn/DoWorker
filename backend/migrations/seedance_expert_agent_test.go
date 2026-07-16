package migrations

import (
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/agentfile/parser"
)

func TestMigration000210SeedanceExpertAgent(t *testing.T) {
	up, err := FS.ReadFile("000210_add_seedance_expert_agent.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upSQL := string(up)
	for _, expected := range []string{
		"'seedance-expert'",
		"'do-agent-acp'",
		"AGENT do-agent",
		"EXECUTABLE do-agent",
		"ENV SEEDANCE_API_KEY SECRET",
		"ENV SEEDANCE_BASE_URL",
		"ENV SEEDANCE_MODEL",
	} {
		if !strings.Contains(upSQL, expected) {
			t.Errorf("up migration must contain %q", expected)
		}
	}
	if strings.Contains(upSQL, "AGENT seedance-expert") {
		t.Error("AgentFile AGENT must be the do-agent binary, not the Worker slug")
	}
	if _, errors := parser.Parse(extractSeedanceAgentFile(t, upSQL)); len(errors) > 0 {
		t.Fatalf("Seedance Expert AgentFile must parse: %v", errors)
	}

	down, err := FS.ReadFile("000210_add_seedance_expert_agent.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downSQL := string(down)
	agentIndex := strings.Index(downSQL, "DELETE FROM agents WHERE slug")
	for _, dependent := range []string{
		"DELETE FROM organization_agent_configs",
		"DELETE FROM organization_agents",
		"DELETE FROM user_agent_configs",
	} {
		index := strings.Index(downSQL, dependent)
		if index < 0 || agentIndex < 0 || index > agentIndex {
			t.Fatalf("down migration must delete %s before the agent", dependent)
		}
	}
}

func extractSeedanceAgentFile(t *testing.T, migration string) string {
	t.Helper()
	start := strings.Index(migration, "E'")
	end := strings.LastIndex(migration, "'\n);")
	if start < 0 || end <= start+2 {
		t.Fatal("migration does not contain the AgentFile literal")
	}
	return strings.ReplaceAll(migration[start+2:end], `\n`, "\n")
}
