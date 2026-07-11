package acp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

type HandshakeHook func(HandshakeRequester, json.RawMessage) error

type HandshakeRequester interface {
	Request(method string, params any) (json.RawMessage, error)
}

type handshakeRequester struct {
	transport *ACPTransport
}

func NewACPTransportWithHandshakeHook(
	callbacks EventCallbacks,
	logger *slog.Logger,
	hook HandshakeHook,
) *ACPTransport {
	transport := NewACPTransport(callbacks, logger)
	transport.handshakeHook = hook
	return transport
}

func (r handshakeRequester) Request(method string, params any) (json.RawMessage, error) {
	request, err := r.transport.tracker.SendRequest(method, params)
	if err != nil {
		return nil, fmt.Errorf("write %s: %w", method, err)
	}
	response, err := r.transport.tracker.WaitResponse(request, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("wait %s response: %w", method, err)
	}
	if response.Error != nil {
		return nil, fmt.Errorf("%s error: code=%d msg=%s",
			method, response.Error.Code, response.Error.Message)
	}
	return response.Result, nil
}
