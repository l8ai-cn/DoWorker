package tunnel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

func TestClient_ConnectAndHello(t *testing.T) {
	got := make(chan tunnelframe.HelloPayload, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_, data, err := c.ReadMessage()
		if err != nil {
			return
		}
		f, _ := tunnelframe.Decode(data)
		var hp tunnelframe.HelloPayload
		_ = json.Unmarshal(f.Payload, &hp)
		got <- hp
		_ = c.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(tunnelframe.Frame{Type: tunnelframe.TypeHelloAck}))
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	cl := NewClient(context.Background(), url, "tok", 7, 3, nil)
	if err := cl.Connect(); err != nil {
		t.Fatal(err)
	}
	defer cl.Stop()
	select {
	case hp := <-got:
		if hp.RunnerID != "7" {
			t.Fatalf("bad runner id %q", hp.RunnerID)
		}
	case <-time.After(time.Second):
		t.Fatal("no hello received")
	}
}

func TestClient_ConnectFailsClosedWithoutHelloAck(t *testing.T) {
	tests := []struct {
		name string
		send func(*websocket.Conn)
	}{
		{
			name: "connection closes",
			send: func(conn *websocket.Conn) {
				_ = conn.Close()
			},
		},
		{
			name: "wrong frame",
			send: func(conn *websocket.Conn) {
				_ = conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(tunnelframe.Frame{Type: tunnelframe.TypePong}))
			},
		},
		{
			name: "wrong stream",
			send: func(conn *websocket.Conn) {
				_ = conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(tunnelframe.Frame{Type: tunnelframe.TypeHelloAck, StreamID: 1}))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := upgradeTunnel(w, r)
				if err != nil {
					return
				}
				defer conn.Close()
				readHelloOnly(t, conn)
				tt.send(conn)
			}))
			defer srv.Close()

			cl := newTestClient(srv.URL)
			if err := cl.Connect(); err == nil {
				t.Fatal("Connect succeeded without valid HELLO_ACK")
			}
			if cl.IsConnected() {
				t.Fatal("client marked connected without HELLO_ACK")
			}
		})
	}
}

func TestClient_ConnectTimesOutWaitingForHelloAck(t *testing.T) {
	helloRead := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeTunnel(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		readHelloOnly(t, conn)
		close(helloRead)
		<-r.Context().Done()
	}))
	defer srv.Close()

	cl := newTestClient(srv.URL)
	cl.readyTimeout = 30 * time.Millisecond
	err := cl.Connect()
	if err == nil || !strings.Contains(err.Error(), "HELLO_ACK") {
		t.Fatalf("Connect error = %v, want HELLO_ACK timeout", err)
	}
	if cl.IsConnected() {
		t.Fatal("client marked connected before HELLO_ACK")
	}
	select {
	case <-helloRead:
	default:
		t.Fatal("server did not read HELLO")
	}
}

func TestClient_StopInterruptsHelloAckWait(t *testing.T) {
	helloRead := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgradeTunnel(w, r)
		if err != nil {
			return
		}
		defer conn.Close()
		readHelloOnly(t, conn)
		close(helloRead)
		<-r.Context().Done()
	}))
	defer srv.Close()

	cl := newTestClient(srv.URL)
	connectDone := make(chan error, 1)
	go func() {
		connectDone <- cl.Connect()
	}()
	select {
	case <-helloRead:
	case <-time.After(time.Second):
		t.Fatal("server did not read HELLO")
	}
	started := time.Now()
	cl.Stop()
	select {
	case err := <-connectDone:
		if err == nil {
			t.Fatal("Connect succeeded after Stop")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Stop did not interrupt HELLO_ACK wait")
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("Stop took %v", elapsed)
	}
}
