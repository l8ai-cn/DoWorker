package sessionapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	sessionmessagesvc "github.com/anthropics/agentsmesh/backend/internal/service/sessionmessage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		&podDomain.Pod{PodKey: "pod-1", RunnerID: 7},
		jsonRaw(`{"content":[{"type":"input_text","text":"hello"}]}`),
	)

	assert.Equal(t, http.StatusAccepted, response.Code)
	require.NotNil(t, outbox.input.Item)
	assert.Equal(t, "conv-1", outbox.input.Item.SessionID)
	assert.Equal(t, "hello", outbox.input.Prompt)
}

func jsonRaw(value string) []byte {
	return []byte(value)
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
