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

	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

// fakeFrameSink collects RESP_* frames emitted by serveLocalHTTP.
type fakeFrameSink struct {
	mu         sync.Mutex
	statusCode int
	buf        bytes.Buffer
	errCode    string
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
	}
	return nil
}

func (f *fakeFrameSink) status() int   { f.mu.Lock(); defer f.mu.Unlock(); return f.statusCode }
func (f *fakeFrameSink) body() string  { f.mu.Lock(); defer f.mu.Unlock(); return f.buf.String() }
func (f *fakeFrameSink) error() string { f.mu.Lock(); defer f.mu.Unlock(); return f.errCode }

func newFakeFrameSink() *fakeFrameSink { return &fakeFrameSink{} }

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
