package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/anthropics/agentsmesh/relay/internal/protocol/tunnelframe"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

// fakeTunnel is an in-memory tunnel that lets tests observe outbound frames and
// inject peer frames onto a stream's response channel.
type fakeTunnel struct {
	mu         sync.Mutex
	streams    map[uint32]*tunnel.Stream
	next       uint32
	window     int
	onReqStart func(st *tunnel.Stream, p tunnelframe.ReqStartPayload)
}

func newFakeTunnel() *fakeTunnel {
	return &fakeTunnel{streams: map[uint32]*tunnel.Stream{}, window: 1 << 20}
}

func (f *fakeTunnel) OpenStream() *tunnel.Stream {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.next++
	st := tunnel.NewStream(f.next, f.window)
	f.streams[st.ID] = st
	return st
}

func (f *fakeTunnel) CloseStream(id uint32) {
	f.mu.Lock()
	st := f.streams[id]
	delete(f.streams, id)
	f.mu.Unlock()
	_ = st
}

func (f *fakeTunnel) WriteFrame(fr tunnelframe.Frame) error {
	if fr.Type == tunnelframe.TypeReqStart && f.onReqStart != nil {
		var p tunnelframe.ReqStartPayload
		_ = json.Unmarshal(fr.Payload, &p)
		f.mu.Lock()
		st := f.streams[fr.StreamID]
		f.mu.Unlock()
		if st != nil {
			go f.onReqStart(st, p)
		}
	}
	return nil
}

func (f *fakeTunnel) inject(id uint32, typ tunnelframe.FrameType, payload []byte) {
	f.mu.Lock()
	st := f.streams[id]
	f.mu.Unlock()
	if st != nil {
		st.Deliver(tunnelframe.Frame{Type: typ, StreamID: id, Payload: payload})
	}
}

func mustJSON(v interface{}) []byte { return tunnelframe.EncodeJSON(v) }

func TestProxyHTTP_StreamsResponse(t *testing.T) {
	ft := newFakeTunnel()
	ft.onReqStart = func(st *tunnel.Stream, p tunnelframe.ReqStartPayload) {
		ft.inject(st.ID, tunnelframe.TypeRespStart, mustJSON(tunnelframe.RespStartPayload{Status: 200, Header: http.Header{"Content-Type": {"text/plain"}}}))
		ft.inject(st.ID, tunnelframe.TypeRespBody, []byte("hi"))
		ft.inject(st.ID, tunnelframe.TypeRespEnd, nil)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/preview/pod1/index.html", nil)

	err := ProxyHTTP(context.Background(), ft, rec, req, ProxyParams{
		PodKey: "pod1", Target: "127.0.0.1:3000", Path: "/index.html", WindowBytes: 1 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || rec.Body.String() != "hi" {
		t.Fatalf("bad response: %d %q", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "text/plain" {
		t.Fatalf("content-type not propagated: %q", rec.Header().Get("Content-Type"))
	}
}

func TestProxyHTTP_UpstreamError(t *testing.T) {
	ft := newFakeTunnel()
	ft.onReqStart = func(st *tunnel.Stream, p tunnelframe.ReqStartPayload) {
		ft.inject(st.ID, tunnelframe.TypeRespError, mustJSON(tunnelframe.RespErrorPayload{Code: "target_unreachable"}))
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/preview/pod1/x", nil)
	err := ProxyHTTP(context.Background(), ft, rec, req, ProxyParams{PodKey: "pod1", Target: "127.0.0.1:3000", Path: "/x", WindowBytes: 1 << 20})
	if err == nil {
		t.Fatal("expected error")
	}
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}
