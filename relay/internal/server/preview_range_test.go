package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/relay/internal/auth"
)

// TestPreviewE2E_VideoRange verifies Range requests (and their 206 responses,
// including Content-Range/Accept-Ranges) pass through the tunnel untouched,
// exactly like any other media response: the runner's local HTTP client
// forwards Range as-is, and the gateway proxy never rewrites status/headers.
func TestPreviewE2E_VideoRange(t *testing.T) {
	const runnerID = int64(9)
	video := bytes.Repeat([]byte{0xAB}, 1<<20) // 1MB fake video payload

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		http.ServeContent(w, r, "movie.mp4", time.Time{}, bytes.NewReader(video))
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

	req, err := http.NewRequest("GET", gw.URL+"/preview/pod1/movie.mp4", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "pod1.preview.example.com"
	req.AddCookie(&http.Cookie{Name: previewCookieName, Value: previewToken})
	req.Header.Set("Range", "bytes=0-1023")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusPartialContent {
		t.Fatalf("expected 206, got %d body=%s", resp.StatusCode, body)
	}
	wantContentRange := "bytes 0-1023/1048576"
	if got := resp.Header.Get("Content-Range"); got != wantContentRange {
		t.Fatalf("Content-Range = %q, want %q", got, wantContentRange)
	}
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		t.Fatalf("expected Accept-Ranges: bytes, got %q", resp.Header.Get("Accept-Ranges"))
	}
	if len(body) != 1024 {
		t.Fatalf("expected 1024 bytes, got %d", len(body))
	}
	if !bytes.Equal(body, video[:1024]) {
		t.Fatalf("range body content mismatch")
	}
}
