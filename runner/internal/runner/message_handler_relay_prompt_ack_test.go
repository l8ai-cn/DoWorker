package runner

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/relay"
)

func TestHandleAcpRelayCommandConfirmsAcceptedPrompt(t *testing.T) {
	handler := newTestHandler()
	client := relay.NewMockClient("wss://relay.example")
	client.SetConnected(true)
	pod := &Pod{PodKey: "test-pod", IO: &mockPodIO{}}
	pod.Relay = NewACPPodRelay(pod.PodKey, nil, nil)
	pod.SetRelayClient(client)

	handler.handleAcpRelayCommand(pod, promptRelayPayload("request-1"))

	event := relayEventPayload(t, client)
	if event["type"] != "contentChunk" || event["requestId"] != "request-1" {
		t.Fatalf("accepted prompt event = %#v", event)
	}
}

func TestHandleAcpRelayCommandReportsPromptFailure(t *testing.T) {
	handler := newTestHandler()
	client := relay.NewMockClient("wss://relay.example")
	client.SetConnected(true)
	pod := &Pod{PodKey: "test-pod", IO: &mockPodIO{sendErr: errors.New("agent unavailable")}}
	pod.Relay = NewACPPodRelay(pod.PodKey, nil, nil)
	pod.SetRelayClient(client)

	handler.handleAcpRelayCommand(pod, promptRelayPayload("request-1"))

	event := relayEventPayload(t, client)
	if event["type"] != "commandFailed" || event["requestId"] != "request-1" {
		t.Fatalf("failed prompt event = %#v", event)
	}
}

func promptRelayPayload(requestID string) []byte {
	payload, _ := json.Marshal(map[string]string{
		"type": "prompt", "prompt": "hello", "requestId": requestID,
	})
	return payload
}

func relayEventPayload(t *testing.T, client *relay.MockClient) map[string]any {
	t.Helper()
	if client.CountSentByType(relay.MsgTypeAcpEvent) != 1 {
		t.Fatalf("Relay event count = %d", client.CountSentByType(relay.MsgTypeAcpEvent))
	}
	var event map[string]any
	if err := json.Unmarshal(client.SentMessages[0].Payload, &event); err != nil {
		t.Fatal(err)
	}
	return event
}
