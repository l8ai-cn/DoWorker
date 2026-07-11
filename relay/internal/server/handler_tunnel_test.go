package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
	"github.com/anthropics/agentsmesh/relay/internal/protocol/tunnelframe"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

func newTestTunnelHandler(t *testing.T) *TunnelHandler {
	t.Helper()
	tv := auth.NewTokenValidator(testSecret, testIssuer)
	return NewTunnelHandler(tv, tunnel.NewRegistry(), auth.NewOriginChecker(nil), 1<<20)
}

func validTunnelToken(t *testing.T, runnerID int64) string {
	t.Helper()
	tok, err := auth.GenerateTypedToken(testSecret, testIssuer, auth.TokenTypeTunnel, "", runnerID, 0, 3, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

func TestHandleTunnelWS_RejectsNonTunnelToken(t *testing.T) {
	h := newTestTunnelHandler(t)
	srv := httptest.NewServer(http.HandlerFunc(h.HandleTunnelWS))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/?token=" + validToken("pod-1")
	_, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err == nil {
		t.Fatal("expected reject")
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestHandleTunnelWS_RegistersTunnel(t *testing.T) {
	h := newTestTunnelHandler(t)
	srv := httptest.NewServer(http.HandlerFunc(h.HandleTunnelWS))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/?token=" + validTunnelToken(t, 7)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	hello, _ := json.Marshal(tunnelframe.HelloPayload{RunnerID: "7"})
	_ = conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(tunnelframe.Frame{Type: tunnelframe.TypeHello, Payload: hello}))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read HELLO_ACK: %v", err)
	}
	ack, err := tunnelframe.Decode(data)
	if err != nil {
		t.Fatalf("decode HELLO_ACK: %v", err)
	}
	if ack.Type != tunnelframe.TypeHelloAck || ack.StreamID != 0 || len(ack.Payload) != 0 {
		t.Fatalf("unexpected HELLO_ACK: %+v", ack)
	}
	waitFor(t, func() bool { return h.registry.Get(7) != nil })
}
