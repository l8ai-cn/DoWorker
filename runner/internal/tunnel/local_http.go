package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/l8ai-cn/agentcloud/runner/internal/safego"
	"github.com/l8ai-cn/agentcloud/runner/internal/tunnelframe"
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

// serveLocalWebSocket dials the local loopback target as a WebSocket and
// relays frames bidirectionally over WS_DATA. Mirrors serveLocalHTTP's
// SSRF guard and error semantics: the upstream dial happens first, and only
// once it succeeds do we emit RESP_START(101) — a dial failure surfaces as
// RESP_ERROR so the gateway can answer the browser with a plain HTTP error
// instead of a half-open socket.
func serveLocalWebSocket(ctx context.Context, sink frameSink, id uint32, reqStart tunnelframe.ReqStartPayload, wsIn <-chan tunnelframe.Frame, sw *creditWindow) {
	if err := validateTarget(reqStart.Target); err != nil {
		_ = sink.Send(respError(id, "target_forbidden", err.Error()))
		return
	}

	dialURL := "ws://" + reqStart.Target + reqStart.Path
	if reqStart.RawQuery != "" {
		dialURL += "?" + reqStart.RawQuery
	}
	header := make(http.Header)
	for k, vs := range reqStart.Header {
		if isWebSocketHopHeader(k) {
			continue
		}
		for _, v := range vs {
			header.Add(k, v)
		}
	}

	dialer := websocket.Dialer{}
	upConn, _, err := dialer.DialContext(ctx, dialURL, header)
	if err != nil {
		_ = sink.Send(respError(id, "target_unreachable", err.Error()))
		return
	}
	defer upConn.Close()

	_ = sink.Send(tunnelframe.Frame{
		Type:     tunnelframe.TypeRespStart,
		StreamID: id,
		Payload:  tunnelframe.EncodeJSON(tunnelframe.RespStartPayload{Status: http.StatusSwitchingProtocols}),
	})

	// upstreamDone signals the read pump (upstream->tunnel) has exited, which
	// happens on upstream close/error; the write side below then unwinds too.
	upstreamDone := make(chan struct{})
	safego.Go("tunnel-local-ws-upstream", func() {
		defer close(upstreamDone)
		for {
			mt, data, rerr := upConn.ReadMessage()
			if rerr != nil {
				return
			}
			if len(data) > 0 {
				if err := sw.acquire(ctx, len(data)); err != nil {
					return
				}
			}
			if err := sink.Send(tunnelframe.Frame{
				Type:     tunnelframe.TypeWSData,
				StreamID: id,
				Payload:  tunnelframe.EncodeJSON(tunnelframe.WSDataPayload{MessageType: mt, Data: data}),
			}); err != nil {
				return
			}
		}
	})

	for {
		select {
		case <-ctx.Done():
			_ = upConn.Close()
			<-upstreamDone
			return
		case <-upstreamDone:
			return
		case f, ok := <-wsIn:
			if !ok {
				_ = upConn.Close()
				<-upstreamDone
				return
			}
			switch f.Type {
			case tunnelframe.TypeWSData:
				var wd tunnelframe.WSDataPayload
				if json.Unmarshal(f.Payload, &wd) == nil {
					if werr := upConn.WriteMessage(wd.MessageType, wd.Data); werr != nil {
						<-upstreamDone
						return
					}
				}
			case tunnelframe.TypeWSClose:
				_ = upConn.Close()
				<-upstreamDone
				return
			}
		}
	}
}

// isWebSocketHopHeader filters headers that must not be forwarded to the
// upstream WS dial (the websocket.Dialer sets its own Upgrade/Connection/Key
// negotiation headers).
func isWebSocketHopHeader(k string) bool {
	switch http.CanonicalHeaderKey(k) {
	case "Connection", "Upgrade", "Sec-Websocket-Key", "Sec-Websocket-Version",
		"Sec-Websocket-Extensions", "Sec-Websocket-Protocol":
		return true
	default:
		return false
	}
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
