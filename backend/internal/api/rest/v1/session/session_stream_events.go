package sessionapi

import "encoding/json"

const (
	sseTurnStarted          = "turn.started"
	sseTurnTextDelta        = "turn.text.delta"
	sseTurnItemDone         = "turn.item.done"
	sseTurnCompleted        = "turn.completed"
	sseTurnReasoningStarted = "turn.reasoning.started"
	sseTurnReasoningDelta   = "turn.reasoning.delta"
	sseElicitationRequest   = "turn.elicitation.request"
	sseElicitationResolved  = "turn.elicitation.resolved"

	sseSessionStatus        = "session.status"
	sseSessionInputConsumed = "session.input.consumed"
	sseSessionInterrupted   = "session.interrupted"
	sseSessionUsage         = "session.usage"
)

func formatSSE(eventType string, data map[string]any) string {
	body, _ := json.Marshal(data)
	return "event: " + eventType + "\ndata: " + string(body) + "\n\n"
}
