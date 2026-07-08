package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/anthropics/agentsmesh/relay/internal/protocol/tunnelframe"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

// tunnelIface is the minimal tunnel surface ProxyHTTP needs. *tunnel.Tunnel
// satisfies it; tests inject a fake.
type tunnelIface interface {
	OpenStream() *tunnel.Stream
	WriteFrame(tunnelframe.Frame) error
	CloseStream(id uint32)
}

// ProxyParams configures a single proxied request.
type ProxyParams struct {
	PodKey      string
	Target      string // loopback target the runner will dial, e.g. 127.0.0.1:3000
	Path        string // path forwarded to the target (already stripped of /preview/{podKey})
	WindowBytes int
	Timeout     time.Duration
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
	header := SanitizeRequestHeaders(r.Header, clientIP, proto, r.Host)

	reqStart := tunnelframe.ReqStartPayload{
		Method:        r.Method,
		Path:          p.Path,
		RawQuery:      r.URL.RawQuery,
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
	go sendRequestBody(ctx, tun, st, r.Body)

	return pumpResponse(ctx, tun, st, w)
}

// sendRequestBody streams r.Body as REQ_BODY frames (credit-gated) then REQ_END.
func sendRequestBody(ctx context.Context, tun tunnelIface, st *tunnel.Stream, body io.ReadCloser) {
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

// pumpResponse reads control/body frames off the stream and writes the HTTP
// response, flushing per chunk and granting CREDIT back to the peer.
func pumpResponse(ctx context.Context, tun tunnelIface, st *tunnel.Stream, w http.ResponseWriter) error {
	flusher, _ := w.(http.Flusher)
	started := false
	for {
		select {
		case <-ctx.Done():
			_ = tun.WriteFrame(cancelFrame(st.ID, "gateway timeout"))
			if !started {
				http.Error(w, "gateway timeout", http.StatusGatewayTimeout)
			}
			return ctx.Err()
		case f := <-st.RespChan():
			switch f.Type {
			case tunnelframe.TypeRespStart:
				var rs tunnelframe.RespStartPayload
				if err := json.Unmarshal(f.Payload, &rs); err != nil {
					http.Error(w, "bad gateway", http.StatusBadGateway)
					return err
				}
				copyHeader(w.Header(), SanitizeResponseHeaders(rs.Header))
				status := rs.Status
				if status == 0 {
					status = http.StatusOK
				}
				w.WriteHeader(status)
				started = true
				if flusher != nil {
					flusher.Flush()
				}
			case tunnelframe.TypeRespBody:
				if !started {
					return fmt.Errorf("proxy: RESP_BODY before RESP_START")
				}
				if len(f.Payload) > 0 {
					if _, err := w.Write(f.Payload); err != nil {
						_ = tun.WriteFrame(cancelFrame(st.ID, "client write error"))
						return err
					}
					if flusher != nil {
						flusher.Flush()
					}
					// Replenish the peer's send window for what we've flushed.
					_ = tun.WriteFrame(creditFrame(st.ID, len(f.Payload)))
				}
			case tunnelframe.TypeRespEnd:
				return nil
			case tunnelframe.TypeRespError:
				var re tunnelframe.RespErrorPayload
				_ = json.Unmarshal(f.Payload, &re)
				if !started {
					http.Error(w, re.Code, statusForCode(re.Code))
				}
				return fmt.Errorf("proxy: upstream error %q", re.Code)
			case tunnelframe.TypeCredit:
				var c tunnelframe.CreditPayload
				if json.Unmarshal(f.Payload, &c) == nil && c.Bytes > 0 {
					st.AddSendCredit(c.Bytes)
				}
			}
		}
	}
}

func cancelFrame(id uint32, reason string) tunnelframe.Frame {
	return tunnelframe.Frame{
		Type:     tunnelframe.TypeStreamCancel,
		StreamID: id,
		Payload:  tunnelframe.EncodeJSON(tunnelframe.StreamCancelPayload{Code: 499, Reason: reason}),
	}
}

func creditFrame(id uint32, n int) tunnelframe.Frame {
	return tunnelframe.Frame{
		Type:     tunnelframe.TypeCredit,
		StreamID: id,
		Payload:  tunnelframe.EncodeJSON(tunnelframe.CreditPayload{Bytes: n}),
	}
}

// statusForCode maps a RESP_ERROR code to an HTTP status.
func statusForCode(code string) int {
	switch code {
	case "target_busy":
		return http.StatusTooManyRequests
	case "target_offline", "target_unreachable", "tunnel_closed":
		return http.StatusBadGateway
	default:
		return http.StatusBadGateway
	}
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
