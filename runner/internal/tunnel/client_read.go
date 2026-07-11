package tunnel

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"

	"github.com/anthropics/agentsmesh/runner/internal/tunnelframe"
)

func (c *Client) awaitHelloAck(conn *websocket.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(c.readyTimeout)); err != nil {
		return fmt.Errorf("tunnel HELLO_ACK deadline: %w", err)
	}
	messageType, data, err := conn.ReadMessage()
	_ = conn.SetReadDeadline(time.Time{})
	if err != nil {
		return fmt.Errorf("tunnel HELLO_ACK read failed: %w", err)
	}
	if messageType != websocket.BinaryMessage {
		return fmt.Errorf("tunnel HELLO_ACK: unexpected websocket message type %d", messageType)
	}
	frame, err := tunnelframe.Decode(data)
	if err != nil {
		return fmt.Errorf("tunnel HELLO_ACK decode failed: %w", err)
	}
	if frame.Type != tunnelframe.TypeHelloAck || frame.StreamID != 0 || len(frame.Payload) != 0 {
		return fmt.Errorf("tunnel HELLO_ACK: unexpected frame")
	}
	return nil
}

func (c *Client) readCurrentConnection(conn *websocket.Conn, generation uint64) error {
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			c.disconnect(conn, generation)
			if c.stopped.Load() {
				return err
			}
			replacement, replacementGeneration := c.currentConnection()
			if replacement != nil && replacementGeneration != generation {
				conn, generation = replacement, replacementGeneration
				continue
			}
			if c.dispatcher != nil {
				c.dispatcher.Close()
			}
			return err
		}
		frame, err := tunnelframe.Decode(data)
		if err != nil {
			continue
		}
		switch frame.Type {
		case tunnelframe.TypePing:
			if err := c.sendOnConnection(conn, generation, tunnelframe.Frame{Type: tunnelframe.TypePong}); err != nil {
				c.disconnect(conn, generation)
				if c.dispatcher != nil {
					c.dispatcher.Close()
				}
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
