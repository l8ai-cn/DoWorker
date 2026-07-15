package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

type fakePodWorkspaceSandbox struct {
	connected bool
	result    *runnerv1.SandboxFsResultEvent
	err       error
	runnerID  int64
	command   *runnerv1.SandboxFsCommand
}

func (f *fakePodWorkspaceSandbox) IsConnected(int64) bool {
	return f.connected
}

func (f *fakePodWorkspaceSandbox) Exec(
	_ context.Context,
	runnerID int64,
	command *runnerv1.SandboxFsCommand,
) (*runnerv1.SandboxFsResultEvent, error) {
	f.runnerID = runnerID
	f.command = command
	return f.result, f.err
}

func TestPodWorkspaceChangesUsesAuthorizedPodRunner(t *testing.T) {
	sandbox := &fakePodWorkspaceSandbox{
		connected: true,
		result: &runnerv1.SandboxFsResultEvent{
			Changes: []*runnerv1.SandboxFsChange{{
				Path: "output/demo.mp4", Name: "demo.mp4", Status: "created", Bytes: 42,
			}},
		},
	}
	handler := podWorkspaceHandler(sandbox)
	recorder, ctx := podWorkspaceRequest(http.MethodGet, "/pods/worker-1/resources/workspace/changes", "worker-1")

	handler.ListWorkspaceArtifacts(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, sandbox.command)
	assert.Equal(t, int64(44), sandbox.runnerID)
	assert.Equal(t, "worker-1", sandbox.command.GetPodKey())
	assert.Equal(t, "changes", sandbox.command.GetOp())
	var body struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Len(t, body.Data, 1)
	assert.Equal(t, "output/demo.mp4", body.Data[0]["path"])
	assert.Equal(t, "created", body.Data[0]["status"])
}

func TestPodWorkspaceReadReturnsBinaryWirePayload(t *testing.T) {
	sandbox := &fakePodWorkspaceSandbox{
		connected: true,
		result: &runnerv1.SandboxFsResultEvent{
			Content: "AAAA", Encoding: "base64", ContentType: "video/mp4", FileBytes: 3,
		},
	}
	handler := podWorkspaceHandler(sandbox)
	recorder, ctx := podWorkspaceRequest(
		http.MethodGet,
		"/pods/worker-1/resources/workspace/filesystem/output/demo.mp4",
		"worker-1",
	)
	ctx.Params = append(ctx.Params, gin.Param{Key: "filepath", Value: "/output/demo.mp4"})

	handler.ReadWorkspaceArtifact(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "read", sandbox.command.GetOp())
	assert.Equal(t, "output/demo.mp4", sandbox.command.GetPath())
	var body map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, "video/mp4", body["content_type"])
	assert.Equal(t, "base64", body["encoding"])
	assert.Equal(t, "AAAA", body["content"])
}

func TestPodWorkspaceRejectsUnauthorizedReaderBeforeRunnerCall(t *testing.T) {
	sandbox := &fakePodWorkspaceSandbox{connected: true}
	handler := podWorkspaceHandler(sandbox)
	recorder, ctx := podWorkspaceRequest(http.MethodGet, "/pods/worker-1/resources/workspace/changes", "worker-1")
	setPodTenantContext(ctx, 2, 22)

	handler.ListWorkspaceArtifacts(ctx)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Nil(t, sandbox.command)
}

func TestPodWorkspaceReportsDisconnectedRunner(t *testing.T) {
	sandbox := &fakePodWorkspaceSandbox{connected: false}
	handler := podWorkspaceHandler(sandbox)
	recorder, ctx := podWorkspaceRequest(http.MethodGet, "/pods/worker-1/resources/workspace/changes", "worker-1")

	handler.ListWorkspaceArtifacts(ctx)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.JSONEq(
		t,
		`{"error":{"code":"runner_unavailable","message":"runner unavailable"}}`,
		recorder.Body.String(),
	)
	assert.Nil(t, sandbox.command)
}

func TestPodWorkspaceRejectsEmptyRunnerResponse(t *testing.T) {
	sandbox := &fakePodWorkspaceSandbox{connected: true}
	handler := podWorkspaceHandler(sandbox)
	recorder, ctx := podWorkspaceRequest(http.MethodGet, "/pods/worker-1/resources/workspace/changes", "worker-1")

	handler.ListWorkspaceArtifacts(ctx)

	assert.Equal(t, http.StatusBadGateway, recorder.Code)
	assert.JSONEq(
		t,
		`{"error":{"message":"runner returned an empty workspace response"}}`,
		recorder.Body.String(),
	)
}

func podWorkspaceHandler(sandbox *fakePodWorkspaceSandbox) *PodHandler {
	return &PodHandler{
		podService: &mockPodService{getPodFn: func(context.Context, string) (*agentpod.Pod, error) {
			return &agentpod.Pod{
				PodKey: "worker-1", RunnerID: 44, OrganizationID: 1, CreatedByID: 11,
			}, nil
		}},
		sandboxFs: sandbox,
	}
}

func podWorkspaceRequest(method, target, podKey string) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, target, nil)
	ctx.Params = gin.Params{{Key: "key", Value: podKey}}
	setPodTenantContext(ctx, 1, 11)
	return recorder, ctx
}
