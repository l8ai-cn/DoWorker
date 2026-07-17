package acp

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

// Transport abstracts the wire protocol between ACPClient and an agent subprocess.
// Implementations register via RegisterTransport from their init() functions:
//   - ACPTransport: JSON-RPC 2.0
//   - agents/claude: Claude stream-json NDJSON protocol
//   - agents/codex: Codex app-server JSON-RPC protocol
type Transport interface {
	// Initialize wires the transport's I/O pipes.
	// Called BEFORE ReadLoop. Must not block on protocol messages.
	Initialize(ctx context.Context, stdin io.Writer, stdout io.Reader, stderr io.Reader) error

	// Handshake performs the protocol-specific initialization.
	// Called AFTER ReadLoop has been started in a goroutine.
	// Returns an auto-discovered sessionID (Claude) or empty string (ACP).
	Handshake(ctx context.Context) (string, error)

	// NewSession creates a new session, returning the sessionID.
	// cwd is the working directory for the session (required by standard ACP).
	// Claude transport returns the cached ID from Handshake.
	NewSession(cwd string, mcpServers map[string]any) (string, error)

	// SendPrompt delivers a prompt to the active session.
	SendPrompt(sessionID, prompt string) error

	// RespondToPermission responds to a permission request.
	// updatedInput is optional; when non-nil, it replaces the tool's original input.
	RespondToPermission(requestID string, approved bool, updatedInput map[string]any) error

	// CancelSession cancels the active session's processing.
	CancelSession(sessionID string) error

	// SendControlRequest sends an outgoing control_request to the agent CLI
	// and blocks until a control_response is received (or timeout).
	// Only supported by transports that implement bidirectional control protocol
	// (e.g., Claude stream-json). Others return ErrControlNotSupported.
	SendControlRequest(sessionID string, subtype string, payload map[string]any) (map[string]any, error)

	// SupportedPermissionModes returns the permission-mode wire values the agent
	// advertised via agentsmeshExtensions.permissionModes, or nil if it advertised
	// none (caller falls back to the Claude default set).
	SupportedPermissionModes() []string
	SupportedArtifactActions() []string

	// ReadLoop continuously reads messages from stdout and dispatches via callbacks.
	// Blocks until EOF or ctx cancellation.
	ReadLoop(ctx context.Context)

	// Close releases transport-internal resources.
	Close()
}

// --- Transport Registry ---

// TransportFactory creates a Transport with the given callbacks and logger.
type TransportFactory func(callbacks EventCallbacks, logger *slog.Logger) Transport

var (
	registryMu sync.RWMutex
	registry   = map[string]TransportFactory{}
)

// RegisterTransport registers a named transport factory.
// Typically called from init() in agent subpackages.
// Panics if the name is already registered (detects accidental duplicates).
func RegisterTransport(name string, factory TransportFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[name]; exists {
		panic("acp: duplicate transport registration: " + name)
	}
	registry[name] = factory
}

// NewTransport creates a transport by name from the registry.
func NewTransport(
	name string,
	callbacks EventCallbacks,
	logger *slog.Logger,
) (Transport, error) {
	registryMu.RLock()
	factory, ok := registry[name]
	registryMu.RUnlock()
	if ok {
		return factory(callbacks, logger), nil
	}
	return nil, fmt.Errorf("unknown ACP adapter %q", name)
}
