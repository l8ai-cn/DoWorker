package sessionapi

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	relayMsgSnapshot = 0x01
	relayMsgOutput   = 0x02
	relayMsgInput    = 0x03
	relayMsgResize   = 0x04
	relayMsgResync   = 0x0A
)

type relaySnapshot struct {
	SerializedContent string `json:"serialized_content"`
}

func encodeRelayInput(payload []byte) []byte {
	out := make([]byte, 1+len(payload))
	out[0] = relayMsgInput
	copy(out[1:], payload)
	return out
}

func encodeRelayResize(cols, rows uint16) []byte {
	out := make([]byte, 5)
	out[0] = relayMsgResize
	binary.BigEndian.PutUint16(out[1:3], cols)
	binary.BigEndian.PutUint16(out[3:5], rows)
	return out
}

func decodeRelayFrame(data []byte) (msgType byte, payload []byte, err error) {
	if len(data) < 1 {
		return 0, nil, errors.New("empty frame")
	}
	return data[0], data[1:], nil
}

func relayOutputToClient(msgType byte, payload []byte) ([]byte, bool) {
	switch msgType {
	case relayMsgOutput:
		return payload, true
	case relayMsgSnapshot:
		var snap relaySnapshot
		if json.Unmarshal(payload, &snap) != nil || snap.SerializedContent == "" {
			return nil, false
		}
		clear := "\x1b[2J\x1b[H\x1b[3J"
		return append([]byte(clear), []byte(snap.SerializedContent)...), true
	default:
		return nil, false
	}
}

type terminalResizeWire struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

func parseTerminalResizeWire(text string) (cols, rows uint16, ok bool) {
	var r terminalResizeWire
	if json.Unmarshal([]byte(text), &r) != nil || r.Type != "resize" {
		return 0, 0, false
	}
	if r.Cols <= 0 || r.Rows <= 0 {
		return 0, 0, false
	}
	return uint16(r.Cols), uint16(r.Rows), true
}

func bridgeTerminalWS(client, upstream *websocket.Conn, readOnly bool) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			msgType, data, err := upstream.ReadMessage()
			if err != nil {
				return
			}
			if msgType != websocket.BinaryMessage {
				continue
			}
			typ, payload, err := decodeRelayFrame(data)
			if err != nil {
				continue
			}
			out, ok := relayOutputToClient(typ, payload)
			if !ok {
				continue
			}
			if err := client.WriteMessage(websocket.BinaryMessage, out); err != nil {
				return
			}
		}
	}()
	for {
		select {
		case <-done:
			return
		default:
		}
		msgType, data, err := client.ReadMessage()
		if err != nil {
			return
		}
		if readOnly {
			continue
		}
		switch msgType {
		case websocket.BinaryMessage:
			_ = upstream.WriteMessage(websocket.BinaryMessage, encodeRelayInput(data))
		case websocket.TextMessage:
			if cols, rows, ok := parseTerminalResizeWire(string(data)); ok {
				_ = upstream.WriteMessage(websocket.BinaryMessage, encodeRelayResize(cols, rows))
			}
		}
	}
}

func dialRelayWS(relayURL, token string) (*websocket.Conn, *http.Response, error) {
	wsURL := strings.Replace(relayURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	if !strings.HasSuffix(wsURL, "/") {
		wsURL += "/"
	}
	wsURL += "browser/relay?token=" + token
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	return dialer.Dial(wsURL, nil)
}
