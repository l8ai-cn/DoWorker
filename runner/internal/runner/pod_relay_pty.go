package runner

import (
	"sync"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
	"github.com/l8ai-cn/agentcloud/runner/internal/relay"
	"github.com/l8ai-cn/agentcloud/runner/internal/safego"
)

// PTYPodRelay implements PodRelay for PTY-mode pods.
type PTYPodRelay struct {
	podKey         string
	io             PodIO
	components     *PTYComponents
	lastSnapshotMu sync.Mutex
	lastSnapshot   []byte
}

// NewPTYPodRelay constructs a PodRelay for PTY mode.
func NewPTYPodRelay(podKey string, io PodIO, comps *PTYComponents) *PTYPodRelay {
	return &PTYPodRelay{podKey: podKey, io: io, components: comps}
}

func (r *PTYPodRelay) SetupHandlers(rc relay.RelayClient) {
	rc.SetMessageHandler(relay.MsgTypeInput, r.inputHandler())
	rc.SetMessageHandler(relay.MsgTypeResize, r.resizeHandler())
	rc.SetMessageHandler(relay.MsgTypeSnapshotRequest, func(_ []byte) {
		r.SendSnapshot(rc)
	})
}

func (r *PTYPodRelay) SendSnapshot(rc relay.RelayClient) {
	log := logger.Pod()
	data := r.materializeSnapshot()
	if data == nil {
		log.Warn("SendSnapshot: no snapshot available", "pod_key", r.podKey)
		return
	}
	_ = rc.Send(relay.MsgTypeSnapshot, data)

	vt := r.components.VirtualTerminal
	term := r.components.Terminal
	if vt != nil && vt.IsAltScreen() && term != nil {
		safego.Go("relay-snapshot-redraw", func() {
			time.Sleep(100 * time.Millisecond)
			if err := term.Redraw(); err != nil {
				log.Warn("Failed to redraw terminal after relay snapshot",
					"pod_key", r.podKey, "error", err)
			}
		})
	}
}

func (r *PTYPodRelay) OnRelayConnected(rc relay.RelayClient) {
	if r.components.Aggregator == nil {
		return
	}
	r.components.Aggregator.SetRelayClient(&cloudRelayWriter{cloud: rc})
}

func (r *PTYPodRelay) OnRelayDisconnected() {
	if r.components.Aggregator != nil {
		r.components.Aggregator.SetRelayClient(nil)
	}
}

// BroadcastEvent is a no-op for PTY pods; PTY output flows through the
// aggregator's cloud relay writer, not via discrete events.
func (r *PTYPodRelay) BroadcastEvent(_ relay.RelayClient, _ byte, _ []byte) {}

var _ PodRelay = (*PTYPodRelay)(nil)
