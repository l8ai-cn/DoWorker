package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
	"github.com/anthropics/agentsmesh/relay/internal/protocol/tunnelframe"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

// fakeRunnerTunnel is a minimal in-memory stand-in for the runner-side tunnel
// client + dispatcher (runner/internal/tunnel), reimplemented here against
// tunnelframe directly so relay's tests don't cross the relay/runner module
// boundary. It proxies REQ_START to a real local HTTP server, exactly like
// runner/internal/tunnel/local_http.go does in production.
type fakeRunnerTunnel struct {
	conn   *websocket.Conn
	target string // e.g. 127.0.0.1:port of the fake local service

	mu        sync.Mutex
	closed    bool
	wsStreams map[uint32]chan tunnelframe.Frame
}

func newFakeRunnerTunnel(t *testing.T, gatewayWSURL, tunnelToken string, target string) *fakeRunnerTunnel {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(gatewayWSURL, nil)
	if err != nil {
		t.Fatalf("fake runner dial failed: %v", err)
	}
	ft := &fakeRunnerTunnel{conn: conn, target: target}
	hello := tunnelframe.HelloPayload{}
	if err := ft.writeFrame(tunnelframe.Frame{Type: tunnelframe.TypeHello, Payload: tunnelframe.EncodeJSON(hello)}); err != nil {
		t.Fatalf("fake runner hello failed: %v", err)
	}
	go ft.readLoop(t)
	return ft
}

func (ft *fakeRunnerTunnel) writeFrame(f tunnelframe.Frame) error {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	if ft.closed {
		return nil
	}
	return ft.conn.WriteMessage(websocket.BinaryMessage, tunnelframe.Encode(f))
}

func (ft *fakeRunnerTunnel) readLoop(t *testing.T) {
	for {
		_, data, err := ft.conn.ReadMessage()
		if err != nil {
			return
		}
		f, derr := tunnelframe.Decode(data)
		if derr != nil {
			continue
		}
		switch f.Type {
		case tunnelframe.TypePing:
			_ = ft.writeFrame(tunnelframe.Frame{Type: tunnelframe.TypePong})
		case tunnelframe.TypeReqStart:
			var p tunnelframe.ReqStartPayload
			if json.Unmarshal(f.Payload, &p) == nil && p.IsWebSocket {
				go ft.serveWebSocket(f.StreamID, p)
				continue
			}
			go ft.serveReqStart(f)
		case tunnelframe.TypeWSData, tunnelframe.TypeWSClose:
			ft.routeToWS(f)
		}
	}
}

// wsRoutes maps an in-flight WS stream id to the channel serveWebSocket reads
// inbound WS_DATA/WS_CLOSE frames from (mirrors runner/internal/tunnel's
// dispatcher wsIn wiring at test scale).
func (ft *fakeRunnerTunnel) routeToWS(f tunnelframe.Frame) {
	ft.mu.Lock()
	ch := ft.wsStreams[f.StreamID]
	ft.mu.Unlock()
	if ch != nil {
		select {
		case ch <- f:
		default:
		}
	}
}

// serveWebSocket mirrors runner/internal/tunnel/local_http.go's
// serveLocalWebSocket: dial the local upstream WS and relay frames in both
// directions over WS_DATA.
func (ft *fakeRunnerTunnel) serveWebSocket(streamID uint32, p tunnelframe.ReqStartPayload) {
	ch := make(chan tunnelframe.Frame, 32)
	ft.mu.Lock()
	if ft.wsStreams == nil {
		ft.wsStreams = make(map[uint32]chan tunnelframe.Frame)
	}
	ft.wsStreams[streamID] = ch
	ft.mu.Unlock()
	defer func() {
		ft.mu.Lock()
		delete(ft.wsStreams, streamID)
		ft.mu.Unlock()
	}()

	dialURL := "ws://" + p.Target + p.Path
	upConn, _, err := websocket.DefaultDialer.Dial(dialURL, nil)
	if err != nil {
		_ = ft.writeFrame(tunnelframe.Frame{Type: tunnelframe.TypeRespError, StreamID: streamID,
			Payload: tunnelframe.EncodeJSON(tunnelframe.RespErrorPayload{Code: "target_unreachable", Message: err.Error()})})
		return
	}
	defer upConn.Close()

	_ = ft.writeFrame(tunnelframe.Frame{Type: tunnelframe.TypeRespStart, StreamID: streamID,
		Payload: tunnelframe.EncodeJSON(tunnelframe.RespStartPayload{Status: http.StatusSwitchingProtocols})})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			mt, data, err := upConn.ReadMessage()
			if err != nil {
				return
			}
			if err := ft.writeFrame(tunnelframe.Frame{Type: tunnelframe.TypeWSData, StreamID: streamID,
				Payload: tunnelframe.EncodeJSON(tunnelframe.WSDataPayload{MessageType: mt, Data: data})}); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case f := <-ch:
			switch f.Type {
			case tunnelframe.TypeWSData:
				var wd tunnelframe.WSDataPayload
				if json.Unmarshal(f.Payload, &wd) == nil {
					_ = upConn.WriteMessage(wd.MessageType, wd.Data)
				}
			case tunnelframe.TypeWSClose:
				return
			}
		}
	}
}

