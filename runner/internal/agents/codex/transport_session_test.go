package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

func TestTransport_NewSession(t *testing.T) {
	stdoutPR, stdoutPW := io.Pipe()
	stdinPR, stdinPW := io.Pipe()
	defer stdoutPR.Close()
	defer stdinPR.Close()

	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.Initialize(ctx, stdinPW, stdoutPR, nil)
	go tr.ReadLoop(ctx)

	go func() {
		scanner := bufio.NewScanner(stdinPR)
		scanner.Scan()
		var request acp.JSONRPCRequest
		json.Unmarshal(scanner.Bytes(), &request)
		writeResponse(stdoutPW, request.ID, map[string]any{
			"thread": map[string]string{"id": "thread-abc"},
			"model":  "gpt-5.4",
		}, nil)
	}()

	sessionID, err := tr.NewSession("", nil)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if sessionID != "thread-abc" {
		t.Errorf("session_id = %q, want thread-abc", sessionID)
	}
	if tr.CurrentModel() != "gpt-5.4" {
		t.Errorf("current model = %q", tr.CurrentModel())
	}
}

func TestTransport_ResumeSession(t *testing.T) {
	stdoutPR, stdoutPW := io.Pipe()
	stdinPR, stdinPW := io.Pipe()
	defer stdoutPR.Close()
	defer stdinPR.Close()

	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.Initialize(ctx, stdinPW, stdoutPR, nil)
	go tr.ReadLoop(ctx)

	go func() {
		scanner := bufio.NewScanner(stdinPR)
		scanner.Scan()
		var request acp.JSONRPCRequest
		json.Unmarshal(scanner.Bytes(), &request)
		if request.Method != "thread/resume" {
			t.Errorf("method = %q, want thread/resume", request.Method)
		}
		writeResponse(stdoutPW, request.ID, map[string]any{
			"thread": map[string]string{"id": "thread-resumed"},
			"model":  "gpt-5.3-codex",
		}, nil)
	}()

	sessionID, err := tr.ResumeSession("", nil, "thread-old")
	if err != nil {
		t.Fatalf("ResumeSession: %v", err)
	}
	if sessionID != "thread-resumed" {
		t.Errorf("session_id = %q, want thread-resumed", sessionID)
	}
	if tr.CurrentModel() != "gpt-5.3-codex" {
		t.Errorf("current model = %q", tr.CurrentModel())
	}
}
