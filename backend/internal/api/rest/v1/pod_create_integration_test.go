package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestPodCreateAPI_InvalidJSON(t *testing.T) {
	// The handler calls ShouldBindJSON first; malformed JSON fails before
	// reaching the orchestrator, so we can pass a nil orchestrator.
	handler := &PodHandler{orchestrator: nil}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/organizations/test/pods", bytes.NewBufferString("{bad json"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreatePod(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "invalid character")
}

func TestPodCreateAPI_FreshCreateRequiresResourceApply(t *testing.T) {
	handler := &PodHandler{}
	body := `{"agent_slug":"codex-cli","runner_id":1,"cols":80,"rows":24}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/organizations/test/pods", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	// Set tenant context (required by handler before calling orchestrator)
	c.Set("tenant", &middleware.TenantContext{
		OrganizationID:   1,
		OrganizationSlug: "test-org",
		UserID:           100,
		UserRole:         "owner",
	})
	c.Set("user_id", int64(100))

	handler.CreatePod(c)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "WORKER_RESOURCE_APPLY_REQUIRED", resp["code"])
}

func TestPodCreateAPI_LegacyAliasUsesResourceApplyGate(t *testing.T) {
	handler := &PodHandler{orchestrator: nil}

	longAlias := string(make([]byte, 101))
	for i := range longAlias {
		longAlias = longAlias[:i] + "a" + longAlias[i+1:]
	}

	reqBody := CreatePodRequest{
		AgentSlug: "claude-code",
		Alias:     &longAlias,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/organizations/test/pods", bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreatePod(c)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "WORKER_RESOURCE_APPLY_REQUIRED", resp["code"])
}

func TestPodCreateAPI_EmptyAliasStillUsesResourceApplyGate(t *testing.T) {
	handler := &PodHandler{}

	emptyAlias := "   "
	reqBody := CreatePodRequest{
		AgentSlug: "codex-cli",
		Alias:     &emptyAlias,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/organizations/test/pods", bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenant", &middleware.TenantContext{
		OrganizationID: 1, OrganizationSlug: "test-org",
		UserID: 100, UserRole: "owner",
	})
	c.Set("user_id", int64(100))

	handler.CreatePod(c)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "WORKER_RESOURCE_APPLY_REQUIRED", resp["code"])
}

// Backend equivalent of clients/core/crates/types/src/pod.rs
// `create_pod_request_resume_without_agent_slug` — guards the protocol
// contract that resume requests omit `agent_slug` (orchestrator inherits
// it from the source pod). A future `binding:"required"` on AgentSlug
// would silently reintroduce the original PR #340 bug; this test red-flags it.
func TestCreatePodRequest_ResumeWithoutAgentSlug_Unmarshals(t *testing.T) {
	body := `{"source_pod_key":"pod-source-123","resume_agent_session":true,"cols":80,"rows":24}`

	var req CreatePodRequest
	err := json.Unmarshal([]byte(body), &req)
	require.NoError(t, err)

	assert.Empty(t, req.AgentSlug, "AgentSlug must accept missing field for resume mode")
	assert.Equal(t, "pod-source-123", req.SourcePodKey)
	assert.Zero(t, req.RunnerID)
	assert.Equal(t, int32(80), req.Cols)
	assert.Equal(t, int32(24), req.Rows)
	require.NotNil(t, req.ResumeAgentSession)
	assert.True(t, *req.ResumeAgentSession)
}

func TestPodCreateAPI_ResumeRejectsRuntimeOverrides(t *testing.T) {
	handler := &PodHandler{}
	body := `{"source_pod_key":"pod-source-123","runner_id":1}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/organizations/test/pods",
		bytes.NewBufferString(body),
	)

	handler.CreatePod(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "WORKER_RESUME_LINEAGE_ONLY", resp["code"])
}

func TestPodCreateAPI_ResumeBuildsLineageOnlyRequest(t *testing.T) {
	resume := true
	creator := &recordingRESTPodCreator{
		result: &agentpod.OrchestrateCreatePodResult{Pod: &podDomain.Pod{
			PodKey: "resumed-pod",
		}},
	}
	handler := &PodHandler{orchestrator: creator}
	bodyBytes, err := json.Marshal(CreatePodRequest{
		TicketSlug:         stringPointer("ticket-1"),
		Cols:               120,
		Rows:               40,
		SourcePodKey:       "pod-source-123",
		ResumeAgentSession: &resume,
		QueueIfOffline:     true,
		QueueTTLMinutes:    45,
	})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/organizations/test/pods",
		bytes.NewReader(bodyBytes),
	)
	c.Set("tenant", &middleware.TenantContext{
		OrganizationID: 1,
		UserID:         100,
	})

	handler.CreatePod(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, creator.request)
	assert.Equal(t, int64(1), creator.request.OrganizationID)
	assert.Equal(t, int64(100), creator.request.UserID)
	assert.Equal(t, "pod-source-123", creator.request.SourcePodKey)
	assert.Equal(t, &resume, creator.request.ResumeAgentSession)
	assert.Equal(t, int32(120), creator.request.Cols)
	assert.Equal(t, int32(40), creator.request.Rows)
	assert.Equal(t, 45*time.Minute, creator.request.QueueTTL)
	assert.Empty(t, creator.request.AgentSlug)
	assert.Zero(t, creator.request.RunnerID)
	assert.Nil(t, creator.request.AgentfileLayer)
	assert.Nil(t, creator.request.ModelResourceID)
	assert.Nil(t, creator.request.WorkerSpecDraft)
	assert.Nil(t, creator.request.WorkerSpecSnapshotID)
}

type recordingRESTPodCreator struct {
	request *agentpod.OrchestrateCreatePodRequest
	result  *agentpod.OrchestrateCreatePodResult
}

func (c *recordingRESTPodCreator) CreatePod(
	_ context.Context,
	request *agentpod.OrchestrateCreatePodRequest,
) (*agentpod.OrchestrateCreatePodResult, error) {
	c.request = request
	return c.result, nil
}

func stringPointer(value string) *string {
	return &value
}
