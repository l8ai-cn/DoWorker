package runner

import (
	"encoding/json"
	"sync"

	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/relay"
)

// ACPPodRelay implements PodRelay for ACP-mode pods.
type ACPPodRelay struct {
	podKey    string
	acpClient *acp.ACPClient
	onCommand func([]byte) // closure bound to pod ref at creation
	sendMu    sync.Mutex
}

// NewACPPodRelay creates a PodRelay for ACP mode.
func NewACPPodRelay(podKey string, acpClient *acp.ACPClient, onCommand func([]byte)) *ACPPodRelay {
	return &ACPPodRelay{
		podKey:    podKey,
		acpClient: acpClient,
		onCommand: onCommand,
	}
}

func (r *ACPPodRelay) SetupHandlers(rc relay.RelayClient) {
	rc.SetMessageHandler(relay.MsgTypeAcpCommand, r.onCommand)
	rc.SetMessageHandler(relay.MsgTypeSnapshotRequest, func(_ []byte) {
		r.SendSnapshot(rc)
	})
}

func (r *ACPPodRelay) SendSnapshot(rc relay.RelayClient) {
	r.sendMu.Lock()
	defer r.sendMu.Unlock()
	if data := r.materializeSnapshot(); data != nil {
		if err := rc.Send(relay.MsgTypeAcpSnapshot, data); err != nil {
			logger.Pod().Warn("Failed to send ACP snapshot via relay", "pod_key", r.podKey, "error", err)
		}
	}
	if r.acpClient != nil {
		if loopalData := r.acpClient.LoopalSnapshot(); loopalData != nil {
			_ = rc.Send(relay.MsgTypeAcpEvent, loopalData)
		}
	}
}

func (r *ACPPodRelay) materializeSnapshot() []byte {
	if r.acpClient == nil {
		return nil
	}
	snapshot := r.acpClient.GetSessionSnapshot()
	data, err := json.Marshal(snapshot)
	if err != nil {
		logger.Pod().Error("Failed to marshal ACP snapshot", "pod_key", r.podKey, "error", err)
		return nil
	}
	return data
}

func (r *ACPPodRelay) OnRelayConnected(rc relay.RelayClient) {
	// No-op: ACP mode has no aggregator to wire
}

func (r *ACPPodRelay) OnRelayDisconnected() {
	// No-op: ACP mode has no aggregator to clear
}

// BroadcastEvent sends an ACP event via the cloud relay client. Best-effort:
// drops when the relay is absent or disconnected.
func (r *ACPPodRelay) BroadcastEvent(rc relay.RelayClient, msgType byte, payload []byte) {
	r.sendMu.Lock()
	defer r.sendMu.Unlock()
	if rc != nil && rc.IsConnected() {
		_ = rc.Send(msgType, payload)
	}
}

// Compile-time interface check.
var _ PodRelay = (*ACPPodRelay)(nil)
