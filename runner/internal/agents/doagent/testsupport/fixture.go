package testsupport

import (
	_ "embed"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/llm_response.jsonl
var fixtureLLMResponse []byte

func BuildFixtureSandbox(t *testing.T) string {
	t.Helper()
	sandbox := t.TempDir()
	dir := filepath.Join(sandbox, "do-agent-home", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("doagent fixture: mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fixture-session.jsonl"), fixtureLLMResponse, 0o644); err != nil {
		t.Fatalf("doagent fixture: write: %v", err)
	}
	return sandbox
}
