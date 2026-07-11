package acp

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestACPTransportHandshakeRunsHookAfterInitialize(t *testing.T) {
	toAgentReader, toAgentWriter := io.Pipe()
	fromAgentReader, fromAgentWriter := io.Pipe()
	defer toAgentReader.Close()
	defer toAgentWriter.Close()
	defer fromAgentReader.Close()
	defer fromAgentWriter.Close()

	hookCalled := make(chan json.RawMessage, 1)
	transport := NewACPTransportWithHandshakeHook(EventCallbacks{}, slog.Default(),
		func(requester HandshakeRequester, initResult json.RawMessage) error {
			hookCalled <- initResult
			_, err := requester.Request("authenticate", map[string]string{"methodId": "xai.api_key"})
			return err
		},
	)
	if err := transport.Initialize(context.Background(), toAgentWriter, fromAgentReader, nil); err != nil {
		t.Fatalf("initialize transport: %v", err)
	}
	go transport.ReadLoop(context.Background())

	agentDone := make(chan error, 1)
	go func() {
		reader := NewReader(toAgentReader, slog.Default())
		writer := NewWriter(fromAgentWriter)

		initialize, err := reader.ReadMessage()
		if err != nil {
			agentDone <- err
			return
		}
		initializeID, _ := initialize.GetID()
		if err := writer.WriteResponse(initializeID, map[string]any{
			"authMethods": []map[string]string{{"id": "xai.api_key"}},
		}, nil); err != nil {
			agentDone <- err
			return
		}

		authenticate, err := reader.ReadMessage()
		if err != nil {
			agentDone <- err
			return
		}
		if authenticate.Method != "authenticate" {
			agentDone <- &unexpectedMethodError{got: authenticate.Method}
			return
		}
		authenticateID, _ := authenticate.GetID()
		agentDone <- writer.WriteResponse(authenticateID, map[string]any{}, nil)
	}()

	if _, err := transport.Handshake(context.Background()); err != nil {
		t.Fatalf("handshake: %v", err)
	}
	select {
	case result := <-hookCalled:
		if string(result) != `{"authMethods":[{"id":"xai.api_key"}]}` {
			t.Fatalf("hook result = %s", result)
		}
	case <-time.After(time.Second):
		t.Fatal("handshake hook was not called")
	}
	if err := <-agentDone; err != nil {
		t.Fatalf("agent exchange: %v", err)
	}
}

type unexpectedMethodError struct {
	got string
}

func (e *unexpectedMethodError) Error() string {
	return "unexpected method: " + e.got
}
