package sessionapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	sessionDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentsession"
	itemDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/conversationitem"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	sessionService "github.com/l8ai-cn/agentcloud/backend/internal/service/agentsession"
	itemService "github.com/l8ai-cn/agentcloud/backend/internal/service/conversationitem"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateSessionRollsBackPodAndSessionWhenInitialItemPersistenceFails(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	require.NoError(t, failConversationItemInserts(db))

	response := createSessionCompensationRequest(deps, `{
		"agent_id":"codex-cli",
		"initial_items":[{
			"type":"message",
			"data":{"role":"user","content":[{"type":"input_text","text":"hello"}]}
		}]
	}`)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(0), activeSessionCount(t, db))
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
}

func TestCreateSessionRollsBackSessionAndItemsWhenPendingCommandPersistenceFails(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	require.NoError(t, failPendingCommandInserts(db))

	response := createSessionCompensationRequest(deps, `{
		"agent_id":"codex-cli",
		"initial_items":[{
			"type":"message",
			"data":{"role":"user","content":[{"type":"input_text","text":"hello"}]}
		}]
	}`)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(0), activeSessionCount(t, db))
	assert.Equal(t, int64(0), conversationItemCount(t, db))
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
}

func TestCreateSessionRemovesItemsPersistedBeforeLaterItemFailure(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	require.NoError(t, failSecondConversationItemInsert(db))

	response := createSessionCompensationRequest(deps, `{
		"agent_id":"codex-cli",
		"initial_items":[
			{"type":"message","data":{"role":"user","content":[{"type":"input_text","text":"first"}]}},
			{"type":"message","data":{"role":"user","content":[{"type":"input_text","text":"second"}]}}
		]
	}`)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(0), conversationItemCount(t, db))
	assert.Equal(t, int64(0), activeSessionCount(t, db))
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
}

func TestCreateSessionCompensationOutlivesCancelledRequest(t *testing.T) {
	deps, _, lifecycle := setupSessionCreationCompensationTest(t)
	requestContext, cancel := context.WithCancel(context.Background())
	cancel()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/v1/sessions",
		bytes.NewBufferString(withSessionWorkerPlan(`{"agent_id":"codex-cli"}`)),
	).WithContext(requestContext)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: 21, UserID: 11})

	deps.handleCreateSession(ctx)
	ctx.Writer.WriteHeaderNow()

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	require.Len(t, lifecycle.contextErrors, 1)
	assert.NoError(t, lifecycle.contextErrors[0])
}

func TestCreateSessionTriggersDrainAfterOwnerItemsAndCommandPersist(t *testing.T) {
	deps, db, _ := setupSessionCreationCompensationTest(t)
	orchestrator := deps.PodOrchestrator.(*fixedSessionPodOrchestrator)
	queue := deps.DispatchQueue.(*recordingSessionDispatchQueue)
	queue.beforeTrigger = func(runnerID int64) {
		assert.Equal(t, int64(3), runnerID)
		assert.Equal(t, int64(1), activeSessionCount(t, db))
		assert.Equal(t, int64(1), conversationItemCount(t, db))
		assert.Equal(t, int64(1), pendingCommandCount(t, db))
	}

	response := createSessionCompensationRequest(deps, `{
		"agent_id":"codex-cli",
		"initial_items":[{
			"type":"message",
			"data":{"role":"user","content":[{"type":"input_text","text":"hello"}]}
		}]
	}`)

	assert.Equal(t, http.StatusOK, response.Code)
	require.NotNil(t, orchestrator.request)
	assert.True(t, orchestrator.request.DeferRunnerDispatch)
	assert.Equal(t, 0, orchestrator.dispatches)
	assert.Equal(t, []int64{3}, queue.triggers)
}

func TestForkSessionTerminatesPodWhenSessionPersistenceFails(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	seedSourceSession(t, deps, nil)
	require.NoError(t, failSessionInserts(db))

	response := forkSessionCompensationRequest(deps)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
}

func TestForkSessionRollsBackPodAndSessionWhenItemCopyFails(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	seedSourceSession(t, deps, &itemDomain.Item{
		ID: "item_source", SessionID: "conv_source", ItemType: "message",
		ResponseID: "resp_source", Status: "completed", Position: 1,
		Payload: []byte(`{"type":"message"}`), CreatedAt: time.Now(),
	})
	require.NoError(t, failConversationItemInserts(db))

	response := forkSessionCompensationRequest(deps)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(1), activeSessionCount(t, db))
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
}

func TestImportSessionTerminatesPodWhenSessionPersistenceFails(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	require.NoError(t, failSessionInserts(db))

	response := importSessionCompensationRequest(t, deps)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
}

func TestImportSessionRollsBackPodAndSessionWhenItemPersistenceFails(t *testing.T) {
	deps, db, lifecycle := setupSessionCreationCompensationTest(t)
	require.NoError(t, failConversationItemInserts(db))

	response := importSessionCompensationRequest(t, deps)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, []string{"new-pod"}, lifecycle.terminated)
	assert.Equal(t, int64(0), activeSessionCount(t, db))
	assert.Equal(t, int64(0), pendingCommandCount(t, db))
}

