package extension

import (
	"encoding/json"
	"testing"
)

// ---------------------------------------------------------------------------
// TableName() methods
// ---------------------------------------------------------------------------

func TestTableNames(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		want      string
	}{
		{"InstalledMcpServer", InstalledMcpServer{}.TableName(), "installed_mcp_servers"},
		{"InstalledSkill", InstalledSkill{}.TableName(), "installed_skills"},
		{"McpMarketItem", McpMarketItem{}.TableName(), "mcp_market_items"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tableName != tt.want {
				t.Errorf("TableName() = %q, want %q", tt.tableName, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// McpMarketItem.GetAgentFilter
// ---------------------------------------------------------------------------

func TestMcpMarketItem_GetAgentFilter_Valid(t *testing.T) {
	item := McpMarketItem{
		AgentFilter: json.RawMessage(`["claude-code","aider"]`),
	}
	result := item.GetAgentFilter()
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result[0] != "claude-code" {
		t.Errorf("expected first item 'claude-code', got %q", result[0])
	}
	if result[1] != "aider" {
		t.Errorf("expected second item 'aider', got %q", result[1])
	}
}

func TestMcpMarketItem_GetAgentFilter_Empty(t *testing.T) {
	item := McpMarketItem{
		AgentFilter: nil,
	}
	result := item.GetAgentFilter()
	if result != nil {
		t.Errorf("expected nil for empty filter, got %v", result)
	}

	item2 := McpMarketItem{
		AgentFilter: json.RawMessage{},
	}
	result2 := item2.GetAgentFilter()
	if result2 != nil {
		t.Errorf("expected nil for empty RawMessage, got %v", result2)
	}
}

func TestMcpMarketItem_GetAgentFilter_Invalid(t *testing.T) {
	item := McpMarketItem{
		AgentFilter: json.RawMessage(`{invalid json`),
	}
	result := item.GetAgentFilter()
	if result != nil {
		t.Errorf("expected nil for invalid JSON, got %v", result)
	}
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestConstants(t *testing.T) {
	if ScopeOrg != "org" {
		t.Errorf("ScopeOrg = %q, want %q", ScopeOrg, "org")
	}
	if ScopeUser != "user" {
		t.Errorf("ScopeUser = %q, want %q", ScopeUser, "user")
	}

	if InstallSourceCatalog != "catalog" {
		t.Errorf("InstallSourceCatalog = %q, want %q", InstallSourceCatalog, "catalog")
	}
	if InstallSourceGitHub != "github" {
		t.Errorf("InstallSourceGitHub = %q, want %q", InstallSourceGitHub, "github")
	}
	if InstallSourceUpload != "upload" {
		t.Errorf("InstallSourceUpload = %q, want %q", InstallSourceUpload, "upload")
	}

	if TransportTypeStdio != "stdio" {
		t.Errorf("TransportTypeStdio = %q, want %q", TransportTypeStdio, "stdio")
	}
	if TransportTypeHTTP != "http" {
		t.Errorf("TransportTypeHTTP = %q, want %q", TransportTypeHTTP, "http")
	}
	if TransportTypeSSE != "sse" {
		t.Errorf("TransportTypeSSE = %q, want %q", TransportTypeSSE, "sse")
	}
}
