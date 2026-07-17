package sessionapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestEmbedWriteCapabilityCannotControlSession(t *testing.T) {
	deps := embedContextDeps(t)
	router := gin.New()
	registerEmbedRoutes(router.Group("/v1"), *deps)
	token := embedSessionToken(t, deps, []string{"read", "write"})

	for _, eventType := range []string{"interrupt", "stop_session"} {
		response := embedSessionEventRequest(router, token, eventType)
		assert.Equal(t, http.StatusForbidden, response.Code, eventType)
	}
}

func TestEmbedControlCapabilityCanInterruptWithoutWrite(t *testing.T) {
	deps := embedContextDeps(t)
	router := gin.New()
	registerEmbedRoutes(router.Group("/v1"), *deps)
	token := embedSessionToken(t, deps, []string{"read", "control"})

	response := embedSessionEventRequest(router, token, "interrupt")

	assert.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

func embedSessionEventRequest(
	router *gin.Engine,
	token string,
	eventType string,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/embed/sessions/conv_embed/events",
		bytes.NewBufferString(`{"type":"`+eventType+`","data":{}}`),
	)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
