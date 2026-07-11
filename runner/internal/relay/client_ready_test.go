package relay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestConnectFailsClosedWithoutPublisherReady(t *testing.T) {
	tests := []struct {
		name string
		send func(*websocket.Conn)
	}{
		{name: "connection closes", send: func(conn *websocket.Conn) { _ = conn.Close() }},
		{
			name: "wrong message type",
			send: func(conn *websocket.Conn) {
				_ = conn.WriteMessage(websocket.BinaryMessage, EncodeMessage(MsgTypeOutput, []byte(`{"type":"publisher_ready"}`)))
			},
		},
		{
			name: "wrong control payload",
			send: func(conn *websocket.Conn) {
				_ = conn.WriteMessage(websocket.BinaryMessage, EncodeMessage(MsgTypeControl, []byte(`{"type":"other"}`)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := testUpgrader.Upgrade(w, r, nil)
				if err == nil {
					tt.send(conn)
				}
			}))
			defer srv.Close()

			c := NewClient(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http"), "pod-1", "test-token", nil)
			if err := c.Connect(); err == nil {
				t.Fatal("Connect succeeded without publisher_ready")
			}
			if c.IsConnected() {
				t.Fatal("client marked connected without publisher_ready")
			}
		})
	}
}

func TestConnectTimesOutWaitingForPublisherReady(t *testing.T) {
	accepted := make(chan struct{})
	srv := silentPublisherServer(t, accepted)
	defer srv.Close()

	c := NewClient(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http"), "pod-1", "test-token", nil)
	c.readyTimeout = 30 * time.Millisecond
	err := c.Connect()
	if err == nil || !strings.Contains(err.Error(), "publisher ready") {
		t.Fatalf("Connect error = %v, want publisher ready timeout", err)
	}
	if c.IsConnected() {
		t.Fatal("client marked connected before publisher_ready")
	}
}

func TestStopInterruptsPublisherReadyWait(t *testing.T) {
	accepted := make(chan struct{})
	srv := silentPublisherServer(t, accepted)
	defer srv.Close()

	c := NewClient(context.Background(), "ws"+strings.TrimPrefix(srv.URL, "http"), "pod-1", "test-token", nil)
	connectDone := make(chan error, 1)
	go func() { connectDone <- c.Connect() }()
	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("server was not reached")
	}

	started := time.Now()
	c.Stop()
	select {
	case err := <-connectDone:
		if err == nil {
			t.Fatal("Connect succeeded after Stop")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Stop did not interrupt publisher ready wait")
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("Stop took %v", elapsed)
	}
}

func silentPublisherServer(t *testing.T, accepted chan<- struct{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		close(accepted)
		defer conn.Close()
		<-r.Context().Done()
	}))
}
