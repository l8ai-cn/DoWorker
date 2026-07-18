package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/relay/internal/protocol/tunnelframe"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

// ProxyWebSocket proxies a WebSocket upgrade request over a tunnel stream.
//
// Unlike a typical same-process reverse proxy, upgrading the browser
// connection is a one-way door: once we've sent the 101 response we can no
// longer report an HTTP error status. So the runner-side dial happens FIRST
// (REQ_START, wait for RESP_START/RESP_ERROR) and only once the upstream
// WebSocket is confirmed do we upgrade the browser side. This means dial
// failures still surface as normal HTTP error responses (502/504) instead of
// a WebSocket that opens and then immediately closes.
func ProxyWebSocket(ctx context.Context, tun tunnelIface, w http.ResponseWriter, r *http.Request, p ProxyParams) error {
	if !previewWebSocketOriginAllowed(r, p.ExpectedOrigin) {
		http.Error(w, "origin forbidden", http.StatusForbidden)
		return fmt.Errorf("proxy: websocket origin rejected")
	}
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}

	st := tun.OpenStream()
	streamOwned := true
	defer func() {
		if streamOwned {
			tun.CloseStream(st.ID)
		}
	}()

	clientIP := clientIP(r)
	proto := forwardedProto(r)
	header := SanitizeRequestHeaders(r.Header, clientIP, proto, r.Host, p.HiddenCookieName)

	reqStart := tunnelframe.ReqStartPayload{
		Method:      r.Method,
		Path:        p.Path,
		RawQuery:    p.RawQuery,
		Header:      header,
		PodKey:      p.PodKey,
		Target:      p.Target,
		IsWebSocket: true,
	}
	if err := tun.WriteFrame(tunnelframe.Frame{Type: tunnelframe.TypeReqStart, StreamID: st.ID, Payload: tunnelframe.EncodeJSON(reqStart)}); err != nil {
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return err
	}

	status, err := waitForWSUpstream(ctx, w, st)
	if err != nil {
		return err
	}
	if status != http.StatusSwitchingProtocols {
		http.Error(w, "websocket upgrade failed", http.StatusBadGateway)
		_ = tun.WriteFrame(cancelFrame(st.ID, "upstream did not switch protocols"))
		return fmt.Errorf("proxy: upstream returned status %d for websocket upgrade", status)
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(request *http.Request) bool {
			return previewWebSocketOriginAllowed(request, p.ExpectedOrigin)
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		_ = tun.WriteFrame(cancelFrame(st.ID, "client upgrade failed"))
		return err
	}
	defer conn.Close()
	streamOwned = false // ownership moves to the pumps; they close the stream on exit

	errCh := make(chan error, 2)
	go pumpBrowserToTunnel(ctx, tun, st, conn, errCh)
	go pumpTunnelToBrowser(ctx, tun, st, conn, errCh)
	go reauthorizeWebSocket(ctx, conn, p, errCh)
	err = <-errCh

	_ = tun.WriteFrame(tunnelframe.Frame{
		Type:     tunnelframe.TypeWSClose,
		StreamID: st.ID,
		Payload:  tunnelframe.EncodeJSON(tunnelframe.WSClosePayload{Code: websocket.CloseNormalClosure}),
	})
	tun.CloseStream(st.ID)
	return err
}

// waitForWSUpstream blocks until the runner replies with RESP_START (carrying
// the upstream's status, expected 101) or RESP_ERROR, translating errors to a
// normal HTTP response since the browser hasn't been upgraded yet.
func waitForWSUpstream(ctx context.Context, w http.ResponseWriter, st *tunnel.Stream) (int, error) {
	select {
	case <-ctx.Done():
		http.Error(w, "gateway timeout", http.StatusGatewayTimeout)
		return 0, ctx.Err()
	case f := <-st.RespChan():
		switch f.Type {
		case tunnelframe.TypeRespStart:
			var rs tunnelframe.RespStartPayload
			if err := json.Unmarshal(f.Payload, &rs); err != nil {
				http.Error(w, "bad gateway", http.StatusBadGateway)
				return 0, err
			}
			return rs.Status, nil
		case tunnelframe.TypeRespError:
			var re tunnelframe.RespErrorPayload
			_ = json.Unmarshal(f.Payload, &re)
			http.Error(w, re.Code, statusForCode(re.Code))
			return 0, fmt.Errorf("proxy: upstream websocket error %q", re.Code)
		default:
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return 0, fmt.Errorf("proxy: unexpected frame %v before websocket RESP_START", f.Type)
		}
	}
}

// pumpBrowserToTunnel relays browser->runner WebSocket frames as WS_DATA,
// credit-gated like any other outbound stream data.
func pumpBrowserToTunnel(ctx context.Context, tun tunnelIface, st *tunnel.Stream, conn *websocket.Conn, errCh chan<- error) {
	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}
		if len(data) > 0 {
			if err := st.AcquireSend(ctx, len(data)); err != nil {
				errCh <- err
				return
			}
		}
		payload := tunnelframe.EncodeJSON(tunnelframe.WSDataPayload{MessageType: mt, Data: data})
		if err := tun.WriteFrame(tunnelframe.Frame{Type: tunnelframe.TypeWSData, StreamID: st.ID, Payload: payload}); err != nil {
			errCh <- err
			return
		}
	}
}

// pumpTunnelToBrowser relays runner->browser WS_DATA frames, replenishing the
// peer's send window (via CREDIT) as each message is flushed to the browser.
func pumpTunnelToBrowser(ctx context.Context, tun tunnelIface, st *tunnel.Stream, conn *websocket.Conn, errCh chan<- error) {
	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		case f := <-st.RespChan():
			switch f.Type {
			case tunnelframe.TypeWSData:
				var wd tunnelframe.WSDataPayload
				if err := json.Unmarshal(f.Payload, &wd); err != nil {
					errCh <- err
					return
				}
				if err := conn.WriteMessage(wd.MessageType, wd.Data); err != nil {
					errCh <- err
					return
				}
				if len(wd.Data) > 0 {
					_ = tun.WriteFrame(creditFrame(st.ID, len(wd.Data)))
				}
			case tunnelframe.TypeWSClose, tunnelframe.TypeRespEnd, tunnelframe.TypeRespError:
				errCh <- nil
				return
			case tunnelframe.TypeCredit:
				var c tunnelframe.CreditPayload
				if json.Unmarshal(f.Payload, &c) == nil && c.Bytes > 0 {
					st.AddSendCredit(c.Bytes)
				}
			}
		}
	}
}
