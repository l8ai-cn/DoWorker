package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
)

// TestPreviewE2E_WebSocket verifies WebSocket upgrades pass through the tunnel:
// the browser connects to /preview/{podKey}/ws/echo, the gateway holds off
// upgrading until the fake runner's upstream WS dial succeeds (RESP_START
// status=101), and messages round-trip in both directions.
func TestPreviewE2E_WebSocket(t *testing.T) {
	const runnerID = int64(21)

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

	gw, registry, _ := newPreviewE2EGateway(t)
	defer gw.Close()

	tunnelToken, err := auth.GenerateTypedToken("s3cret", "iss", auth.TokenTypeTunnel, "", runnerID, 0, 3, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	wsURL := "ws" + strings.TrimPrefix(gw.URL, "http") + "/runner/tunnel?token=" + tunnelToken
	fr := newFakeRunnerTunnel(t, wsURL, tunnelToken, target)
	defer fr.Close()
	waitForRunnerRegistered(t, registry, runnerID)

	previewToken := mustPreviewToken(t, "pod1", runnerID, target)

	previewWSURL := "ws" + strings.TrimPrefix(gw.URL, "http") + "/preview/pod1/echo?token=" + previewToken
	conn, resp, err := websocket.DefaultDialer.Dial(previewWSURL, nil)
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("dial failed (status=%d): %v", status, err)
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
		t.Fatal(err)
	}
	mt, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if mt != websocket.TextMessage || string(data) != "ping" {
		t.Fatalf("unexpected echo: type=%d data=%q", mt, data)
	}
}

// TestPreviewE2E_WebSocketUpstreamOfflineReturnsHTTPError verifies that when
// the runner-side dial fails, the browser gets a normal HTTP error response
// instead of a half-open WebSocket, because the gateway waits for RESP_START
// before upgrading the client connection.
func TestPreviewE2E_WebSocketUpstreamOfflineReturnsHTTPError(t *testing.T) {
	const runnerID = int64(22)

	// No upstream listening on this address at all.
	target := "127.0.0.1:1"

	gw, registry, _ := newPreviewE2EGateway(t)
	defer gw.Close()

	tunnelToken, err := auth.GenerateTypedToken("s3cret", "iss", auth.TokenTypeTunnel, "", runnerID, 0, 3, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	wsURL := "ws" + strings.TrimPrefix(gw.URL, "http") + "/runner/tunnel?token=" + tunnelToken
	fr := newFakeRunnerTunnel(t, wsURL, tunnelToken, target)
	defer fr.Close()
	waitForRunnerRegistered(t, registry, runnerID)

	previewToken := mustPreviewToken(t, "pod1", runnerID, target)
	previewWSURL := "ws" + strings.TrimPrefix(gw.URL, "http") + "/preview/pod1/echo?token=" + previewToken
	_, resp, err := websocket.DefaultDialer.Dial(previewWSURL, nil)
	if err == nil {
		t.Fatal("expected dial to fail when upstream is offline")
	}
	if resp == nil || resp.StatusCode != http.StatusBadGateway {
		got := 0
		if resp != nil {
			got = resp.StatusCode
		}
		t.Fatalf("expected 502, got %d", got)
	}
}
