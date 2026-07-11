package tunnel

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

const maxPermanentConnectFailures = 5

type reconnectPolicy struct {
	initial, maximum time.Duration
	maxPermanent     int
	jitter           func(time.Duration) time.Duration
}

func defaultReconnectPolicy() reconnectPolicy {
	return reconnectPolicy{
		initial:      500 * time.Millisecond,
		maximum:      30 * time.Second,
		maxPermanent: maxPermanentConnectFailures,
		jitter:       randomJitter,
	}
}

func randomJitter(delay time.Duration) time.Duration {
	if delay <= 0 {
		return 0
	}
	spread := delay / 5
	return delay - spread + time.Duration(rand.Int63n(int64(2*spread+1)))
}

type connectError struct {
	err       error
	permanent bool
}

func (e *connectError) Error() string { return e.err.Error() }
func (e *connectError) Unwrap() error { return e.err }

func (c *Client) connectOnce() (*websocket.Conn, uint64, error) {
	tunnelURL, err := c.tunnelURL()
	if err != nil {
		return nil, 0, &connectError{err: err, permanent: true}
	}
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second, Proxy: http.ProxyFromEnvironment}
	var dialConn net.Conn
	dialer.NetDialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
		if err != nil {
			return nil, err
		}
		dialConn = conn
		c.connMu.Lock()
		if c.stopped.Load() {
			c.connMu.Unlock()
			_ = conn.Close()
			return nil, c.ctx.Err()
		}
		c.dialConn = conn
		c.connMu.Unlock()
		return conn, nil
	}
	conn, response, err := dialer.DialContext(c.ctx, tunnelURL.String(), nil)
	defer c.clearDialConnection(dialConn)
	if err != nil {
		permanent := errors.Is(err, websocket.ErrBadHandshake)
		if response != nil {
			status := response.StatusCode
			permanent = status >= 400 && status < 500 && status != http.StatusRequestTimeout &&
				status != http.StatusTooManyRequests
		}
		return nil, 0, &connectError{err: fmt.Errorf("tunnel dial failed: %w", err), permanent: permanent}
	}
	conn.SetReadLimit(8 * 1024 * 1024)
	if err := c.writeFrame(conn, c.helloFrame()); err != nil {
		_ = conn.Close()
		return nil, 0, fmt.Errorf("tunnel hello failed: %w", err)
	}

	c.connMu.Lock()
	if c.stopped.Load() {
		c.connMu.Unlock()
		_ = conn.Close()
		return nil, 0, c.ctx.Err()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.generation++
	generation := c.generation
	c.conn = conn
	c.connected.Store(true)
	c.connMu.Unlock()
	c.logger.Info("tunnel connected", "host", tunnelURL.Host)
	return conn, generation, nil
}

func (c *Client) connectionLoop() {
	conn, generation := c.currentConnection()
	if conn == nil {
		var err error
		conn, generation, err = c.connectOnce()
		if err != nil {
			c.reconnectLoop(err)
			return
		}
	}
	if err := c.readCurrentConnection(conn, generation); err != nil && !c.stopped.Load() {
		c.logger.Debug("tunnel connection lost", "error", err)
		c.reconnectLoop(nil)
	}
}

func (c *Client) reconnectLoop(initialErr error) {
	backoff := c.reconnect.initial
	permanentFailures := 0
	connectErr := initialErr

	for !c.stopped.Load() {
		if connectErr != nil {
			var classified *connectError
			if errors.As(connectErr, &classified) && classified.permanent {
				permanentFailures++
				if permanentFailures >= c.reconnect.maxPermanent {
					c.logger.Error("tunnel reconnect stopped after permanent handshake failures",
						"failures", permanentFailures, "error", connectErr)
					return
				}
			} else {
				permanentFailures = 0
			}
		}

		if !c.waitForReconnect(backoff) {
			return
		}
		conn, generation, err := c.connectOnce()
		if err != nil {
			connectErr = err
			backoff = nextBackoff(backoff, c.reconnect.maximum)
			continue
		}
		permanentFailures = 0

		err = c.readCurrentConnection(conn, generation)
		if c.stopped.Load() {
			return
		}
		c.logger.Debug("tunnel connection lost", "error", err)
		connectErr = nil
		backoff = nextBackoff(backoff, c.reconnect.maximum)
	}
}

func (c *Client) waitForReconnect(delay time.Duration) bool {
	timer := time.NewTimer(c.reconnect.jitter(delay))
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-c.ctx.Done():
		return false
	}
}

func (c *Client) readCurrentConnection(conn *websocket.Conn, generation uint64) error {
	for {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				c.disconnect(conn, generation)
				if c.stopped.Load() {
					return err
				}
				replacement, replacementGeneration := c.currentConnection()
				if replacement == nil || replacementGeneration == generation {
					return err
				}
				conn, generation = replacement, replacementGeneration
				break
			}
			frame, err := tunnelframe.Decode(data)
			if err != nil {
				continue
			}
			switch frame.Type {
			case tunnelframe.TypePing:
				if err := c.writeFrame(conn, tunnelframe.Frame{Type: tunnelframe.TypePong}); err != nil {
					c.disconnect(conn, generation)
					return err
				}
			case tunnelframe.TypePong:
			default:
				if c.dispatcher != nil {
					c.dispatcher.Dispatch(frame)
				}
			}
		}
	}
}
