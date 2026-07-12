package sessionapi

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemsvc "github.com/anthropics/agentsmesh/backend/internal/service/conversationitem"
	runnerservice "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPostMessageEventDoesNotPersistWhenRunnerRejectsPrompt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/v1/sessions/conv-1/events",
		bytes.NewBufferString(`{"content":[{"type":"input_text","text":"hello"}]}`),
	)

	deps := &Deps{
		CommandSender: runnerservice.NewNoOpCommandSender(slog.Default()),
		Items:         itemsvc.NewService(nil),
		Hub:           &SessionHub{},
	}
	deps.postMessageEvent(
		ctx,
		&sessionDomain.Session{ID: "conv-1"},
		&podDomain.Pod{PodKey: "pod-1", RunnerID: 7},
		jsonRaw(`{"content":[{"type":"input_text","text":"hello"}]}`),
	)

	assert.Equal(t, http.StatusBadGateway, response.Code)
}

func jsonRaw(value string) []byte {
	return []byte(value)
}
