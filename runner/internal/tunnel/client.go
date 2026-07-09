package tunnel

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/runner/internal/safego"
	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

// Dispatcher consumes inbound tunnel frames (REQ_START/BODY/END, CREDIT,
// STREAM_CANCEL, WS_DATA...) and drives local stream handling.
type Dispatcher interface {
	Dispatch(f tunnelframe.Frame)
	// SetSender wires the frame sink used to reply on streams.
	SetSender(send func(tunnelframe.Frame) error)
	// Close tears down all in-flight local streams.
	Close()
}

// Client is the runner-side outbound tunnel WebSocket client. It follows the
// same reconnect/backoff shape as runner/internal/relay but multiplexes HTTP
// requests via tunnelframe instead of the terminal protocol.
type Client struct {
	gatewayURL string
	token      string
	runnerID   int64
	orgID      int64
	dispatcher Dispatcher

	conn   *websocket.Conn
	connMu sync.RWMutex

	writeMu sync.Mutex

	connected atomic.Bool
	stopped   atomic.Bool
	stopCh    chan struct{}
	stopOnce  sync.Once

	ctx    context.Context
	cancel context.CancelFunc
	logger *slog.Logger
}

// NewClient creates a tunnel client. dispatcher may be nil (connect+hello only).
func NewClient(parentCtx context.Context, gatewayURL, token string, runnerID, orgID int64, dispatcher Dispatcher) *Client {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	c := &Client{
		gatewayURL: gatewayURL,
		token:      token,
		runnerID:   runnerID,
		orgID:      orgID,
		dispatcher: dispatcher,
		stopCh:     make(chan struct{}),
		ctx:        ctx,
		cancel:     cancel,
		logger:     slog.With("component", "tunnel_client", "runner_id", runnerID),
	}
	if dispatcher != nil {
		dispatcher.SetSender(c.Send)
	}
	return c
}

// GatewayURL returns the configured gateway URL.
func (c *Client) GatewayURL() string { return c.gatewayURL }

// UpdateToken swaps the JWT used on the next (re)connect.
func (c *Client) UpdateToken(newToken string) {
	c.connMu.Lock()
	c.token = newToken
	c.connMu.Unlock()
}

// Connect dials the gateway tunnel endpoint and sends the HELLO frame.
func (c *Client) Connect() error {
	c.connMu.RLock()
	token := c.token
	c.connMu.RUnlock()

	u, err := url.Parse(c.gatewayURL)
	if err != nil {
		return fmt.Errorf("invalid gateway URL: %w", err)
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	const tunnelPath = "/runner/tunnel"
	if path.Clean(u.Path) == tunnelPath {
		u.Path = tunnelPath
	} else {
		u.Path = path.Join(u.Path, tunnelPath)
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second, Proxy: http.ProxyFromEnvironment}
	conn, _, err := dialer.DialContext(c.ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("tunnel dial failed: %w", err)
	}
	conn.SetReadLimit(8 * 1024 * 1024)

	c.connMu.Lock()
	c.conn = conn
	c.connMu.Unlock()
	c.connected.Store(true)

	hello := tunnelframe.HelloPayload{}
	if c.runnerID > 0 {
		hello.RunnerID = strconv.FormatInt(c.runnerID, 10)
	}
	if c.orgID > 0 {
		hello.OrgID = strconv.FormatInt(c.orgID, 10)
	}
	if err := c.Send(tunnelframe.Frame{Type: tunnelframe.TypeHello, Payload: tunnelframe.EncodeJSON(hello)}); err != nil {
		_ = conn.Close()
		c.connected.Store(false)
		return fmt.Errorf("tunnel hello failed: %w", err)
	}
	c.logger.Info("tunnel connected", "host", u.Host)
	return nil
}

// Send writes a frame to the tunnel under the write mutex.
func (c *Client) Send(f tunnelframe.Frame) error {
	c.connMu.RLock()
	conn := c.conn
	c.connMu.RUnlock()
	if conn == nil {
		return fmt.Errorf("tunnel not connected")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(f))
}

// Start launches the read loop.
func (c *Client) Start() {
	if c.stopped.Load() {
		return
	}
	safego.Go("tunnel-read", c.readLoop)
}

func (c *Client) readLoop() {
	c.connMu.RLock()
	conn := c.conn
	c.connMu.RUnlock()
	if conn == nil {
		return
	}
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			c.connected.Store(false)
			c.logger.Debug("tunnel read loop exiting", "error", err)
			return
		}
		f, derr := tunnelframe.Decode(data)
		if derr != nil {
			continue
		}
		switch f.Type {
		case tunnelframe.TypePing:
			_ = c.Send(tunnelframe.Frame{Type: tunnelframe.TypePong})
		case tunnelframe.TypePong:
			// no-op
		default:
			if c.dispatcher != nil {
				c.dispatcher.Dispatch(f)
			}
		}
	}
}

// IsConnected reports connection state.
func (c *Client) IsConnected() bool { return c.connected.Load() }

// Stop closes the tunnel connection and dispatcher.
func (c *Client) Stop() {
	c.stopOnce.Do(func() {
		c.stopped.Store(true)
		c.connected.Store(false)
		close(c.stopCh)
		c.cancel()
		c.connMu.Lock()
		if c.conn != nil {
			_ = c.conn.Close()
		}
		c.connMu.Unlock()
		if c.dispatcher != nil {
			c.dispatcher.Close()
		}
	})
}
