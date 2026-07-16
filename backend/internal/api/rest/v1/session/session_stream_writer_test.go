package sessionapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSessionStreamWritesOutliveServerWriteTimeout(t *testing.T) {
	const (
		frameCount   = 5
		frameGap     = 30 * time.Millisecond
		writeTimeout = 60 * time.Millisecond
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := prepareSessionStreamWriter(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		for range frameCount {
			if err := writeSessionStreamFrame(w, ": keepalive\n\n"); err != nil {
				return
			}
			time.Sleep(frameGap)
		}
	})

	server := httptest.NewUnstartedServer(handler)
	server.Config.WriteTimeout = writeTimeout
	server.Start()
	defer server.Close()

	response, err := http.Get(server.URL) //nolint:noctx
	if err != nil {
		t.Fatalf("GET stream: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read stream: %v", err)
	}
	if got := strings.Count(string(body), "keepalive"); got != frameCount {
		t.Fatalf("received %d frames, want %d", got, frameCount)
	}
}
