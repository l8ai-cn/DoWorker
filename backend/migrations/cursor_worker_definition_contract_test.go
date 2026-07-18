package migrations

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCursorWorkerDefinitionMatchesMigrationAndRunnerContract(t *testing.T) {
	raw, err := os.ReadFile("../../config/worker-types/cursor-cli/definition.json")
	if err != nil {
		t.Fatalf("read cursor definition: %v", err)
	}
	var definition struct {
		Slug             string   `json:"slug"`
		Executable       string   `json:"executable"`
		AdapterID        string   `json:"adapter_id"`
		InteractionModes []string `json:"interaction_modes"`
	}
	if err := json.Unmarshal(raw, &definition); err != nil {
		t.Fatalf("decode cursor definition: %v", err)
	}
	if definition.Slug != "cursor-cli" ||
		definition.Executable != "agent" ||
		definition.AdapterID != "cursor-acp" {
		t.Fatalf("unexpected cursor definition: %+v", definition)
	}
	if strings.Join(definition.InteractionModes, ",") != "pty,acp" {
		t.Fatalf("unexpected cursor interaction modes: %v", definition.InteractionModes)
	}

	agentFile, err := os.ReadFile("../../config/worker-types/cursor-cli/AgentFile")
	if err != nil {
		t.Fatalf("read cursor AgentFile: %v", err)
	}
	for _, fragment := range []string{"AGENT agent", "EXECUTABLE agent", "MODE acp \"acp\""} {
		if !strings.Contains(string(agentFile), fragment) {
			t.Errorf("cursor AgentFile must contain %q", fragment)
		}
	}
}
