package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func TestCreatePodToolRequiresOnlyResourceManifest(t *testing.T) {
	tool := new(HTTPServer).createCreatePodTool()
	properties := tool.InputSchema["properties"].(map[string]interface{})
	if len(properties) != 1 || properties["resource"] == nil {
		t.Fatalf("create_pod properties = %#v", properties)
	}
	required := tool.InputSchema["required"].([]string)
	if len(required) != 1 || required[0] != "resource" {
		t.Fatalf("create_pod required = %#v", required)
	}
	if tool.InputSchema["additionalProperties"] != false {
		t.Fatal("create_pod must reject legacy fields")
	}
}

func TestCreateWorkflowToolRequiresResourceManifest(t *testing.T) {
	tool := new(HTTPServer).createCreateWorkflowTool()
	properties := tool.InputSchema["properties"].(map[string]interface{})
	if len(properties) != 2 || properties["resource"] == nil ||
		properties["enabled"] == nil {
		t.Fatalf("create_workflow properties = %#v", properties)
	}
	required := tool.InputSchema["required"].([]string)
	if len(required) != 1 || required[0] != "resource" {
		t.Fatalf("create_workflow required = %#v", required)
	}
	if tool.InputSchema["additionalProperties"] != false {
		t.Fatal("create_workflow must reject legacy fields")
	}
}

func TestResourceManifestArgument(t *testing.T) {
	resource, err := resourceManifestArgument(map[string]interface{}{
		"resource": map[string]interface{}{
			"apiVersion": "agentsmesh.io/v1alpha1",
			"kind":       "Worker",
		},
	})
	if err != nil {
		t.Fatalf("resourceManifestArgument() error = %v", err)
	}
	var manifest map[string]interface{}
	if err := json.Unmarshal(resource, &manifest); err != nil {
		t.Fatalf("decode resource = %v", err)
	}
	if manifest["kind"] != "Worker" {
		t.Fatalf("resource kind = %#v", manifest["kind"])
	}
}

func TestResourceManifestArgumentRejectsNonObject(t *testing.T) {
	for _, args := range []map[string]interface{}{
		{},
		{"resource": "kind: Worker"},
		{"resource": []interface{}{}},
	} {
		if _, err := resourceManifestArgument(args); err == nil {
			t.Fatalf("resourceManifestArgument(%#v) succeeded", args)
		}
	}
}

type createPodBindingFailureClient struct {
	tools.CollaborationClient
}

func (*createPodBindingFailureClient) CreatePod(
	context.Context,
	*tools.PodCreateRequest,
) (*tools.PodCreateResponse, error) {
	return &tools.PodCreateResponse{
		PodKey: "created-without-binding",
		Status: string(tools.PodStatusInitializing),
	}, nil
}

func (*createPodBindingFailureClient) RequestBinding(
	context.Context,
	string,
	[]tools.BindingScope,
) (*tools.Binding, error) {
	return nil, errors.New("binding denied")
}

func TestCreatePodToolReturnsBindingFailure(t *testing.T) {
	tool := new(HTTPServer).createCreatePodTool()
	result, err := tool.Handler(
		context.Background(),
		&createPodBindingFailureClient{},
		map[string]interface{}{
			"resource": map[string]interface{}{
				"apiVersion": "agentsmesh.io/v1alpha1",
				"kind":       "Worker",
			},
		},
	)
	if err == nil || err.Error() !=
		"bind created pod created-without-binding: binding denied" {
		t.Fatalf("Handler() error = %v", err)
	}
	if result != nil {
		t.Fatalf("Handler() result = %#v, want nil", result)
	}
}