// serveReqStart mirrors runner/internal/tunnel/local_http.go's serveLocalHTTP
// at test scale: fetch from the fake local service and stream RESP_* back.
func (ft *fakeRunnerTunnel) serveReqStart(f tunnelframe.Frame) {
	var p tunnelframe.ReqStartPayload
	if err := json.Unmarshal(f.Payload, &p); err != nil {
		return
	}
	rawURL := "http://" + p.Target + p.Path
	if p.RawQuery != "" {
		rawURL += "?" + p.RawQuery
	}
	req, err := http.NewRequest(p.Method, rawURL, nil)
	if err != nil {
		_ = ft.writeFrame(tunnelframe.Frame{Type: tunnelframe.TypeRespError, StreamID: f.StreamID,
			Payload: tunnelframe.EncodeJSON(tunnelframe.RespErrorPayload{Code: "bad_request", Message: err.Error()})})
		return
	}
	if rangeHdr := p.Header.Get("Range"); rangeHdr != "" {
		req.Header.Set("Range", rangeHdr)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		_ = ft.writeFrame(tunnelframe.Frame{Type: tunnelframe.TypeRespError, StreamID: f.StreamID,
			Payload: tunnelframe.EncodeJSON(tunnelframe.RespErrorPayload{Code: "target_unreachable", Message: err.Error()})})
		return
	}
	defer resp.Body.Close()

	_ = ft.writeFrame(tunnelframe.Frame{
		Type:     tunnelframe.TypeRespStart,
		StreamID: f.StreamID,
		Payload: tunnelframe.EncodeJSON(tunnelframe.RespStartPayload{
			Status: resp.StatusCode,
			Header: resp.Header,
		}),
	})

	buf := make([]byte, tunnelframe.MaxChunk)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			if err := ft.writeFrame(tunnelframe.Frame{Type: tunnelframe.TypeRespBody, StreamID: f.StreamID, Payload: chunk}); err != nil {
				return
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return
		}
	}
	_ = ft.writeFrame(tunnelframe.Frame{Type: tunnelframe.TypeRespEnd, StreamID: f.StreamID})
}

func (ft *fakeRunnerTunnel) Close() {
	ft.mu.Lock()
	ft.closed = true
	ft.mu.Unlock()
	_ = ft.conn.Close()
}

// newPreviewE2EGateway wires a bare Gateway HTTP mux (tunnel + preview
// endpoints only, no backend registration) for end-to-end tests.
func newPreviewE2EGateway(t *testing.T) (gw *httptest.Server, registry *tunnel.Registry, validator *auth.TokenValidator) {
	t.Helper()
	validator = auth.NewTokenValidator("s3cret", "iss")
	registry = tunnel.NewRegistry()
	tunnelHandler := NewTunnelHandler(validator, registry, auth.NewOriginChecker(nil), 1<<20)
	limiter := tunnel.NewPodLimiter(32, 16, 5*time.Second)
	previewHandler := NewPreviewHandler(validator, registry, limiter, PreviewConfig{
		ReconnectGrace:    2 * time.Second,
		StreamTimeout:     10 * time.Second,
		StreamWindowBytes: 1 << 20,
	})
	mux := http.NewServeMux()
	mux.HandleFunc("/runner/tunnel", tunnelHandler.HandleTunnelWS)
	mux.HandleFunc("/preview/", previewHandler.route)
	gw = httptest.NewServer(mux)
	return gw, registry, validator
}

func waitForRunnerRegistered(t *testing.T, registry *tunnel.Registry, runnerID int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if registry.Get(runnerID) != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("runner %d never registered", runnerID)
}

func TestPreviewE2E_HTMLAndImage(t *testing.T) {
	const runnerID = int64(7)

	imageBytes := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 1, 2, 3, 4, 5}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/app":
			w.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(w, "<h1>ok</h1>")
		case "/app/logo.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(imageBytes)
		case "/app/%2e%2e":
			if got := r.URL.EscapedPath(); got != "/app/%252e%252e" {
				t.Errorf("escaped upstream path = %q, want /app/%%252e%%252e", got)
				http.Error(w, "wrong escaped path", http.StatusBadRequest)
				return
			}
			_, _ = io.WriteString(w, "literal-double-encoding")
		default:
			http.NotFound(w, r)
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

	previewToken := mustPreviewTokenWithPath(t, "pod1", runnerID, target, "/app")

	htmlResp, err := http.Get(gw.URL + "/preview/pod1/?token=" + previewToken)
	if err != nil {
		t.Fatal(err)
	}
	defer htmlResp.Body.Close()
	htmlBody, _ := io.ReadAll(htmlResp.Body)
	if htmlResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", htmlResp.StatusCode, htmlBody)
	}
	if !strings.Contains(string(htmlBody), "<h1>ok</h1>") {
		t.Fatalf("unexpected html body: %s", htmlBody)
	}
	if ct := htmlResp.Header.Get("Content-Type"); ct != "text/html" {
		t.Fatalf("unexpected content-type: %s", ct)
	}

	imgResp, err := http.Get(gw.URL + "/preview/pod1/logo.png?token=" + previewToken)
	if err != nil {
		t.Fatal(err)
	}
	defer imgResp.Body.Close()
	imgBody, _ := io.ReadAll(imgResp.Body)
	if imgResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", imgResp.StatusCode)
	}
	if imgResp.Header.Get("Content-Type") != "image/png" {
		t.Fatalf("unexpected content-type: %s", imgResp.Header.Get("Content-Type"))
	}
	if string(imgBody) != string(imageBytes) {
		t.Fatalf("image body mismatch: got %d bytes, want %d bytes", len(imgBody), len(imageBytes))
	}

	encodedResp, err := http.Get(gw.URL + "/preview/pod1/%252e%252e?token=" + previewToken)
	if err != nil {
		t.Fatal(err)
	}
	defer encodedResp.Body.Close()
	encodedBody, _ := io.ReadAll(encodedResp.Body)
	if encodedResp.StatusCode != http.StatusOK || string(encodedBody) != "literal-double-encoding" {
		t.Fatalf("double-encoded path: status=%d body=%q", encodedResp.StatusCode, encodedBody)
	}
}
