package codex

import (
	"bufio"
	"encoding/json"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
)

func TestTransport_SetModel(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()
	fixture.transport.models = []string{"gpt-5.4", "gpt-5.3-codex"}
	fixture.transport.model = "gpt-5.4"

	go func() {
		scanner := bufio.NewScanner(fixture.StdinPR)
		scanner.Scan()
		var request acp.JSONRPCRequest
		json.Unmarshal(scanner.Bytes(), &request)
		if request.Method != "thread/settings/update" {
			t.Errorf("method = %q", request.Method)
		}
		var params threadSettingsUpdateParams
		json.Unmarshal(request.Params, &params)
		if params.ThreadID != "thread-1" || params.Model != "gpt-5.3-codex" {
			t.Errorf("params = %+v", params)
		}
		writeResponse(fixture.PW, request.ID, map[string]any{}, nil)
	}()

	result, err := fixture.transport.SendControlRequest("thread-1", "set_model", map[string]any{
		"model": "gpt-5.3-codex",
	})
	if err != nil {
		t.Fatalf("set_model: %v", err)
	}
	if result["model"] != "gpt-5.3-codex" || fixture.transport.CurrentModel() != "gpt-5.3-codex" {
		t.Fatalf("model was not updated: result=%v current=%q", result, fixture.transport.CurrentModel())
	}
	if _, err := fixture.transport.SendControlRequest("thread-1", "set_model", map[string]any{
		"model": "unknown",
	}); err == nil {
		t.Fatal("expected unsupported model error")
	}
}

func TestTransport_SendPromptUsesThreadModel(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()
	fixture.transport.models = []string{"gpt-5.4", "gpt-5.3-codex"}
	fixture.transport.model = "gpt-5.3-codex"

	received := make(chan acp.JSONRPCRequest, 1)
	go func() {
		scanner := bufio.NewScanner(fixture.StdinPR)
		scanner.Scan()
		var request acp.JSONRPCRequest
		json.Unmarshal(scanner.Bytes(), &request)
		received <- request
		writeResponse(fixture.PW, request.ID, map[string]any{}, nil)
	}()

	if err := fixture.transport.SendPrompt("thread-1", "hello codex"); err != nil {
		t.Fatalf("SendPrompt: %v", err)
	}
	select {
	case request := <-received:
		var params turnStartParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			t.Fatalf("decode turn/start: %v", err)
		}
		var raw map[string]any
		if err := json.Unmarshal(request.Params, &raw); err != nil {
			t.Fatalf("decode raw turn/start: %v", err)
		}
		if _, exists := raw["model"]; exists {
			t.Fatalf("turn/start overrides thread model: %v", raw["model"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for turn/start")
	}
}

func TestTransport_HandshakeAcceptsCatalogWithoutDefaultModel(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()

	go func() {
		scanner := bufio.NewScanner(fixture.StdinPR)
		scanner.Scan()
		var request acp.JSONRPCRequest
		json.Unmarshal(scanner.Bytes(), &request)
		writeResponse(fixture.PW, request.ID, map[string]any{}, nil)
		scanner.Scan()
		scanner.Scan()
		json.Unmarshal(scanner.Bytes(), &request)
		writeResponse(fixture.PW, request.ID, map[string]any{
			"data": []map[string]any{{
				"model": "gpt-5.4", "isDefault": false, "hidden": false,
			}},
			"nextCursor": nil,
		}, nil)
	}()

	if _, err := fixture.transport.Handshake(fixture.transport.ctx); err != nil {
		t.Fatalf("handshake: %v", err)
	}
}

func TestTransport_SetModelErrorKeepsCurrentModel(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()
	fixture.transport.models = []string{"gpt-5.4", "gpt-5.3-codex"}
	fixture.transport.model = "gpt-5.4"

	go func() {
		scanner := bufio.NewScanner(fixture.StdinPR)
		scanner.Scan()
		var request acp.JSONRPCRequest
		json.Unmarshal(scanner.Bytes(), &request)
		writeResponse(fixture.PW, request.ID, nil, &acp.JSONRPCError{
			Code: -32602, Message: "model unavailable",
		})
	}()

	_, err := fixture.transport.SendControlRequest(
		"thread-1",
		"set_model",
		map[string]any{"model": "gpt-5.3-codex"},
	)
	if err == nil {
		t.Fatal("expected thread/settings/update error")
	}
	if fixture.transport.CurrentModel() != "gpt-5.4" {
		t.Fatalf("current model changed after error: %q", fixture.transport.CurrentModel())
	}
}
