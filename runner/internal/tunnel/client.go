package tunnel

import (
	"context"
	"fmt"
	"log/slog"
	"net"
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

type Dispatcher interface {
	Dispatch(f tunnelframe.Frame)
	SetSender(send func(tunnelframe.Frame) error)
	Close()
}

type Client struct {
	gatewayURL string
	token      string
	runnerID   int64
	orgID      int64
	dispatcher Dispatcher

	conn     *websocket.Conn
	dialConn net.Conn
	connMu   sync.RWMutex
	writeMu  sync.Mutex

	connected    atomic.Bool
	stopped      atomic.Bool
	stopOnce     sync.Once
	lifecycleMu  sync.Mutex
	started      bool
	wg           sync.WaitGroup
	generation   uint64
	reconnect    reconnectPolicy
	readyTimeout time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
	logger       *slog.Logger
}

func NewClient(parentCtx context.Context, gatewayURL, token string, runnerID, orgID int64, dispatcher Dispatcher) *Client {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(parentCtx)
	c := &Client{
		gatewayURL:   gatewayURL,
		token:        token,
		runnerID:     runnerID,
		orgID:        orgID,
		dispatcher:   dispatcher,
		reconnect:    defaultReconnectPolicy(),
		readyTimeout: 10 * time.Second,
		ctx:          ctx,
		cancel:       cancel,
		logger:       slog.With("component", "tunnel_client", "runner_id", runnerID),
	}
	if dispatcher != nil {
		dispatcher.SetSender(func(tunnelframe.Frame) error {
			return fmt.Errorf("tunnel not connected")
		})
	}
	return c
}

func (c *Client) GatewayURL() string { return c.gatewayURL }

func (c *Client) UpdateToken(newToken string) {
	c.connMu.Lock()
	c.token = newToken
	c.connMu.Unlock()
}

func (c *Client) Connect() error { _, _, err := c.connectOnce(); return err }

func (c *Client) tunnelURL() (*url.URL, error) {
	u, err := url.Parse(c.gatewayURL)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway URL: %w", err)
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	const tunnelPath = "/runner/tunnel"
	if path.Clean(u.Path) == tunnelPath {
		u.Path = tunnelPath
	} else {
		u.Path = path.Join(u.Path, tunnelPath)
	}
	c.connMu.RLock()
	token := c.token
	c.connMu.RUnlock()
	query := u.Query()
	query.Set("token", token)
	u.RawQuery = query.Encode()
	return u, nil
}

func (c *Client) helloFrame() tunnelframe.Frame {
	hello := tunnelframe.HelloPayload{}
	if c.runnerID > 0 {
		hello.RunnerID = strconv.FormatInt(c.runnerID, 10)
	}
	if c.orgID > 0 {
		hello.OrgID = strconv.FormatInt(c.orgID, 10)
	}
	return tunnelframe.Frame{Type: tunnelframe.TypeHello, Payload: tunnelframe.EncodeJSON(hello)}
}

func (c *Client) Send(f tunnelframe.Frame) error {
	conn, generation := c.currentConnection()
	if conn == nil {
		return fmt.Errorf("tunnel not connected")
	}
	return c.sendOnConnection(conn, generation, f)
}

func (c *Client) sendOnConnection(conn *websocket.Conn, generation uint64, frame tunnelframe.Frame) error {
	c.connMu.RLock()
	isCurrent := c.conn == conn && c.generation == generation
	c.connMu.RUnlock()
	if !isCurrent {
		return fmt.Errorf("tunnel connection replaced")
	}
	return c.writeFrame(conn, frame)
}

func (c *Client) writeFrame(conn *websocket.Conn, frame tunnelframe.Frame) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(frame))
}

func (c *Client) Start() {
	c.lifecycleMu.Lock()
	defer c.lifecycleMu.Unlock()
	if c.stopped.Load() || c.started {
		return
	}
	c.started = true
	c.wg.Add(1)
	safego.Go("tunnel-reconnect", func() {
		defer c.wg.Done()
		c.connectionLoop()
	})
}

func (c *Client) IsConnected() bool { return c.connected.Load() }

func (c *Client) currentConnection() (*websocket.Conn, uint64) {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.conn, c.generation
}

func (c *Client) disconnect(conn *websocket.Conn, generation uint64) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.conn != conn || c.generation != generation {
		return
	}
	c.conn = nil
	c.connected.Store(false)
	_ = conn.Close()
}

func (c *Client) clearDialConnection(conn net.Conn) {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.dialConn == conn {
		c.dialConn = nil
	}
}

func nextBackoff(current, maximum time.Duration) time.Duration {
	if current >= maximum/2 {
		return maximum
	}
	return current * 2
}
