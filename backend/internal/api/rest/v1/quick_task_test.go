package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	agentsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

func postQuickTask(t *testing.T, h *PodHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/quick-tasks", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CreateQuickTask(c)
	return w
}

func TestQuickTask_EmptyPrompt_400(t *testing.T) {
	w := postQuickTask(t, &PodHandler{}, `{"prompt":"   "}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_FAILED", resp["code"])
}

func TestQuickTask_PromptTooLong_400(t *testing.T) {
	long := strings.Repeat("x", quickTaskPromptMaxLen+1)
	w := postQuickTask(t, &PodHandler{}, `{"prompt":"`+long+`"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestQuickTask_TTLOutOfRange_400(t *testing.T) {
	for _, ttl := range []string{"-5", "1441"} {
		w := postQuickTask(t, &PodHandler{}, `{"prompt":"do it","queue_ttl_minutes":`+ttl+`}`)
		assert.Equal(t, http.StatusBadRequest, w.Code, "ttl=%s", ttl)
	}
}

func TestQuickTask_InvalidJSON_400(t *testing.T) {
	w := postQuickTask(t, &PodHandler{}, `{"prompt":`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMapQuickTaskError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantErr  string
	}{
		{"queue full -> 429", podDomain.ErrQueueFull, http.StatusTooManyRequests, "QUEUE_FULL"},
		{"no runner -> 422", agentsvc.ErrNoAvailableRunner, http.StatusUnprocessableEntity, "NO_RUNNER_FOR_AGENT"},
		{"missing agent -> 404", agentsvc.ErrMissingAgentSlug, http.StatusNotFound, "AGENT_NOT_FOUND"},
		{"config build -> 500", agentsvc.ErrConfigBuildFailed, http.StatusInternalServerError, "CONFIG_BUILD_FAILED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			mapQuickTaskError(c, tt.err)
			assert.Equal(t, tt.wantCode, w.Code)
			var resp map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			assert.Equal(t, tt.wantErr, resp["code"])
		})
	}
}
