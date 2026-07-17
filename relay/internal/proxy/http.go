package proxy

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	otelmetrics "github.com/anthropics/agentsmesh/relay/internal/otel"
	"github.com/anthropics/agentsmesh/relay/internal/protocol/tunnelframe"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

// proxyByteCounters accumulates bytes moved by a single proxied
// request/stream for the gateway.preview.bytes OTel counter. Request-body
// upload bytes (in) are tracked best-effort: sendRequestBody runs
// concurrently and ProxyHTTP doesn't wait for it, so a slow/large upload
// still in flight when the response completes is undercounted rather than
// blocking the response on the upload finishing.
type proxyByteCounters struct {
	in  atomic.Int64
	out atomic.Int64
}

// tunnelIface is the minimal tunnel surface ProxyHTTP needs. *tunnel.Tunnel
// satisfies it; tests inject a fake.
type tunnelIface interface {
	OpenStream() *tunnel.Stream
	WriteFrame(tunnelframe.Frame) error
	CloseStream(id uint32)
}

// ProxyParams configures a single proxied request.
type ProxyParams struct {
	PodKey           string
	Target           string // loopback target the runner will dial, e.g. 127.0.0.1:3000
	Path             string // path forwarded to the target (already stripped of /preview/{podKey})
	RawQuery         string
	HiddenCookieName string
	ExpectedOrigin   string
	Reauthorize      func(context.Context) error
	ReauthorizeEvery time.Duration
	WindowBytes      int
	Timeout          time.Duration
}

// ProxyHTTP forwards a browser HTTP request over a tunnel stream and streams the
// upstream response back to w. The whole exchange is credit-controlled and
// bounded to a single stream: a timeout cancels only this stream (via
// STREAM_CANCEL), never the shared tunnel.
func ProxyHTTP(ctx context.Context, tun tunnelIface, w http.ResponseWriter, r *http.Request, p ProxyParams) error {
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}

	st := tun.OpenStream()
	defer tun.CloseStream(st.ID)

	// Capture forwarding metadata before sanitizing (which strips inbound XFF).
	clientIP := clientIP(r)
	proto := forwardedProto(r)
	header := SanitizeRequestHeaders(r.Header, clientIP, proto, r.Host, p.HiddenCookieName)

	reqStart := tunnelframe.ReqStartPayload{
		Method:        r.Method,
		Path:          p.Path,
		RawQuery:      p.RawQuery,
		Header:        header,
		PodKey:        p.PodKey,
		Target:        p.Target,
		ContentLength: r.ContentLength,
		IsWebSocket:   false,
	}
	if err := tun.WriteFrame(tunnelframe.Frame{Type: tunnelframe.TypeReqStart, StreamID: st.ID, Payload: tunnelframe.EncodeJSON(reqStart)}); err != nil {
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return err
	}

	// Stream the request body concurrently so CREDIT replenishment (handled in
	// the response loop below) can unblock large uploads.
	counters := &proxyByteCounters{}
	go sendRequestBody(ctx, tun, st, r.Body, counters)

	status, err := pumpResponse(ctx, tun, st, w, counters, p.HiddenCookieName)
	otelmetrics.RecordPreviewRequest(ctx, strconv.Itoa(status))
	otelmetrics.RecordPreviewBytes(ctx, "in", counters.in.Load())
	otelmetrics.RecordPreviewBytes(ctx, "out", counters.out.Load())
	return err
}

// sendRequestBody streams r.Body as REQ_BODY frames (credit-gated) then REQ_END.
func sendRequestBody(ctx context.Context, tun tunnelIface, st *tunnel.Stream, body io.ReadCloser, counters *proxyByteCounters) {
	if body != nil {
		buf := make([]byte, tunnelframe.MaxChunk)
		for {
			n, rerr := body.Read(buf)
			if n > 0 {
				if err := st.AcquireSend(ctx, n); err != nil {
					_ = tun.WriteFrame(cancelFrame(st.ID, "request cancelled"))
					return
				}
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				if err := tun.WriteFrame(tunnelframe.Frame{Type: tunnelframe.TypeReqBody, StreamID: st.ID, Payload: chunk}); err != nil {
					return
				}
				counters.in.Add(int64(n))
			}
			if rerr == io.EOF {
				break
			}
			if rerr != nil {
				_ = tun.WriteFrame(cancelFrame(st.ID, "request read error"))
				return
			}
		}
	}
	_ = tun.WriteFrame(tunnelframe.Frame{Type: tunnelframe.TypeReqEnd, StreamID: st.ID})
}

func copyHeader(dst, src http.Header) {
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func forwardedProto(r *http.Request) string {
	if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
		return p
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
