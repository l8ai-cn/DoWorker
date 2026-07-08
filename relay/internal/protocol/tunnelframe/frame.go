// Package tunnelframe implements the binary multiplexing frame protocol used by
// the gateway HTTP data plane. Each frame is [1B type][4B stream_id BE][payload].
// stream_id 0 is reserved for connection-level messages (HELLO/PING/PONG).
package tunnelframe

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"
)

// FrameType identifies the kind of a tunnel frame.
type FrameType byte

const (
	TypeHello        FrameType = 0x01
	TypePing         FrameType = 0x02
	TypePong         FrameType = 0x03
	TypeReqStart     FrameType = 0x10
	TypeReqBody      FrameType = 0x11
	TypeReqEnd       FrameType = 0x12
	TypeStreamCancel FrameType = 0x13
	TypeRespStart    FrameType = 0x20
	TypeRespBody     FrameType = 0x21
	TypeRespEnd      FrameType = 0x22
	TypeRespError    FrameType = 0x23
	TypeWSData       FrameType = 0x30
	TypeWSClose      FrameType = 0x31
	TypeCredit       FrameType = 0x40
)

// HeaderSize is the fixed frame header length (1B type + 4B stream_id).
const HeaderSize = 5

// MaxChunk is the maximum body chunk size carried in a single body frame.
const MaxChunk = 256 << 10

var ErrShortFrame = errors.New("tunnelframe: short frame")

// Frame is a decoded tunnel frame.
type Frame struct {
	Type     FrameType
	StreamID uint32
	Payload  []byte
}

// Encode serializes a frame into its wire format.
func Encode(f Frame) []byte {
	buf := make([]byte, HeaderSize+len(f.Payload))
	buf[0] = byte(f.Type)
	binary.BigEndian.PutUint32(buf[1:5], f.StreamID)
	copy(buf[5:], f.Payload)
	return buf
}

// Decode parses a frame from its wire format.
func Decode(raw []byte) (Frame, error) {
	if len(raw) < HeaderSize {
		return Frame{}, ErrShortFrame
	}
	// Copy payload so callers can retain it after the read buffer is reused.
	payload := make([]byte, len(raw)-HeaderSize)
	copy(payload, raw[HeaderSize:])
	return Frame{
		Type:     FrameType(raw[0]),
		StreamID: binary.BigEndian.Uint32(raw[1:5]),
		Payload:  payload,
	}, nil
}

// --- JSON payload types ---

// HelloPayload is the TUNNEL_HELLO body sent by the runner on connect.
type HelloPayload struct {
	RunnerID     string   `json:"runner_id"`
	OrgID        string   `json:"org_id,omitempty"`
	Version      string   `json:"version,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// ReqStartPayload is the REQ_START body describing an outgoing request.
type ReqStartPayload struct {
	Method        string      `json:"method"`
	Path          string      `json:"path"`
	RawQuery      string      `json:"raw_query,omitempty"`
	Header        http.Header `json:"header,omitempty"`
	PodKey        string      `json:"pod_key"`
	Target        string      `json:"target"`
	ContentLength int64       `json:"content_length,omitempty"`
	IsWebSocket   bool        `json:"is_websocket,omitempty"`
}

// RespStartPayload is the RESP_START body carrying status and headers.
type RespStartPayload struct {
	Status int         `json:"status"`
	Header http.Header `json:"header,omitempty"`
}

// RespErrorPayload is the RESP_ERROR body carrying an error code.
type RespErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

// WSClosePayload is the WS_CLOSE body.
type WSClosePayload struct {
	Code   int    `json:"code"`
	Reason string `json:"reason,omitempty"`
}

// StreamCancelPayload is the STREAM_CANCEL body.
type StreamCancelPayload struct {
	Code   int    `json:"code"`
	Reason string `json:"reason,omitempty"`
}

// CreditPayload is the CREDIT body: the number of bytes the receiver has
// flushed and is now willing to accept again (flow-control replenishment).
type CreditPayload struct {
	Bytes int `json:"bytes"`
}

// WSDataPayload wraps a single websocket frame relayed over the tunnel.
// MessageType matches gorilla/websocket message types (1=text, 2=binary).
type WSDataPayload struct {
	MessageType int    `json:"mt"`
	Data        []byte `json:"data"`
}

// MarshalJSON payload helpers -------------------------------------------------

// EncodeJSON serializes v as a frame payload; panics only on programmer error.
func EncodeJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
