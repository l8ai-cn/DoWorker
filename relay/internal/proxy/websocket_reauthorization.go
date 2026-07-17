package proxy

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const maximumReauthorizationDuration = 5 * time.Second

func reauthorizeWebSocket(
	ctx context.Context,
	conn *websocket.Conn,
	params ProxyParams,
	errs chan<- error,
) {
	if params.Reauthorize == nil || params.ReauthorizeEvery <= 0 {
		return
	}
	ticker := time.NewTicker(params.ReauthorizeEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			authCtx, cancel := context.WithTimeout(
				ctx,
				reauthorizationTimeout(params.ReauthorizeEvery),
			)
			err := params.Reauthorize(authCtx)
			cancel()
			if err == nil {
				continue
			}
			_ = conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					websocket.ClosePolicyViolation,
					"preview authorization revoked",
				),
				time.Now().Add(time.Second),
			)
			errs <- fmt.Errorf("proxy: websocket authorization failed: %w", err)
			return
		}
	}
}

func reauthorizationTimeout(interval time.Duration) time.Duration {
	if interval < maximumReauthorizationDuration {
		return interval
	}
	return maximumReauthorizationDuration
}
