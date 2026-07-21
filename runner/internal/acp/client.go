package acp

import (
	"context"
	"log/slog"
	"sync"

	"github.com/l8ai-cn/agentcloud/runner/internal/processmgr"
)

// TransportType constants for ClientConfig.
// TransportTypeACP is the default protocol. Agent-specific transport types
// are defined in their respective packages under internal/agents/<name>/.
const (
	TransportTypeACP = "acp" // JSON-RPC 2.0 (Gemini, OpenCode, default)
)

// ClientConfig configures the ACP client.
type ClientConfig struct {
	Command       string
	Args          []string
	WorkDir       string
	Env           []string
	Logger        *slog.Logger
	Callbacks     EventCallbacks
	TransportType string // registered adapter id
	// Vendor session id to pass to session/resume instead of session/new.
	ResumeExternalSessionID string
}

// ACPClient manages an agent subprocess communicating via a pluggable
// Transport (JSON-RPC 2.0 or Claude stream-json).
type ACPClient struct {
	cfg       ClientConfig
	proc      processmgr.Handle
	transport Transport

	// State management
	state   string
	stateMu sync.RWMutex

	// Session tracking
	sessionID string
	sessionMu sync.RWMutex

	// Message history for snapshots
	messages    []ContentChunk
	messagesMu  sync.RWMutex
	maxMessages int

	// Tool call history for snapshots (keyed by tool_call_id)
	toolCalls   map[string]*ToolCallSnapshot
	toolCallsMu sync.RWMutex

	// Current plan for snapshots
	plan   []PlanStep
	planMu sync.RWMutex

	// Thinking history for snapshots (accumulator parallel to message history,
	// so late subscribers see prior thinking blocks, not just future incremental
	// events).
	thinkings    []ThinkingUpdate
	thinkingsMu  sync.RWMutex
	maxThinkings int

	// Log history for snapshots.
	logs    []LogEntry
	logsMu  sync.RWMutex
	maxLogs int

	// Loopal control-panel accumulator for loopal.snapshot on resubscribe.
	loopal *loopalState

	// Current configuration (permission_mode, model) for snapshots and broadcast.
	// Writes go through applyConfiguration (callback-wrap) or SeedConfiguration (init).
	configuration Configuration
	configMu      sync.RWMutex

	// Pending permission requests for snapshots
	pendingPerms   []PermissionRequest
	pendingPermsMu sync.RWMutex

	// Lifecycle
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	stopOnce sync.Once

	logger *slog.Logger
}

// NewClient creates an unstarted ACP client with the given configuration.
func NewClient(cfg ClientConfig) *ACPClient {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &ACPClient{
		cfg:          cfg,
		state:        StateUninitialized,
		messages:     make([]ContentChunk, 0, 256),
		maxMessages:  1000,
		thinkings:    make([]ThinkingUpdate, 0, 64),
		maxThinkings: 200,
		logs:         make([]LogEntry, 0, 64),
		maxLogs:      200,
		loopal:       newLoopalState(),
		toolCalls:    make(map[string]*ToolCallSnapshot),
		ctx:          ctx,
		cancel:       cancel,
		done:         make(chan struct{}),
		logger:       cfg.Logger.With("component", "acp-client"),
	}
}

// State returns the current client state.
func (c *ACPClient) State() string {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}

// ForceIdleIfBusy emits idle when a turn was interrupted by subprocess exit
// so downstream session publishers can flush buffered assistant text.
func (c *ACPClient) ForceIdleIfBusy() {
	c.stateMu.RLock()
	state := c.state
	c.stateMu.RUnlock()
	if state == StateProcessing || state == StateWaitingPermission {
		c.setState(StateIdle)
	}
}

func (c *ACPClient) setState(state string) {
	c.stateMu.Lock()
	old := c.state
	c.state = state
	c.stateMu.Unlock()

	if old != state && c.cfg.Callbacks.OnStateChange != nil {
		c.cfg.Callbacks.OnStateChange(state)
	}
}

// SessionID returns the current session ID.
func (c *ACPClient) SessionID() string {
	c.sessionMu.RLock()
	defer c.sessionMu.RUnlock()
	return c.sessionID
}
