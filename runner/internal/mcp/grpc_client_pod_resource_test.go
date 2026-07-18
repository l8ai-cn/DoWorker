package mcp

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func TestGRPCCollaborationClientCreatePodSendsOnlyPlanID(t *testing.T) {
	sender := &recordingMCPRequestSender{
		messages: make(chan *runnerv1.RunnerMessage, 1),
	}
	rpc := client.NewRPCClient(sender)
	defer rpc.Stop()

	collaboration := NewGRPCCollaborationClient(rpc, "source-pod")
	result := make(chan struct {
		pod *tools.PodCreateResponse
		err error
	}, 1)
	go func() {
		pod, err := collaboration.CreatePod(context.Background(), &tools.PodCreateRequest{
			PlanID: "11111111-1111-4111-8111-111111111111",
		})
		result <- struct {
			pod *tools.PodCreateResponse
			err error
		}{pod: pod, err: err}
	}()

	var message *runnerv1.RunnerMessage
	select {
	case message = <-sender.messages:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for MCP request")
	}
	request := message.GetMcpRequest()
	if request == nil {
		t.Fatal("MCP request was not sent")
	}
	if request.PodKey != "source-pod" || request.Method != "create_pod" {
		t.Fatalf("MCP request = pod %q method %q", request.PodKey, request.Method)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(request.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	wantPayload := map[string]interface{}{
		"plan_id": "11111111-1111-4111-8111-111111111111",
	}
	if !reflect.DeepEqual(payload, wantPayload) {
		t.Fatalf("payload = %#v, want %#v", payload, wantPayload)
	}

	rpc.HandleResponse(&runnerv1.McpResponse{
		RequestId: request.RequestId,
		Success:   true,
		Payload:   []byte(`{"pod":{"pod_key":"new-pod","status":"running"}}`),
	})
	call := <-result
	if call.err != nil {
		t.Fatalf("CreatePod() error = %v", call.err)
	}
	if call.pod == nil || call.pod.PodKey != "new-pod" || call.pod.Status != "running" {
		t.Fatalf("CreatePod() response = %#v", call.pod)
	}
}

type recordingMCPRequestSender struct {
	client.ConnectionSender
	messages chan *runnerv1.RunnerMessage
}

func (s *recordingMCPRequestSender) SendMessage(message *runnerv1.RunnerMessage) error {
	s.messages <- message
	return nil
}
