package server

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/l8ai-cn/agentcloud/relay/internal/auth"
)

// TestPreviewE2E_SSE verifies event-stream responses are flushed frame-by-frame
// through the tunnel rather than buffered until the upstream closes: each
// RESP_BODY is written+flushed to the browser connection immediately
// (relay/internal/proxy/http.go's pumpResponse), and the runner's local HTTP
// client forwards each upstream Read()/Flush() as its own frame.
func TestPreviewE2E_SSE(t *testing.T) {
	const runnerID = int64(13)
	const eventDelay = 40 * time.Millisecond

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		flusher := w.(http.Flusher)
		for i := 0; i < 3; i++ {
			fmt.Fprintf(w, "data: %d\n\n", i)
			flusher.Flush()
			if i < 2 {
				time.Sleep(eventDelay)
			}
		}
	}))
	defer upstream.Close()
	target := strings.TrimPrefix(upstream.URL, "http://")

	gw, registry, _, _ := newPreviewE2EGateway(t)
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

	resp := doPreviewGET(t, gw.URL+"/preview/pod1/events", "pod1", previewToken)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("unexpected content-type: %s", ct)
	}

	reader := bufio.NewReader(resp.Body)
	var received []time.Time
	for i := 0; i < 3; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read event %d failed: %v", i, err)
		}
		want := fmt.Sprintf("data: %d\n", i)
		if line != want {
			t.Fatalf("event %d = %q, want %q", i, line, want)
		}
		received = append(received, time.Now())
		if _, err := reader.ReadString('\n'); err != nil { // consume the blank line terminator
			t.Fatalf("read event %d terminator failed: %v", i, err)
		}
	}

	// If the pipeline buffered until the upstream handler finished, all three
	// events would arrive back-to-back instead of spread across ~2*eventDelay.
	spread := received[2].Sub(received[0])
	if spread < eventDelay {
		t.Fatalf("events arrived without streaming (spread=%v, want >= %v): buffering suspected", spread, eventDelay)
	}
}
