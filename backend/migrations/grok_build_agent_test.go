package migrations

import (
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/agentfile/parser"
)

func TestMigration000191GrokBuildAgent(t *testing.T) {
	up, err := FS.ReadFile("000191_add_grok_build_agent.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upStr := string(up)
	for _, want := range []string{
		"'grok-build'",
		"AGENT grok",
		"EXECUTABLE grok",
		"MODE pty \"--no-auto-update\"",
		"MODE acp \"--no-auto-update\" \"agent\" \"stdio\"",
		"ENV XAI_API_KEY SECRET",
		"ENV GROK_HOME",
		"CAPABILITY subagents true",
		"CAPABILITY model_family multi",
	} {
		if !strings.Contains(upStr, want) {
			t.Errorf("up migration missing %q", want)
		}
	}
	if _, errs := parser.Parse(extractGrokAgentFile(t, upStr)); len(errs) > 0 {
		t.Fatalf("Grok AgentFile must parse: %v", errs)
	}

	down, err := FS.ReadFile("000191_add_grok_build_agent.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downStr := string(down)
	orgConfigsIdx := strings.Index(downStr, "DELETE FROM organization_agent_configs")
	orgAgentsIdx := strings.Index(downStr, "DELETE FROM organization_agents WHERE")
	userConfigsIdx := strings.Index(downStr, "DELETE FROM user_agent_configs")
	agentsIdx := strings.Index(downStr, "DELETE FROM agents WHERE slug")
	if orgConfigsIdx < 0 || orgAgentsIdx < 0 || userConfigsIdx < 0 || agentsIdx < 0 {
		t.Fatal("down migration must clear dependent rows before agents")
	}
	if orgConfigsIdx > agentsIdx || orgAgentsIdx > agentsIdx || userConfigsIdx > agentsIdx {
		t.Error("down migration must delete dependent rows before deleting agents")
	}
}

func extractGrokAgentFile(t *testing.T, migration string) string {
	t.Helper()
	start := strings.Index(migration, "E'")
	end := strings.LastIndex(migration, "'\n)")
	if start < 0 || end < 0 || end <= start+2 {
		t.Fatalf("migration does not contain expected E'...' AgentFile literal")
	}
	return strings.ReplaceAll(migration[start+2:end], `\n`, "\n")
}
