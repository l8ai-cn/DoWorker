package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

func TestTransport_Handshake(t *testing.T) {
	stdoutPR, stdoutPW := io.Pipe()
	stdinPR, stdinPW := io.Pipe()
	defer stdoutPR.Close()
	defer stdinPR.Close()

	tr := newTransport(acp.EventCallbacks{}, slog.Default())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.Initialize(ctx, stdinPW, stdoutPR, nil)
	go tr.ReadLoop(ctx)

	go func() {
		scanner := bufio.NewScanner(stdinPR)
		scanner.Scan()
		var req acp.JSONRPCRequest
		json.Unmarshal(scanner.Bytes(), &req)
		writeResponse(stdoutPW, req.ID, map[string]any{"server_info": map[string]string{"name": "codex"}}, nil)
		scanner.Scan()
		scanner.Scan()
		json.Unmarshal(scanner.Bytes(), &req)
		writeModelListResponse(stdoutPW, req.ID)
	}()

	sid, err := tr.Handshake(ctx)
	if err != nil {
		t.Fatalf("Handshake: %v", err)
	}
	if sid != "" {
		t.Errorf("expected empty session_id, got %q", sid)
	}
	if got := tr.SupportedModels(); len(got) != 2 || got[0] != "gpt-5.4" {
		t.Errorf("supported models = %v", got)
	}
	if got := tr.CurrentModel(); got != "" {
		t.Errorf("current model = %q", got)
	}
}

func TestTransport_Handshake_Error(t *testing.T) {
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
		var req acp.JSONRPCRequest
		json.Unmarshal(scanner.Bytes(), &req)
		writeResponse(stdoutPW, req.ID, nil, &acp.JSONRPCError{Code: -32600, Message: "bad request"})
		io.Copy(io.Discard, stdinPR)
	}()

	_, err := tr.Handshake(ctx)
	if err == nil {
		t.Fatal("expected error from handshake")
	}
}

