package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anthropics/agentsmesh/relay/internal/protocol/tunnelframe"
	"github.com/anthropics/agentsmesh/relay/internal/tunnel"
)

func pumpResponse(
	ctx context.Context,
	tun tunnelIface,
	st *tunnel.Stream,
	w http.ResponseWriter,
	counters *proxyByteCounters,
	hiddenCookieName string,
) (int, error) {
	flusher, _ := w.(http.Flusher)
	started := false
	status := 0
	for {
		select {
		case <-ctx.Done():
			_ = tun.WriteFrame(cancelFrame(st.ID, "gateway timeout"))
			if !started {
				http.Error(w, "gateway timeout", http.StatusGatewayTimeout)
				status = http.StatusGatewayTimeout
			}
			return status, ctx.Err()
		case frame := <-st.RespChan():
			switch frame.Type {
			case tunnelframe.TypeRespStart:
				var response tunnelframe.RespStartPayload
				if err := json.Unmarshal(frame.Payload, &response); err != nil {
					http.Error(w, "bad gateway", http.StatusBadGateway)
					return http.StatusBadGateway, err
				}
				copyHeader(w.Header(), SanitizeResponseHeaders(response.Header, hiddenCookieName))
				status = response.Status
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
					return http.StatusBadGateway, fmt.Errorf("proxy: RESP_BODY before RESP_START")
				}
				if len(frame.Payload) == 0 {
					continue
				}
				if _, err := w.Write(frame.Payload); err != nil {
					_ = tun.WriteFrame(cancelFrame(st.ID, "client write error"))
					return status, err
				}
				if flusher != nil {
					flusher.Flush()
				}
				counters.out.Add(int64(len(frame.Payload)))
				_ = tun.WriteFrame(creditFrame(st.ID, len(frame.Payload)))
			case tunnelframe.TypeRespEnd:
				return status, nil
			case tunnelframe.TypeRespError:
				var responseError tunnelframe.RespErrorPayload
				_ = json.Unmarshal(frame.Payload, &responseError)
				if !started {
					status = statusForCode(responseError.Code)
					http.Error(w, responseError.Code, status)
				}
				return status, fmt.Errorf("proxy: upstream error %q", responseError.Code)
			case tunnelframe.TypeCredit:
				var credit tunnelframe.CreditPayload
				if json.Unmarshal(frame.Payload, &credit) == nil && credit.Bytes > 0 {
					st.AddSendCredit(credit.Bytes)
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

func creditFrame(id uint32, bytes int) tunnelframe.Frame {
	return tunnelframe.Frame{
		Type:     tunnelframe.TypeCredit,
		StreamID: id,
		Payload:  tunnelframe.EncodeJSON(tunnelframe.CreditPayload{Bytes: bytes}),
	}
}

func statusForCode(code string) int {
	if code == "target_busy" {
		return http.StatusTooManyRequests
	}
	return http.StatusBadGateway
}
