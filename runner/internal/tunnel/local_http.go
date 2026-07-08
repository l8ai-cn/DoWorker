package tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

// frameSink is the outbound frame writer serveLocalHTTP replies on.
type frameSink interface {
	Send(tunnelframe.Frame) error
}

// hopByHopHeaders are stripped from upstream responses before relaying.
var hopByHopHeaders = []string{
	"Connection", "Proxy-Connection", "Keep-Alive",
	"Transfer-Encoding", "TE", "Trailer", "Upgrade",
}

// validateTarget enforces that the request target is a loopback address (or
// "localhost"). This is the runner-side SSRF guard: even though the gateway
// injects target via a signed token, the runner independently refuses to reach
// anything but its own pod's loopback ports.
func validateTarget(target string) error {
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		return fmt.Errorf("invalid target %q: %w", target, err)
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("target host %q is not loopback", host)
	}
	return nil
}

// serveLocalHTTP forwards a single tunneled HTTP request to the local loopback
// target and streams the response back as RESP_* frames. Outbound body frames
// are credit-gated via sw so the runner never outruns the gateway/browser.
func serveLocalHTTP(ctx context.Context, sink frameSink, id uint32, reqStart tunnelframe.ReqStartPayload, body io.Reader, sw *creditWindow) {
	if err := validateTarget(reqStart.Target); err != nil {
		_ = sink.Send(respError(id, "target_forbidden", err.Error()))
		return
	}

	rawURL := "http://" + reqStart.Target + reqStart.Path
	if reqStart.RawQuery != "" {
		rawURL += "?" + reqStart.RawQuery
	}
	req, err := http.NewRequestWithContext(ctx, reqStart.Method, rawURL, body)
	if err != nil {
		_ = sink.Send(respError(id, "bad_request", err.Error()))
		return
	}
	for k, vs := range reqStart.Header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Do(req)
	if err != nil {
		_ = sink.Send(respError(id, "target_unreachable", err.Error()))
		return
	}
	defer resp.Body.Close()

	_ = sink.Send(tunnelframe.Frame{
		Type:     tunnelframe.TypeRespStart,
		StreamID: id,
		Payload: tunnelframe.EncodeJSON(tunnelframe.RespStartPayload{
			Status: resp.StatusCode,
			Header: sanitizeResponseHeaders(resp.Header),
		}),
	})

	buf := make([]byte, tunnelframe.MaxChunk)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if err := sw.acquire(ctx, n); err != nil {
				return
			}
			chunk := make([]byte, n)
			copy(chunk, buf[:n])
			if err := sink.Send(tunnelframe.Frame{Type: tunnelframe.TypeRespBody, StreamID: id, Payload: chunk}); err != nil {
				return
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			_ = sink.Send(respError(id, "target_read_error", rerr.Error()))
			return
		}
	}
	_ = sink.Send(tunnelframe.Frame{Type: tunnelframe.TypeRespEnd, StreamID: id})
}

func respError(id uint32, code, msg string) tunnelframe.Frame {
	return tunnelframe.Frame{
		Type:     tunnelframe.TypeRespError,
		StreamID: id,
		Payload:  tunnelframe.EncodeJSON(tunnelframe.RespErrorPayload{Code: code, Message: msg}),
	}
}

func sanitizeResponseHeaders(in http.Header) http.Header {
	out := in.Clone()
	if out == nil {
		out = http.Header{}
	}
	for _, h := range hopByHopHeaders {
		out.Del(h)
	}
	return out
}