func TestTransport_SendPrompt(t *testing.T) {
	stdoutPR, stdoutPW := io.Pipe()
	stdinPR, stdinPW := io.Pipe()
	defer stdoutPR.Close()
	defer stdinPR.Close()

	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.Initialize(ctx, stdinPW, stdoutPR, nil)
	go tr.ReadLoop(ctx)

	received := make(chan acp.JSONRPCRequest, 1)
	go func() {
		scanner := bufio.NewScanner(stdinPR)
		scanner.Scan()
		var req acp.JSONRPCRequest
		json.Unmarshal(scanner.Bytes(), &req)
		received <- req
		writeResponse(stdoutPW, req.ID, map[string]any{}, nil)
	}()

	if err := tr.SendPrompt("thread-1", "hello codex"); err != nil {
		t.Fatalf("SendPrompt: %v", err)
	}

	select {
	case req := <-received:
		if req.Method != "turn/start" {
			t.Errorf("method = %q, want turn/start", req.Method)
		}
		var params turnStartParams
		json.Unmarshal(req.Params, &params)
		if params.ThreadID != "thread-1" {
			t.Errorf("threadId = %q", params.ThreadID)
		}
		if len(params.Input) == 0 || params.Input[0].Text != "hello codex" {
			t.Errorf("input = %+v", params.Input)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for turn/start")
	}
}

func TestTransport_SendPrompt_EmptyThreadID(t *testing.T) {
	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	if err := tr.SendPrompt("", "hello"); err == nil {
		t.Fatal("expected error for empty thread id")
	}
}

func TestTransport_RespondToPermission(t *testing.T) {
	stdoutPR, _ := io.Pipe()
	stdinPR, stdinPW := io.Pipe()
	defer stdoutPR.Close()
	defer stdinPR.Close()

	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.Initialize(ctx, stdinPW, stdoutPR, nil)

	received := make(chan json.RawMessage, 1)
	go func() {
		scanner := bufio.NewScanner(stdinPR)
		scanner.Scan()
		received <- json.RawMessage(scanner.Bytes())
	}()

	if err := tr.RespondToPermission("42", true, nil); err != nil {
		t.Fatalf("RespondToPermission: %v", err)
	}

	select {
	case raw := <-received:
		var msg struct {
			ID     *int64          `json:"id"`
			Result json.RawMessage `json:"result"`
		}
		json.Unmarshal(raw, &msg)
		if msg.ID == nil || *msg.ID != 42 {
			t.Errorf("expected id=42, got %v", msg.ID)
		}
		var result map[string]any
		json.Unmarshal(msg.Result, &result)
		if result["decision"] != "accept" {
			t.Errorf("decision = %v, want accept", result["decision"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestTransport_RespondToPermission_Decline(t *testing.T) {
	stdoutPR, _ := io.Pipe()
	stdinPR, stdinPW := io.Pipe()
	defer stdoutPR.Close()
	defer stdinPR.Close()

	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.Initialize(ctx, stdinPW, stdoutPR, nil)

	received := make(chan json.RawMessage, 1)
	go func() {
		scanner := bufio.NewScanner(stdinPR)
		scanner.Scan()
		received <- json.RawMessage(scanner.Bytes())
	}()

	if err := tr.RespondToPermission("7", false, nil); err != nil {
		t.Fatalf("RespondToPermission: %v", err)
	}

	select {
	case raw := <-received:
		var msg struct {
			ID     *int64          `json:"id"`
			Result json.RawMessage `json:"result"`
		}
		json.Unmarshal(raw, &msg)
		if msg.ID == nil || *msg.ID != 7 {
			t.Errorf("expected id=7, got %v", msg.ID)
		}
		var result map[string]any
		json.Unmarshal(msg.Result, &result)
		if result["decision"] != "decline" {
			t.Errorf("decision = %v, want decline", result["decision"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestTransport_RespondToUserInputUsesCodexAnswerShape(t *testing.T) {
	stdoutPR, _ := io.Pipe()
	stdinPR, stdinPW := io.Pipe()
	defer stdoutPR.Close()
	defer stdinPR.Close()

	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.Initialize(ctx, stdinPW, stdoutPR, nil)
	tr.rememberPermissionMethod("42", "item/tool/requestUserInput")

	received := make(chan json.RawMessage, 1)
	go func() {
		scanner := bufio.NewScanner(stdinPR)
		scanner.Scan()
		received <- json.RawMessage(scanner.Bytes())
	}()

	err := tr.RespondToPermission("42", true, map[string]any{
		"answers": map[string]any{
			"framework": []any{"React"},
			"features":  []any{"Auth", "Cache"},
		},
	})
	if err != nil {
		t.Fatalf("RespondToPermission: %v", err)
	}

	select {
	case raw := <-received:
		var msg struct {
			Result struct {
				Answers map[string]struct {
					Answers []string `json:"answers"`
				} `json:"answers"`
			} `json:"result"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("response JSON: %v", err)
		}
		if got := msg.Result.Answers["framework"].Answers; len(got) != 1 || got[0] != "React" {
			t.Fatalf("framework answers = %#v", got)
		}
		if got := msg.Result.Answers["features"].Answers; len(got) != 2 || got[1] != "Cache" {
			t.Fatalf("feature answers = %#v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestTransport_RespondToMcpElicitationUsesActionAndContent(t *testing.T) {
	stdoutPR, _ := io.Pipe()
	stdinPR, stdinPW := io.Pipe()
	defer stdoutPR.Close()
	defer stdinPR.Close()

	tr := newTransport(acp.EventCallbacks{}, discardLogger())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr.Initialize(ctx, stdinPW, stdoutPR, nil)
	tr.rememberPermissionMethod("43", "mcpServer/elicitation/request")

	received := make(chan json.RawMessage, 1)
	go func() {
		scanner := bufio.NewScanner(stdinPR)
		scanner.Scan()
		received <- json.RawMessage(scanner.Bytes())
	}()

	if err := tr.RespondToPermission("43", true, map[string]any{
		"answers": map[string]any{
			"project": []any{"Agent Cloud"},
			"targets": []any{"web", "embed"},
		},
	}); err != nil {
		t.Fatalf("RespondToPermission: %v", err)
	}

	select {
	case raw := <-received:
		var msg struct {
			Result map[string]any `json:"result"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("response JSON: %v", err)
		}
		if msg.Result["action"] != "accept" {
			t.Fatalf("action = %#v", msg.Result["action"])
		}
		content, ok := msg.Result["content"].(map[string]any)
		if !ok || content["project"] != "Agent Cloud" {
			t.Fatalf("content = %#v", msg.Result["content"])
		}
		targets, ok := content["targets"].([]any)
		if !ok || len(targets) != 2 || targets[1] != "embed" {
			t.Fatalf("targets = %#v", content["targets"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestTransport_CancelSessionRequiresActiveTurn(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()
	go respondToBackgroundTerminalRequests(t, fixture, nil)
	if err := fixture.transport.CancelSession("thread-1"); err == nil {
		t.Fatal("expected missing active turn error")
	}
}

func TestTransport_Close(t *testing.T) {
	newTransport(acp.EventCallbacks{}, discardLogger()).Close()
}
