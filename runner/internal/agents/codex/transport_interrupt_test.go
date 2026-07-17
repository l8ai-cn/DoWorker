package codex

import (
	"bufio"
	"encoding/json"
	"testing"
	"time"
)

func TestTransportCancelSessionIncludesActiveTurn(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()
	fixture.transport.handleNotification("turn/started", mustMarshal(map[string]any{
		"threadId": "thread-1",
		"turn":     map[string]any{"id": "turn-7"},
	}))

	requests := make(chan json.RawMessage, 3)
	go func() {
		scanner := bufio.NewScanner(fixture.StdinPR)
		for scanner.Scan() {
			var request struct {
				ID     int64           `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			requireNoError(t, json.Unmarshal(scanner.Bytes(), &request))
			requests <- append(json.RawMessage(nil), scanner.Bytes()...)
			switch request.Method {
			case "turn/interrupt":
				writeResponse(fixture.PW, request.ID, map[string]any{}, nil)
			case "thread/backgroundTerminals/list":
				writeResponse(fixture.PW, request.ID, map[string]any{
					"data": []map[string]any{{"processId": "process-9"}},
				}, nil)
			case "thread/backgroundTerminals/terminate":
				writeResponse(fixture.PW, request.ID, map[string]any{"terminated": true}, nil)
				return
			}
		}
	}()

	requireNoError(t, fixture.transport.CancelSession("thread-1"))
	for index, expectedMethod := range []string{
		"turn/interrupt",
		"thread/backgroundTerminals/list",
		"thread/backgroundTerminals/terminate",
	} {
		select {
		case raw := <-requests:
			var message struct {
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			requireNoError(t, json.Unmarshal(raw, &message))
			if message.Method != expectedMethod {
				t.Fatalf("request %d method = %q", index, message.Method)
			}
			assertInterruptParams(t, expectedMethod, message.Params)
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for request %d", index)
		}
	}
}

func TestTransportTurnLifecyclePublishesProcessingThenIdle(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()

	fixture.transport.handleNotification("turn/started", mustMarshal(map[string]any{
		"turn": map[string]any{"id": "turn-7"},
	}))
	fixture.transport.handleNotification("turn/completed", mustMarshal(map[string]any{
		"turn": map[string]any{"id": "turn-7", "status": "completed"},
	}))

	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	if len(fixture.StateChanges) != 2 ||
		fixture.StateChanges[0] != "processing" ||
		fixture.StateChanges[1] != "idle" {
		t.Fatalf("state changes = %v", fixture.StateChanges)
	}
}

func assertInterruptParams(t *testing.T, method string, raw json.RawMessage) {
	t.Helper()
	switch method {
	case "turn/interrupt":
		var message struct {
			ThreadID string `json:"threadId"`
			TurnID   string `json:"turnId"`
		}
		requireNoError(t, json.Unmarshal(raw, &message))
		if message.ThreadID != "thread-1" || message.TurnID != "turn-7" {
			t.Fatalf("interrupt params = %+v", message)
		}
	case "thread/backgroundTerminals/terminate":
		var message backgroundTerminalTerminateParams
		requireNoError(t, json.Unmarshal(raw, &message))
		if message.ProcessID != "process-9" {
			t.Fatalf("terminate params = %+v", message)
		}
	}
}

func TestTransportCompletedTurnCannotBeInterrupted(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()
	fixture.transport.handleNotification("turn/started", mustMarshal(map[string]any{
		"turn": map[string]any{"id": "turn-7"},
	}))
	fixture.transport.handleNotification("turn/completed", mustMarshal(map[string]any{
		"turn": map[string]any{"id": "turn-7", "status": "completed"},
	}))

	go respondToBackgroundTerminalRequests(t, fixture, nil)
	if err := fixture.transport.CancelSession("thread-1"); err == nil {
		t.Fatal("expected completed turn without background terminals to fail")
	}
}

func TestTransportCancelSessionTerminatesBackgroundTerminalWithoutActiveTurn(t *testing.T) {
	fixture := newFixture()
	defer fixture.Close()

	go respondToBackgroundTerminalRequests(t, fixture, []string{"process-9"})
	requireNoError(t, fixture.transport.CancelSession("thread-1"))
}

func respondToBackgroundTerminalRequests(
	t *testing.T,
	fixture *testFixture,
	processIDs []string,
) {
	t.Helper()
	scanner := bufio.NewScanner(fixture.StdinPR)
	for scanner.Scan() {
		var request struct {
			ID     int64  `json:"id"`
			Method string `json:"method"`
		}
		requireNoError(t, json.Unmarshal(scanner.Bytes(), &request))
		switch request.Method {
		case "thread/backgroundTerminals/list":
			data := make([]map[string]any, 0, len(processIDs))
			for _, processID := range processIDs {
				data = append(data, map[string]any{"processId": processID})
			}
			writeResponse(fixture.PW, request.ID, map[string]any{"data": data}, nil)
			if len(processIDs) == 0 {
				return
			}
		case "thread/backgroundTerminals/terminate":
			writeResponse(fixture.PW, request.ID, map[string]any{"terminated": true}, nil)
			return
		default:
			t.Errorf("unexpected request method %q", request.Method)
			return
		}
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