func setupSessionCreationCompensationTest(
	t *testing.T,
) (*Deps, *gorm.DB, *recordingSessionPodLifecycle) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/session-compensation.db"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&sessionDomain.Session{},
		&itemDomain.Item{},
		&sqlitePendingCommand{},
	))
	lifecycle := &recordingSessionPodLifecycle{}
	queue := &recordingSessionDispatchQueue{
		ttl:          time.Hour,
		enabled:      true,
		maxPerRunner: 20,
	}
	deps := &Deps{
		Sessions: sessionService.NewService(db),
		Items:    itemService.NewService(db),
		PodOrchestrator: &fixedSessionPodOrchestrator{
			result: &agentpod.OrchestrateCreatePodResult{
				Pod: &podDomain.Pod{
					PodKey: "new-pod", OrganizationID: 21, CreatedByID: 11,
					RunnerID: 3, AgentSlug: "codex-cli", Status: podDomain.StatusQueued,
				},
				DeferredCreateCommand: &runnerv1.CreatePodCommand{PodKey: "new-pod"},
			},
		},
		DeferredCommitter: sessionService.NewDeferredCommitter(db),
		DispatchQueue:     queue,
		PodCoordinator:    lifecycle,
		WorkerCreation: &recordingSessionWorkerDraftFactory{
			draft: sessionTestWorkerDraft(t, "do-agent"),
		},
	}
	return deps, db, lifecycle
}

func seedSourceSession(t *testing.T, deps *Deps, item *itemDomain.Item) {
	t.Helper()
	require.NoError(t, deps.Sessions.Create(context.Background(), &sessionDomain.Session{
		ID: "conv_source", OrganizationID: 21, UserID: 11,
		PodKey: "source-pod", AgentSlug: "codex-cli", Status: "idle",
	}))
	if item != nil {
		require.NoError(t, deps.Items.Append(context.Background(), item))
	}
}

func createSessionCompensationRequest(deps *Deps, body string) *httptest.ResponseRecorder {
	return sessionCreationRequest(deps.handleCreateSession, "/v1/sessions", nil, withSessionWorkerPlan(body))
}

func forkSessionCompensationRequest(deps *Deps) *httptest.ResponseRecorder {
	return sessionCreationRequest(
		deps.handleForkSession,
		"/v1/sessions/conv_source/fork",
		gin.Params{{Key: "id", Value: "conv_source"}},
		withSessionWorkerPlan(`{"agent_id":"do-agent"}`),
	)
}

func importSessionCompensationRequest(t *testing.T, deps *Deps) *httptest.ResponseRecorder {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rollout-test.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(
		"{\"timestamp\":\"t\",\"type\":\"session_meta\",\"payload\":{\"session_id\":\"source\"}}\n"+
			"{\"timestamp\":\"t\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":[{\"type\":\"input_text\",\"text\":\"hello\"}]}}\n",
	), 0o600))
	return sessionCreationRequest(
		deps.handleImportSession,
		"/v1/sessions/import",
		nil,
		withSessionWorkerPlan(`{"source_path":"`+path+`","agent_id":"do-agent"}`),
	)
}

func withSessionWorkerPlan(body string) string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		panic(err)
	}
	if _, ok := payload["worker_spec"]; !ok {
		payload["worker_spec"] = map[string]any{
			"options_revision":    "rev-1",
			"runtime_image_id":    4,
			"placement_policy":    "explicit",
			"compute_target_id":   8,
			"deployment_mode":     "pooled",
			"resource_profile_id": 9,
		}
	}
	if _, ok := payload["automation_level"]; !ok {
		payload["automation_level"] = "autonomous"
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func sessionCreationRequest(
	handler gin.HandlerFunc,
	path string,
	params gin.Params,
	body string,
) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = params
	ctx.Set("tenant", &middleware.TenantContext{OrganizationID: 21, UserID: 11})
	handler(ctx)
	ctx.Writer.WriteHeaderNow()
	return response
}

func failSessionInserts(db *gorm.DB) error {
	return db.Exec(`
		CREATE TRIGGER fail_session_insert
		BEFORE INSERT ON agent_sessions
		BEGIN
			SELECT RAISE(FAIL, 'session insert blocked');
		END
	`).Error
}

func failConversationItemInserts(db *gorm.DB) error {
	return db.Exec(`
		CREATE TRIGGER fail_item_insert
		BEFORE INSERT ON conversation_items
		BEGIN
			SELECT RAISE(FAIL, 'item insert blocked');
		END
	`).Error
}

func failSecondConversationItemInsert(db *gorm.DB) error {
	return db.Exec(`
		CREATE TRIGGER fail_second_item_insert
		BEFORE INSERT ON conversation_items
		WHEN NEW.position = 2
		BEGIN
			SELECT RAISE(FAIL, 'second item insert blocked');
		END
	`).Error
}

func failPendingCommandInserts(db *gorm.DB) error {
	return db.Exec(`
		CREATE TRIGGER fail_pending_command_insert
		BEFORE INSERT ON pending_runner_commands
		BEGIN
			SELECT RAISE(FAIL, 'pending command insert blocked');
		END
	`).Error
}

func activeSessionCount(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&sessionDomain.Session{}).
		Where("deleted_at IS NULL").
		Count(&count).Error)
	return count
}

func conversationItemCount(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&itemDomain.Item{}).Count(&count).Error)
	return count
}

type fixedSessionPodOrchestrator struct {
	request        *agentpod.OrchestrateCreatePodRequest
	result         *agentpod.OrchestrateCreatePodResult
	dispatches     int
	beforeDispatch func()
	dispatchErr    error
}

func (o *fixedSessionPodOrchestrator) CreatePod(
	_ context.Context,
	req *agentpod.OrchestrateCreatePodRequest,
) (*agentpod.OrchestrateCreatePodResult, error) {
	o.request = req
	return o.result, nil
}

func (o *fixedSessionPodOrchestrator) DispatchDeferredPod(
	_ context.Context,
	_ *agentpod.OrchestrateCreatePodRequest,
	result *agentpod.OrchestrateCreatePodResult,
) (*agentpod.OrchestrateCreatePodResult, error) {
	o.dispatches++
	if o.beforeDispatch != nil {
		o.beforeDispatch()
	}
	return result, o.dispatchErr
}
