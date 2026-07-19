package sessionapi

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestForkSessionRejectsSameAgentWorkerConfigChange(t *testing.T) {
	deps, _, _ := setupSessionCreationCompensationTest(t)
	seedSourceSession(t, deps, nil)

	response := sessionCreationRequest(
		deps.handleForkSession,
		"/v1/sessions/conv_source/fork",
		gin.Params{{Key: "id", Value: "conv_source"}},
		`{"agent_id":"codex-cli","model_resource_id":42}`,
	)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.JSONEq(t, `{
		"error":"same-agent operation cannot change worker configuration",
		"code":"validation_failed"
	}`, response.Body.String())
}
