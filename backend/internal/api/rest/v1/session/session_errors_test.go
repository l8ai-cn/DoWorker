package sessionapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	agentsessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestWriteOrchestratorErrorMapsDisabledModelResource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	writeOrchestratorError(context, airesource.ErrDisabled)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "model_resource_disabled", body["code"])
	require.Equal(t, "selected model resource is disabled", body["error"])
}

func TestWriteOrchestratorErrorMapsStaleSessionBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	writeOrchestratorError(context, errors.Join(
		agentpod.ErrSessionProvisionFailed,
		agentsessionsvc.ErrSessionBindingChanged,
	))

	require.Equal(t, http.StatusConflict, recorder.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "session_binding_changed", body["code"])
}
