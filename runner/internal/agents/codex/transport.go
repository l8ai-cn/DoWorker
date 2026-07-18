package codex

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
)

// Transport implements acp.Transport for the Codex CLI app-server
// JSON-RPC 2.0 protocol (launched via `codex app-server`).
type transport struct {
	tracker   *acp.RequestTracker
	reader    *acp.Reader
	callbacks acp.EventCallbacks

	sessionID string
	sessionMu sync.RWMutex
	workDir   string
	turnID    string
	turnMu    sync.RWMutex

	streamMu            sync.Mutex
	streamedAgentMsgIDs map[string]struct{}
	streamedSinceMsg    bool
	messageBoundaryDue  bool
	toolMu              sync.Mutex
	toolOutputs         map[string]toolOutputBuffer
	permissionMu        sync.Mutex
	permissionMethods   map[string]string

	idleMu             sync.Mutex
	idleTimer          *time.Timer
	idleFallback       time.Duration
	hasLifecycleSignal bool

	ctx    context.Context
	logger *slog.Logger
}

// idleFallbackDelay is how long an agentMessage completion waits before it is
// treated as end-of-turn, for Codex builds that omit turn/completed. Any further
// turn activity cancels it, so a preamble message followed by tool calls does not
// end the turn prematurely.
const idleFallbackDelay = 4 * time.Second

func newTransport(callbacks acp.EventCallbacks, logger *slog.Logger) *transport {
	return &transport{
		callbacks:         callbacks,
		logger:            logger,
		idleFallback:      idleFallbackDelay,
		toolOutputs:       make(map[string]toolOutputBuffer),
		permissionMethods: make(map[string]string),
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
	params := initializeParams{
		ClientInfo: initializeClientInfo{
			Name:    "do-worker-runner",
			Version: "1.0.0",
		},
		Capabilities: initializeCapabilities{
			ExperimentalAPI: true,
		},
	}

	pr, err := t.tracker.SendRequest("initialize", params)
	if err != nil {
		return "", fmt.Errorf("write initialize: %w", err)
	}

	resp, err := t.tracker.WaitResponse(pr, 30*time.Second)
	if err != nil {
		return "", fmt.Errorf("wait initialize response: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("initialize error: code=%d msg=%s",
			resp.Error.Code, resp.Error.Message)
	}

	if err := t.tracker.Writer.WriteNotification("initialized", nil); err != nil {
		return "", fmt.Errorf("write initialized: %w", err)
	}

	t.logger.Info("Codex initialize succeeded")
	return "", nil
}

func (t *transport) SendPrompt(sessionID, prompt string) error {
	if sessionID == "" {
		return fmt.Errorf("thread id required")
	}
	params := mergeHeadlessFields(map[string]any{
		"threadId": sessionID,
		"input": []turnInput{{
			Type: "text",
			Text: prompt,
		}},
	}, t.workDir)

	pr, err := t.tracker.SendRequest("turn/start", params)
	if err != nil {
		return fmt.Errorf("write turn/start: %w", err)
	}

	go func() {
		resp, err := t.tracker.WaitResponse(pr, 5*time.Minute)
		if err != nil {
			t.logger.Error("turn/start response error", "error", err)
		} else if resp.Error != nil {
			t.logger.Error("turn/start error",
				"code", resp.Error.Code, "message", resp.Error.Message)
		}
	}()

	return nil
}

func (t *transport) CancelSession(sessionID string) error {
	params := turnInterruptParams{ThreadID: sessionID}
	pr, err := t.tracker.SendRequest("turn/interrupt", params)
	if err != nil {
		return fmt.Errorf("write turn/interrupt: %w", err)
	}
	go func() {
		t.tracker.WaitResponse(pr, 10*time.Second)
	}()
	return nil
}

func (t *transport) SendControlRequest(_ string, _ string, _ map[string]any) (map[string]any, error) {
	return nil, acp.ErrControlNotSupported
}

func (t *transport) SupportedPermissionModes() []string { return nil }
func (t *transport) SupportedArtifactActions() []string { return nil }

func (t *transport) ReadLoop(ctx context.Context) {
	for {
		msg, err := t.reader.ReadMessage()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				t.logger.Error("codex read error", "error", err)
				return
			}
		}
		t.dispatchMessage(msg)
	}
}

func (t *transport) Close() {
	t.cancelIdleFallback()
	t.clearToolOutputs()
	t.clearPermissionMethods()
}
