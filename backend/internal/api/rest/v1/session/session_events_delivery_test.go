package sessionapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	sessionmessagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionmessage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPostMessageEventQueuesPersistedPrompt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/v1/sessions/conv-1/events",
		bytes.NewBufferString(`{"content":[{"type":"input_text","text":"hello"}]}`),
	)
	outbox := &promptOutboxStub{}

	deps := &Deps{
		MessageOutbox: outbox,
		Hub:           &SessionHub{},
	}
	deps.postMessageEvent(
		ctx,
		&sessionDomain.Session{ID: "conv-1", OrganizationID: 11},
		&podDomain.Pod{PodKey: "pod-1", RunnerID: 7, Status: podDomain.StatusRunning},
		jsonRaw(`{"content":[{"type":"input_text","text":"hello"}]}`),
	)

	assert.Equal(t, http.StatusAccepted, response.Code)
	require.NotNil(t, outbox.input.Item)
	assert.Equal(t, "conv-1", outbox.input.Item.SessionID)
	assert.Equal(t, "hello", outbox.input.Prompt)
}

func TestPostMessageEventResumesCompletedPodBeforeQueueingPrompt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/v1/sessions/conv-1/events",
		bytes.NewBufferString(`{"content":[{"type":"input_text","text":"continue"}]}`),
	)
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/session.db"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&sessionDomain.Session{}))
	sessions := sessionsvc.NewService(db)
	row := &sessionDomain.Session{
		ID: "conv-1", OrganizationID: 11, UserID: 5, PodKey: "completed-pod",
		AgentSlug: "codex-cli", Status: "idle", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, sessions.Create(context.Background(), row))
	resumed := &podDomain.Pod{
		PodKey: "resumed-pod", RunnerID: 7, OrganizationID: 11,
		CreatedByID: 5, AgentSlug: "codex-cli", Status: podDomain.StatusInitializing,
	}
	orchestrator := &podOrchestratorStub{
		result: &agentpod.OrchestrateCreatePodResult{Pod: resumed},
	}
	outbox := &promptOutboxStub{}
	deps := &Deps{
		MessageOutbox:   outbox,
		Hub:             &SessionHub{},
		PodOrchestrator: orchestrator,
		Sessions:        sessions,
	}

	deps.postMessageEvent(
		ctx,
		row,
		&podDomain.Pod{
			PodKey: "completed-pod", RunnerID: 7, OrganizationID: 11,
			CreatedByID: 5, AgentSlug: "codex-cli", Status: podDomain.StatusCompleted,
		},
		jsonRaw(`{"content":[{"type":"input_text","text":"continue"}]}`),
	)

	assert.Equal(t, http.StatusAccepted, response.Code)
	require.NotNil(t, orchestrator.request)
	assert.Equal(t, "completed-pod", orchestrator.request.SourcePodKey)
	assert.Empty(t, orchestrator.request.AgentSlug)
	assert.Equal(t, "conv-1", orchestrator.request.AgentSessionID)
	assert.Equal(t, "resumed-pod", outbox.input.PodKey)
	stored, err := sessions.Get(context.Background(), "conv-1")
	require.NoError(t, err)
	assert.Equal(t, "resumed-pod", stored.PodKey)
}

func jsonRaw(value string) []byte {
	return []byte(value)
}

type podOrchestratorStub struct {
	request *agentpod.OrchestrateCreatePodRequest
	result  *agentpod.OrchestrateCreatePodResult
	err     error
}

func (s *podOrchestratorStub) CreatePod(
	_ context.Context,
	request *agentpod.OrchestrateCreatePodRequest,
) (*agentpod.OrchestrateCreatePodResult, error) {
	s.request = request
	return s.result, s.err
}

func (s *podOrchestratorStub) DispatchDeferredPod(
	_ context.Context,
	_ *agentpod.OrchestrateCreatePodRequest,
	result *agentpod.OrchestrateCreatePodResult,
) (*agentpod.OrchestrateCreatePodResult, error) {
	return result, s.err
}

type promptOutboxStub struct {
	input sessionmessagesvc.PromptInput
	err   error
}

func (s *promptOutboxStub) PersistAndQueue(
	_ context.Context,
	input sessionmessagesvc.PromptInput,
) error {
	s.input = input
	return s.err
}
