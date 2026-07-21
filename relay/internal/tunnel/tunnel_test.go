package tunnel

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/relay/internal/protocol/tunnelframe"
)

func mustJSON(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// mockConn is a no-op frameConn used to construct a Tunnel without a real
// websocket. Reads block forever (until closed), writes are captured.
type mockConn struct {
	writes  chan []byte
	closeCh chan struct{}
}

func newMockConn() *mockConn {
	return &mockConn{writes: make(chan []byte, 64), closeCh: make(chan struct{})}
}

func (m *mockConn) WriteMessage(_ int, data []byte) error {
	select {
	case m.writes <- data:
	default:
	}
	return nil
}

func (m *mockConn) ReadMessage() (int, []byte, error) {
	<-m.closeCh
	return 0, nil, errClosed
}

func (m *mockConn) Close() error {
	select {
	case <-m.closeCh:
	default:
		close(m.closeCh)
	}
	return nil
}

var errClosed = &connClosedError{}

type connClosedError struct{}

func (*connClosedError) Error() string { return "closed" }

func newTunnelForTest(conn frameConn) *Tunnel {
	return newTunnel(conn, 7, 0, 1<<20, nil)
}

func TestTunnel_OpenStreamRoutesResponse(t *testing.T) {
	tun := newTunnelForTest(newMockConn())

	st := tun.OpenStream()
	tun.dispatch(tunnelframe.Frame{Type: tunnelframe.TypeRespStart, StreamID: st.ID,
		Payload: mustJSON(tunnelframe.RespStartPayload{Status: 200})})

	select {
	case f := <-st.respCh:
		if f.Type != tunnelframe.TypeRespStart {
			t.Fatalf("unexpected %v", f.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("response not routed to stream")
	}
}

func TestTunnel_CloseDrainsStreams(t *testing.T) {
	tun := newTunnelForTest(newMockConn())
	st := tun.OpenStream()
	tun.Close()
	select {
	case f := <-st.respCh:
		if f.Type != tunnelframe.TypeRespError {
			t.Fatalf("expected synthetic RESP_ERROR, got %v", f.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("stream not drained on close")
	}
}
