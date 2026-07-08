package tunnel

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

// fakeFrameSink collects RESP_*/WS_* frames emitted by serveLocalHTTP /
// serveLocalWebSocket.
type fakeFrameSink struct {
	mu         sync.Mutex
	statusCode int
	buf        bytes.Buffer
	errCode    string
	wsOut      []tunnelframe.WSDataPayload
	wsOutCh    chan tunnelframe.WSDataPayload
}

func (f *fakeFrameSink) Send(fr tunnelframe.Frame) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch fr.Type {
	case tunnelframe.TypeRespStart:
		var rs tunnelframe.RespStartPayload
		_ = json.Unmarshal(fr.Payload, &rs)
		f.statusCode = rs.Status
	case tunnelframe.TypeRespBody:
		f.buf.Write(fr.Payload)
	case tunnelframe.TypeRespError:
		var re tunnelframe.RespErrorPayload
		_ = json.Unmarshal(fr.Payload, &re)
		f.errCode = re.Code
	case tunnelframe.TypeWSData:
		var wd tunnelframe.WSDataPayload
		_ = json.Unmarshal(fr.Payload, &wd)
		f.wsOut = append(f.wsOut, wd)
		if f.wsOutCh != nil {
			f.wsOutCh <- wd
		}
	}
	return nil
}

func (f *fakeFrameSink) status() int   { f.mu.Lock(); defer f.mu.Unlock(); return f.statusCode }
func (f *fakeFrameSink) body() string  { f.mu.Lock(); defer f.mu.Unlock(); return f.buf.String() }
func (f *fakeFrameSink) error() string { f.mu.Lock(); defer f.mu.Unlock(); return f.errCode }

func newFakeFrameSink() *fakeFrameSink {
	return &fakeFrameSink{wsOutCh: make(chan tunnelframe.WSDataPayload, 32)}
}

func TestServeLocalHTTP_RejectsNonLoopback(t *testing.T) {
	if err := validateTarget("10.0.0.5:80"); err == nil {
		t.Fatal("non-loopback must be rejected")
	}
	if err := validateTarget("127.0.0.1:3000"); err != nil {
		t.Fatalf("loopback should pass: %v", err)
	}
	if err := validateTarget("localhost:8080"); err != nil {
		t.Fatalf("localhost should pass: %v", err)
	}
	if err := validateTarget("[::1]:8080"); err != nil {
		t.Fatalf("ipv6 loopback should pass: %v", err)
	}
}

func TestServeLocalHTTP_StreamsUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, "pong")
	}))
	defer upstream.Close()
	target := strings.TrimPrefix(upstream.URL, "http://") // 127.0.0.1:port

	fc := newFakeFrameSink()
	serveLocalHTTP(context.Background(), fc, 1, tunnelframe.ReqStartPayload{
		Method: "GET", Path: "/", Target: target,
	}, nil, newCreditWindow(1<<20))

	if fc.status() != 200 || fc.body() != "pong" {
		t.Fatalf("bad: %d %q", fc.status(), fc.body())
	}
}

func TestServeLocalHTTP_NonLoopbackEmitsError(t *testing.T) {
	fc := newFakeFrameSink()
	serveLocalHTTP(context.Background(), fc, 1, tunnelframe.ReqStartPayload{
		Method: "GET", Path: "/", Target: "10.0.0.5:80",
	}, nil, newCreditWindow(1<<20))
	if fc.error() == "" {
		t.Fatal("expected RESP_ERROR for non-loopback target")
	}
}

func TestServeLocalWebSocket_EchoesMessages(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(mt, data); err != nil {
				return
			}
		}
	}))
	defer upstream.Close()
	target := strings.TrimPrefix(upstream.URL, "http://")

	fc := newFakeFrameSink()
	wsIn := make(chan tunnelframe.Frame, 8)
	done := make(chan struct{})
	go func() {
		defer close(done)
		serveLocalWebSocket(context.Background(), fc, 1, tunnelframe.ReqStartPayload{
			Method: "GET", Path: "/", Target: target, IsWebSocket: true,
		}, wsIn, newCreditWindow(1<<20))
	}()

	// Wait for RESP_START (status 101) before sending WS_DATA.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && fc.status() == 0 {
		time.Sleep(5 * time.Millisecond)
	}
	if fc.status() != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101, got %d", fc.status())
	}

	wsIn <- tunnelframe.Frame{
		Type:    tunnelframe.TypeWSData,
		Payload: tunnelframe.EncodeJSON(tunnelframe.WSDataPayload{MessageType: websocket.TextMessage, Data: []byte("hello")}),
	}

	select {
	case echoed := <-fc.wsOutCh:
		if echoed.MessageType != websocket.TextMessage || string(echoed.Data) != "hello" {
			t.Fatalf("unexpected echo: %+v", echoed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for echoed WS_DATA")
	}

	close(wsIn)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serveLocalWebSocket did not exit after wsIn closed")
	}
}

func TestServeLocalWebSocket_RejectsNonLoopback(t *testing.T) {
	fc := newFakeFrameSink()
	wsIn := make(chan tunnelframe.Frame)
	serveLocalWebSocket(context.Background(), fc, 1, tunnelframe.ReqStartPayload{
		Method: "GET", Path: "/", Target: "10.0.0.5:80", IsWebSocket: true,
	}, wsIn, newCreditWindow(1<<20))
	if fc.error() == "" {
		t.Fatal("expected RESP_ERROR for non-loopback target")
	}
}

func TestServeLocalWebSocket_UnreachableEmitsError(t *testing.T) {
	fc := newFakeFrameSink()
	wsIn := make(chan tunnelframe.Frame)
	serveLocalWebSocket(context.Background(), fc, 1, tunnelframe.ReqStartPayload{
		Method: "GET", Path: "/", Target: "127.0.0.1:1", IsWebSocket: true,
	}, wsIn, newCreditWindow(1<<20))
	if fc.error() == "" {
		t.Fatal("expected RESP_ERROR when upstream dial fails")
	}
}
