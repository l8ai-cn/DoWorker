package runner

import (
	"encoding/json"

	"github.com/l8ai-cn/agentcloud/runner/internal/relay"
)

func sendAcpViaRelay(pod *Pod, eventType, sessionID string, data any) {
	if pod.Relay == nil {
		return
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	var flat map[string]any
	if err := json.Unmarshal(dataBytes, &flat); err != nil || flat == nil {
		flat = map[string]any{}
	}
	flat["type"] = eventType
	flat["sessionId"] = sessionID
	payload, err := json.Marshal(flat)
	if err != nil {
		return
	}
	pod.Relay.BroadcastEvent(
		pod.GetRelayClient(),
		relay.MsgTypeAcpEvent,
		payload,
	)
}
