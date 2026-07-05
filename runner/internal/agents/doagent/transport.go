package doagent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

const TransportType = "do-agent"

type transport struct {
	tracker   *acp.RequestTracker
	reader    *acp.Reader
	handler   *acp.Handler
	callbacks acp.EventCallbacks
	permModes []string

	ctx    context.Context
	logger *slog.Logger
}

func newTransport(callbacks acp.EventCallbacks, logger *slog.Logger) *transport {
	return &transport{
		handler:   acp.NewHandler(callbacks, logger),
		callbacks: callbacks,
		logger:    logger,
		permModes: []string{"allow", "restricted"},
	}
}

func (t *transport) Initialize(ctx context.Context, stdin io.Writer, stdout io.Reader, _ io.Reader) error {
	t.ctx = ctx
	writer := acp.NewWriter(stdin)
	t.reader = acp.NewReader(stdout, t.logger)
	t.tracker = acp.NewRequestTracker(writer, t.logger, func() <-chan struct{} { return ctx.Done() })
	return nil
}

func (t *transport) Handshake(_ context.Context) (string, error) {
	params := map[string]any{
		"protocolVersion": 1,
		"clientInfo": map[string]any{
			"name":    "do-worker-runner",
			"version": "1.0.0",
		},
		"clientCapabilities": map[string]any{},
	}

	pr, err := t.tracker.SendRequest("initialize", params)
	if err != nil {
		return "", fmt.Errorf("write initialize: %w", err)
	}

	resp, err := t.tracker.WaitResponse(pr, rpcTimeout)
	if err != nil {
		return "", fmt.Errorf("wait initialize response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("initialize error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
	}

	t.logger.Info("do-agent ACP initialize succeeded")
	return "", nil
}

func (t *transport) ReadLoop(ctx context.Context) {
	for {
		msg, err := t.reader.ReadMessage()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				t.logger.Error("do-agent read error", "error", err)
				return
			}
		}
		t.dispatchMessage(msg)
	}
}

func (t *transport) Close() {}

func (t *transport) SupportedPermissionModes() []string {
	return t.permModes
}

func (t *transport) dispatchMessage(msg *acp.JSONRPCMessage) {
	switch {
	case msg.IsResponse():
		t.tracker.HandleResponse(msg)
	case msg.IsNotification():
		if msg.Method == "permission.updated" {
			t.handlePermissionUpdated(msg.Params)
			return
		}
		t.handler.HandleNotification(msg.Method, msg.Params)
	case msg.IsRequest():
		if msg.Method == "session/request_permission" {
			id, _ := msg.GetID()
			t.handler.HandlePermissionRequest(id, msg.Params)
			return
		}
		t.tracker.RejectRequest(msg)
	}
}

func (t *transport) callRPC(method string, params map[string]any) (map[string]any, error) {
	pr, err := t.tracker.SendRequest(method, params)
	if err != nil {
		return nil, fmt.Errorf("write %s: %w", method, err)
	}
	resp, err := t.tracker.WaitResponse(pr, rpcTimeout)
	if err != nil {
		return nil, fmt.Errorf("wait %s response: %w", method, err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("%s error: code=%d msg=%s", method, resp.Error.Code, resp.Error.Message)
	}
	if len(resp.Result) == 0 {
		return map[string]any{}, nil
	}
	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse %s result: %w", method, err)
	}
	return result, nil
}
