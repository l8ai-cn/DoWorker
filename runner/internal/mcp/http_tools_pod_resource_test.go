package mcp

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/l8ai-cn/agentcloud/runner/internal/mcp/tools"
)

func TestCreatePodToolAcceptsOnlyWorkerPlanID(t *testing.T) {
	tool := (&HTTPServer{}).createCreatePodTool()
	properties := tool.InputSchema["properties"].(map[string]interface{})
	required := tool.InputSchema["required"].([]string)

	if !reflect.DeepEqual([]string{"plan_id"}, required) {
		t.Fatalf("required = %v, want [plan_id]", required)
	}
	if len(properties) != 1 {
		t.Fatalf("properties = %v, want only plan_id", properties)
	}
	plan, ok := properties["plan_id"].(map[string]interface{})
	if !ok || plan["type"] != "string" {
		t.Fatalf("plan_id schema = %v, want string", properties["plan_id"])
	}
	if tool.InputSchema["additionalProperties"] != false {
		t.Fatalf("additionalProperties = %v, want false", tool.InputSchema["additionalProperties"])
	}
}

func TestCreatePodToolConsumesPlanAndRequestsPodBinding(t *testing.T) {
	client := &resourcePodClient{}
	result, err := (&HTTPServer{}).createCreatePodTool().Handler(context.Background(), client, map[string]interface{}{
		"plan_id": "11111111-1111-4111-8111-111111111111",
	})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if client.createRequest == nil || client.createRequest.PlanID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("CreatePod() request = %#v", client.createRequest)
	}
	if client.bindingTarget != "new-pod" {
		t.Fatalf("RequestBinding() target = %q, want new-pod", client.bindingTarget)
	}
	wantScopes := []tools.BindingScope{tools.ScopePodRead, tools.ScopePodWrite}
	if !reflect.DeepEqual(client.bindingScopes, wantScopes) {
		t.Fatalf("RequestBinding() scopes = %v, want %v", client.bindingScopes, wantScopes)
	}
	if result != "Pod: new-pod | Status: initializing | Binding: #10 (pending)" {
		t.Fatalf("Handler() result = %v", result)
	}
}

func TestCreatePodToolRejectsFieldsOutsideWorkerPlan(t *testing.T) {
	client := &resourcePodClient{}
	_, err := (&HTTPServer{}).createCreatePodTool().Handler(context.Background(), client, map[string]interface{}{
		"plan_id": "11111111-1111-4111-8111-111111111111",
		"prompt":  "secret runtime override",
	})

	if err == nil || err.Error() != "create_pod accepts only plan_id" {
		t.Fatalf("Handler() error = %v", err)
	}
	if client.createRequest != nil {
		t.Fatalf("CreatePod() request = %#v, want no call", client.createRequest)
	}
}

func TestCreatePodToolReturnsErrorWhenAutomaticBindingFails(t *testing.T) {
	client := &resourcePodClient{
		bindingErr: errors.New("binding backend password=super-secret"),
	}
	_, err := (&HTTPServer{}).createCreatePodTool().Handler(context.Background(), client, map[string]interface{}{
		"plan_id": "11111111-1111-4111-8111-111111111111",
	})

	if err == nil || err.Error() != "pod new-pod was created but automatic binding failed" {
		t.Fatalf("Handler() error = %v", err)
	}
}

type resourcePodClient struct {
	tools.CollaborationClient
	createRequest *tools.PodCreateRequest
	bindingTarget string
	bindingScopes []tools.BindingScope
	bindingErr    error
}

func (c *resourcePodClient) CreatePod(_ context.Context, req *tools.PodCreateRequest) (*tools.PodCreateResponse, error) {
	c.createRequest = req
	return &tools.PodCreateResponse{PodKey: "new-pod", Status: "initializing"}, nil
}

func (c *resourcePodClient) RequestBinding(_ context.Context, target string, scopes []tools.BindingScope) (*tools.Binding, error) {
	c.bindingTarget = target
	c.bindingScopes = scopes
	if c.bindingErr != nil {
		return nil, c.bindingErr
	}
	return &tools.Binding{ID: 10, Status: tools.BindingStatusPending}, nil
}
