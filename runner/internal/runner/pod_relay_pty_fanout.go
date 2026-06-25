package runner

import (
	"github.com/anthropics/agentsmesh/runner/internal/relay"
)

// cloudRelayWriter adapts a relay.RelayClient to the aggregator's RelayWriter
// interface, mapping SendOutput to a MsgTypeOutput frame on the cloud relay.
type cloudRelayWriter struct {
	cloud relay.RelayClient
}

func (a *cloudRelayWriter) SendOutput(data []byte) error {
	if a.cloud != nil && a.cloud.IsConnected() {
		_ = a.cloud.Send(relay.MsgTypeOutput, data)
	}
	return nil
}

func (a *cloudRelayWriter) IsConnected() bool {
	return a.cloud != nil && a.cloud.IsConnected()
}
