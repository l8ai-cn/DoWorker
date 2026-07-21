package tunnel

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"github.com/l8ai-cn/agentcloud/relay/internal/protocol/tunnelframe"
)

// frameConn is the minimal connection surface a Tunnel needs. *websocket.Conn
// satisfies it; tests inject a mock.
type frameConn interface {
	WriteMessage(messageType int, data []byte) error
	ReadMessage() (messageType int, p []byte, err error)
	Close() error
}

const (
	defaultStreamWindow = 1 << 20
	pingInterval        = 10 * time.Second
	pongTimeout         = 3 * pingInterval
)

// Tunnel is a single runner<->gateway multiplexed connection. It owns the
// stream table and the read/write/heartbeat loops.
type Tunnel struct {
	RunnerID int64
	OrgID    int64

	conn    frameConn
	writeMu sync.Mutex

	window int

	mu      sync.Mutex
	streams map[uint32]*Stream
	nextID  uint32

	closed    chan struct{}
	closeOnce sync.Once

	lastPong atomic.Int64 // unix nano

	logger *slog.Logger
}

// NewTunnel wraps a websocket connection into a Tunnel.
func NewTunnel(conn *websocket.Conn, runnerID, orgID int64, window int, logger *slog.Logger) *Tunnel {
	return newTunnel(conn, runnerID, orgID, window, logger)
}

func newTunnel(conn frameConn, runnerID, orgID int64, window int, logger *slog.Logger) *Tunnel {
	if window <= 0 {
		window = defaultStreamWindow
	}
	if logger == nil {
		logger = slog.Default()
	}
	t := &Tunnel{
		RunnerID: runnerID,
		OrgID:    orgID,
		conn:     conn,
		window:   window,
		streams:  make(map[uint32]*Stream),
		closed:   make(chan struct{}),
		logger:   logger.With("component", "tunnel", "runner_id", runnerID),
	}
	t.lastPong.Store(time.Now().UnixNano())
	return t
}

// OpenStream allocates a new outbound stream and registers it.
func (t *Tunnel) OpenStream() *Stream {
	id := atomic.AddUint32(&t.nextID, 1)
	st := newStream(id, t.window)
	t.mu.Lock()
	t.streams[id] = st
	t.mu.Unlock()
	return st
}

func (t *Tunnel) getStream(id uint32) *Stream {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.streams[id]
}

// CloseStream removes and tears down a stream by id.
func (t *Tunnel) CloseStream(id uint32) {
	t.mu.Lock()
	st := t.streams[id]
	delete(t.streams, id)
	t.mu.Unlock()
	if st != nil {
		st.closeStream()
	}
}

// WriteFrame serializes and writes a frame under the write mutex.
func (t *Tunnel) WriteFrame(f tunnelframe.Frame) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	return t.conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(f))
}

// Start runs the read and heartbeat loops until the tunnel closes.
func (t *Tunnel) Start() {
	go t.pingLoop()
	t.readLoop()
}

func (t *Tunnel) readLoop() {
	defer t.Close()
	for {
		_, data, err := t.conn.ReadMessage()
		if err != nil {
			t.logger.Debug("tunnel read loop exiting", "error", err)
			return
		}
		f, derr := tunnelframe.Decode(data)
		if derr != nil {
			t.logger.Warn("bad frame", "error", derr)
			continue
		}
		t.dispatch(f)
	}
}

// dispatch routes an inbound frame. Connection-level frames (stream_id 0) are
// handled inline; stream frames are delivered to the owning stream's respCh.
func (t *Tunnel) dispatch(f tunnelframe.Frame) {
	if f.StreamID == 0 {
		switch f.Type {
		case tunnelframe.TypePing:
			_ = t.WriteFrame(tunnelframe.Frame{Type: tunnelframe.TypePong})
		case tunnelframe.TypePong:
			t.lastPong.Store(time.Now().UnixNano())
		case tunnelframe.TypeHello:
			// HELLO handled by the endpoint before Start(); ignore here.
		}
		return
	}
	st := t.getStream(f.StreamID)
	if st == nil {
		return
	}
	select {
	case st.respCh <- f:
	case <-t.closed:
	}
}

func (t *Tunnel) pingLoop() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-t.closed:
			return
		case <-ticker.C:
			if time.Since(time.Unix(0, t.lastPong.Load())) > pongTimeout {
				t.logger.Warn("tunnel heartbeat timeout, closing")
				t.Close()
				return
			}
			if err := t.WriteFrame(tunnelframe.Frame{Type: tunnelframe.TypePing}); err != nil {
				t.Close()
				return
			}
		}
	}
}

// Closed returns a channel closed when the tunnel is torn down.
func (t *Tunnel) Closed() <-chan struct{} { return t.closed }

// Close tears down the tunnel: closes the connection and drains all streams,
// delivering a synthetic RESP_ERROR so in-flight proxy requests see a clear 502.
func (t *Tunnel) Close() {
	t.closeOnce.Do(func() {
		close(t.closed)
		_ = t.conn.Close()
		t.mu.Lock()
		streams := t.streams
		t.streams = make(map[uint32]*Stream)
		t.mu.Unlock()
		errFrame := tunnelframe.Frame{
			Type:    tunnelframe.TypeRespError,
			Payload: tunnelframe.EncodeJSON(tunnelframe.RespErrorPayload{Code: "target_offline", Message: "tunnel closed"}),
		}
		for _, st := range streams {
			ef := errFrame
			ef.StreamID = st.ID
			select {
			case st.respCh <- ef:
			default:
			}
			st.closeStream()
		}
	})
}

// StreamCount returns the number of active streams (for stats/metrics).
func (t *Tunnel) StreamCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.streams)
}
