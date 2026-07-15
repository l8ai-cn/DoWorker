package codex

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

func TestHandlerCommandExecutionCarriesCommandEvidence(t *testing.T) {
	f := newFixture()
	defer f.Close()

	writeNotification(f.PW, "item/started", map[string]any{
		"item": map[string]any{
			"id": "command-1", "type": "commandExecution",
			"command": "node verify.js", "cwd": "/workspace",
		},
	})
	f.Drain()

	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ToolUpdates) != 1 {
		t.Fatalf("tool updates = %d, want 1", len(f.ToolUpdates))
	}
	var arguments map[string]any
	if err := json.Unmarshal([]byte(f.ToolUpdates[0].ArgumentsJSON), &arguments); err != nil {
		t.Fatalf("arguments JSON: %v", err)
	}
	if arguments["command"] != "node verify.js" || arguments["cwd"] != "/workspace" {
		t.Fatalf("arguments = %#v", arguments)
	}
}

func TestHandlerCommandExecutionUsesOutputDeltas(t *testing.T) {
	f := newFixture()
	defer f.Close()

	writeNotification(f.PW, "item/started", map[string]any{
		"item": map[string]any{
			"id": "command-2", "type": "commandExecution", "command": "node verify.js",
		},
	})
	writeNotification(f.PW, "item/commandExecution/outputDelta", map[string]any{
		"itemId": "command-2", "delta": "first line\n",
	})
	writeNotification(f.PW, "item/commandExecution/outputDelta", map[string]any{
		"itemId": "command-2", "delta": "second line\n",
	})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{
			"id": "command-2", "type": "commandExecution",
			"status": "completed", "exitCode": 0, "aggregatedOutput": nil,
		},
	})
	f.Drain()

	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ToolResults) != 1 {
		t.Fatalf("tool results = %d, want 1", len(f.ToolResults))
	}
	if f.ToolResults[0].ResultText != "first line\nsecond line\n" {
		t.Fatalf("result = %q", f.ToolResults[0].ResultText)
	}
}

func TestHandlerFileChangeCarriesPathsAndDiffs(t *testing.T) {
	f := newFixture()
	defer f.Close()
	changes := []map[string]any{{
		"path": "index.html",
		"kind": map[string]any{"type": "add"},
		"diff": "+<html></html>",
	}}

	writeNotification(f.PW, "item/started", map[string]any{
		"item": map[string]any{
			"id": "file-1", "type": "fileChange",
			"status": "inProgress", "changes": changes,
		},
	})
	writeNotification(f.PW, "item/completed", map[string]any{
		"item": map[string]any{
			"id": "file-1", "type": "fileChange",
			"status": "completed", "changes": changes,
		},
	})
	f.Drain()

	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ToolUpdates) != 1 {
		t.Fatalf("tool updates = %d, want 1", len(f.ToolUpdates))
	}
	var arguments struct {
		Changes []fileUpdateChange `json:"changes"`
	}
	if err := json.Unmarshal([]byte(f.ToolUpdates[0].ArgumentsJSON), &arguments); err != nil {
		t.Fatalf("arguments JSON: %v", err)
	}
	if len(arguments.Changes) != 1 ||
		arguments.Changes[0].Path != "index.html" ||
		arguments.Changes[0].Diff != "+<html></html>" {
		t.Fatalf("arguments = %#v", arguments)
	}
	if len(f.ToolResults) != 1 || !strings.Contains(f.ToolResults[0].ResultText, "index.html") {
		t.Fatalf("tool results = %+v", f.ToolResults)
	}
}

func TestToolOutputIsBoundedAndReportsTruncation(t *testing.T) {
	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	tr.appendToolOutput("large-command", strings.Repeat("x", maxToolOutputBytes+128))

	output := tr.takeToolOutput("large-command")

	if len(output) > maxToolOutputBytes+len(toolOutputTruncatedMarker) {
		t.Fatalf("output length = %d", len(output))
	}
	if !strings.HasSuffix(output, toolOutputTruncatedMarker) {
		t.Fatalf("output does not report truncation: %q", output[len(output)-80:])
	}
}

func TestTransportCloseClearsToolAndPermissionState(t *testing.T) {
	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	tr.appendToolOutput("command", "output")
	tr.rememberPermissionMethod("77", "item/tool/requestUserInput")

	tr.Close()

	if output := tr.takeToolOutput("command"); output != "" {
		t.Fatalf("tool output survived Close: %q", output)
	}
	if method := tr.takePermissionMethod("77"); method != "" {
		t.Fatalf("permission method survived Close: %q", method)
	}
}
