package migrations

import (
	"strings"
	"testing"
)

// TestMigration000133AgentUsesLegacyColumns verifies that migration 000133
// adds the uses_legacy_columns column and backfills the Claude-family agents.
// This is a static SQL audit — it does not require running migrations against
// a database. The intent is to catch accidental edits to the embedded SQL.
func TestMigration000133AgentUsesLegacyColumns(t *testing.T) {
	up, err := FS.ReadFile("000133_agent_uses_legacy_columns.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upStr := string(up)

	if !strings.Contains(upStr, "ADD COLUMN uses_legacy_columns") {
		t.Error("up migration must add uses_legacy_columns column")
	}
	if !strings.Contains(upStr, "claude-code") {
		t.Error("up migration must backfill claude-code")
	}
	if !strings.Contains(upStr, "'claude'") {
		t.Error("up migration must backfill 'claude' (Claude family)")
	}
	if !strings.Contains(upStr, "uses_legacy_columns = TRUE") {
		t.Error("up migration must SET uses_legacy_columns = TRUE")
	}

	down, err := FS.ReadFile("000133_agent_uses_legacy_columns.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	if !strings.Contains(string(down), "DROP COLUMN") {
		t.Error("down migration must drop the column")
	}
}

// TestMigration000157CursorCLIAgent locks two failure modes in the
// cursor-cli builtin migration:
//   - AgentFile AGENT decl MUST be the actual binary name `cursor-agent`,
//     NOT the DB slug `cursor-cli`. agentfile/eval/eval_decl.go writes the
//     AGENT decl directly to LaunchCommand which the runner exec()s; using
//     the slug there makes every pod fail with ENOENT.
//   - Down migration MUST clear referencing rows first or it fails on the
//     organization_agents FK (000093 declares NO ACTION).
func TestMigration000157CursorCLIAgent(t *testing.T) {
	up, err := FS.ReadFile("000157_add_cursor_cli_agent.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upStr := string(up)

	if !strings.Contains(upStr, "AGENT cursor-agent") {
		t.Error("AgentFile AGENT must declare the binary name `cursor-agent` (not the slug)")
	}
	if strings.Contains(upStr, "AGENT cursor-cli") {
		t.Error("AGENT must NOT be the DB slug `cursor-cli` — that value becomes LaunchCommand and is exec'd; the binary is `cursor-agent`")
	}
	if !strings.Contains(upStr, "EXECUTABLE cursor-agent") {
		t.Error("EXECUTABLE must be `cursor-agent`")
	}
	if !strings.Contains(upStr, "'cursor-cli'") {
		t.Error("DB slug column must be 'cursor-cli' (matches claude-code / codex-cli / gemini-cli convention)")
	}
	if !strings.Contains(upStr, "'cursor-agent'") {
		t.Error("launch_command / executable columns must be 'cursor-agent'")
	}
	if !strings.Contains(upStr, "ENV CURSOR_API_KEY SECRET OPTIONAL") {
		t.Error("AgentFile must declare CURSOR_API_KEY as an optional secret so the credential UI can render a curated field")
	}

	down, err := FS.ReadFile("000157_add_cursor_cli_agent.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downStr := string(down)

	// Anchor on the full `DELETE FROM <table>` statement, NOT the bare table
	// name — the latter also appears in the explanatory comment block, so
	// indexing on it would assert comment-text ordering rather than statement
	// ordering (a stale lock-in that passes even when the SQL is reordered).
	// Both NO-ACTION-FK tables (000093) must be deleted BEFORE agents.
	orgConfigsIdx := strings.Index(downStr, "DELETE FROM organization_agent_configs")
	orgAgentsIdx := strings.Index(downStr, "DELETE FROM organization_agents WHERE")
	agentsIdx := strings.Index(downStr, "DELETE FROM agents WHERE slug")
	if orgConfigsIdx < 0 || orgAgentsIdx < 0 || agentsIdx < 0 {
		t.Fatal("down migration must DELETE from organization_agent_configs, organization_agents, AND agents")
	}
	if orgConfigsIdx > agentsIdx || orgAgentsIdx > agentsIdx {
		t.Error("down migration must clear organization_agent_configs/organization_agents BEFORE agents (FK NO ACTION blocks otherwise)")
	}

	// user_agent_credential_profiles was dropped in 000146 — a DELETE against
	// it would make any rollback fail with `relation does not exist`. Anchor
	// on the statement, not the bare name (the NOTE comment mentions it).
	if strings.Contains(downStr, "DELETE FROM user_agent_credential_profiles") {
		t.Error("down migration must NOT DELETE FROM user_agent_credential_profiles — dropped in 000146")
	}
}

func TestMigration000159DoAgentAgent(t *testing.T) {
	up, err := FS.ReadFile("000159_add_doagent_agent.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upStr := string(up)

	if !strings.Contains(upStr, "AGENT do-agent") {
		t.Error("AgentFile AGENT must declare the binary name `do-agent`")
	}
	if !strings.Contains(upStr, "EXECUTABLE do-agent") {
		t.Error("EXECUTABLE must be `do-agent`")
	}
	if !strings.Contains(upStr, "'do-agent'") {
		t.Error("DB slug / launch_command must be 'do-agent'")
	}
	if !strings.Contains(upStr, "MODE acp \"acp\"") {
		t.Error("AgentFile must declare ACP subcommand for acp mode")
	}
	if !strings.Contains(upStr, "ENV DO_AGENT_HOME") {
		t.Error("AgentFile must isolate home via DO_AGENT_HOME for settings copy")
	}
	if !strings.Contains(upStr, "ENV DO_AGENT_SETTINGS") {
		t.Error("AgentFile must isolate settings via DO_AGENT_SETTINGS")
	}
	if !strings.Contains(upStr, "ENV DO_AGENT_LOG_DIR") {
		t.Error("AgentFile must isolate JSONL logs via DO_AGENT_LOG_DIR")
	}
	if !strings.Contains(upStr, ".agent/config.json") {
		t.Error("AgentFile must write MCP overlay to workspace .agent/config.json")
	}

	down, err := FS.ReadFile("000159_add_doagent_agent.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downStr := string(down)

	orgConfigsIdx := strings.Index(downStr, "DELETE FROM organization_agent_configs")
	orgAgentsIdx := strings.Index(downStr, "DELETE FROM organization_agents WHERE")
	agentsIdx := strings.Index(downStr, "DELETE FROM agents WHERE slug")
	if orgConfigsIdx < 0 || orgAgentsIdx < 0 || agentsIdx < 0 {
		t.Fatal("down migration must DELETE from organization_agent_configs, organization_agents, AND agents")
	}
	if orgConfigsIdx > agentsIdx || orgAgentsIdx > agentsIdx {
		t.Error("down migration must clear org rows BEFORE agents (FK NO ACTION)")
	}
}

// TestMigration000160Coordinator locks the coordinator schema: the three tables
// the coordinator service depends on, the dedupe UNIQUE that keeps repeated
// scans idempotent, and the FK that cascades executions when a project is
// deleted. The down migration must drop executions BEFORE projects (FK order).
func TestMigration000160Coordinator(t *testing.T) {
	up, err := FS.ReadFile("000160_add_coordinator.up.sql")
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	upStr := string(up)

	for _, table := range []string{
		"CREATE TABLE coordinator_projects",
		"CREATE TABLE ticket_external_links",
		"CREATE TABLE coordinator_executions",
	} {
		if !strings.Contains(upStr, table) {
			t.Errorf("up migration must contain %q", table)
		}
	}
	if !strings.Contains(upStr, "UNIQUE (organization_id, platform_type, external_id)") {
		t.Error("ticket_external_links must dedupe on (organization_id, platform_type, external_id)")
	}
	if !strings.Contains(upStr, "REFERENCES coordinator_projects(id) ON DELETE CASCADE") {
		t.Error("coordinator_executions must cascade from coordinator_projects")
	}

	down, err := FS.ReadFile("000160_add_coordinator.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	downStr := string(down)
	execIdx := strings.Index(downStr, "DROP TABLE IF EXISTS coordinator_executions")
	projIdx := strings.Index(downStr, "DROP TABLE IF EXISTS coordinator_projects")
	if execIdx < 0 || projIdx < 0 {
		t.Fatal("down migration must drop coordinator_executions and coordinator_projects")
	}
	if execIdx > projIdx {
		t.Error("down migration must drop coordinator_executions BEFORE coordinator_projects (FK)")
	}
}
