package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type workbenchPublisherStub struct {
	podKey      string
	executionID string
	declaration json.RawMessage
	contextErr  error
}

func (s *workbenchPublisherStub) PublishWorkbenchArtifact(
	ctx context.Context,
	podKey string,
	executionID string,
	declaration json.RawMessage,
) (interface{}, error) {
	s.podKey = podKey
	s.executionID = executionID
	s.declaration = declaration
	s.contextErr = ctx.Err()
	return map[string]interface{}{"artifact_id": "demo-video", "revision": 1}, nil
}

func TestWorkbenchPublishArtifactToolUsesExactPodScope(t *testing.T) {
	server := NewHTTPServer(nil, 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "codex")
	publisher := &workbenchPublisherStub{}
	server.SetWorkbenchArtifactPublisher(publisher)
	body := bytes.NewBufferString(`{
		"jsonrpc":"2.0",
		"id":1,
		"method":"tools/call",
		"params":{
			"name":"workbench.publish_artifact",
			"arguments":{"declaration":{
				"schema_version":"agentsmesh.agent-workbench.artifact/v1",
				"artifact_id":"demo-video"
			}}
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "test-pod", publisher.podKey)
	require.NotEmpty(t, publisher.executionID)
	require.JSONEq(t, `{
		"schema_version":"agentsmesh.agent-workbench.artifact/v1",
		"artifact_id":"demo-video"
	}`, string(publisher.declaration))
	var response MCPResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	require.Nil(t, response.Error)
}

func TestWorkbenchPublishArtifactToolFailsWhenPublisherIsUnavailable(t *testing.T) {
	server := NewHTTPServer(nil, 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "codex")
	body := bytes.NewBufferString(`{
		"jsonrpc":"2.0",
		"id":1,
		"method":"tools/call",
		"params":{
			"name":"workbench.publish_artifact",
			"arguments":{"declaration":{}}
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/mcp", body)
	req.Header.Set("X-Pod-Key", "test-pod")
	rec := httptest.NewRecorder()

	server.handleMCP(rec, req)

	var response struct {
		Result MCPToolResult `json:"result"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	require.True(t, response.Result.IsError)
	require.Contains(t, response.Result.Content[0].Text, "publisher is unavailable")
}

func TestWorkbenchPublishArtifactToolUsesRequestContext(t *testing.T) {
	server := NewHTTPServer(nil, 9090)
	server.RegisterPod("test-pod", "test-org", nil, nil, "codex")
	publisher := &workbenchPublisherStub{}
	server.SetWorkbenchArtifactPublisher(publisher)
	body := bytes.NewBufferString(`{
		"jsonrpc":"2.0",
		"id":1,
		"method":"tools/call",
		"params":{
			"name":"workbench.publish_artifact",
			"arguments":{"declaration":{}}
		}
	}`)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/mcp", body).WithContext(ctx)
	req.Header.Set("X-Pod-Key", "test-pod")

	server.handleMCP(httptest.NewRecorder(), req)

	require.ErrorIs(t, publisher.contextErr, context.Canceled)
}

func TestWorkbenchProducerSchemaDoesNotAcceptToolExecutionID(t *testing.T) {
	properties := workbenchProducerSchema()["properties"].(map[string]interface{})

	require.NotContains(t, properties, "tool_execution_id")
}
