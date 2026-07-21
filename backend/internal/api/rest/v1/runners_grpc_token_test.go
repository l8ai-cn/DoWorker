package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRespondRegisterWithTokenErrorTreatsMissingClusterAsInvalidToken(t *testing.T) {
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)

	respondRegisterWithTokenError(ctx, runner.ErrExecutionClusterNotFound)

	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.JSONEq(t, `{"error":"Invalid token","code":"INVALID_TOKEN"}`, response.Body.String())
}
