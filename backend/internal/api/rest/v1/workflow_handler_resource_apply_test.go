package v1

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	workflowService "github.com/l8ai-cn/agentcloud/backend/internal/service/workflow"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTriggerWorkflowReportsResourceApplyPrecondition(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testkit.SetupTestDB(t)
	workflowSvc := workflowService.NewWorkflowService(
		infra.NewWorkflowRepository(db),
	)
	runSvc := workflowService.NewWorkflowRunService(
		infra.NewWorkflowRunRepository(db),
	)
	bus := eventbus.NewEventBus(nil, slog.Default())
	t.Cleanup(func() { bus.Close() })
	orchestrator := workflowService.NewWorkflowOrchestrator(
		workflowSvc,
		runSvc,
		bus,
		slog.Default(),
	)
	_, err := workflowSvc.Create(context.Background(), &workflowService.CreateWorkflowRequest{
		OrganizationID: 7,
		CreatedByID:    9,
		Name:           "Legacy workflow",
		Slug:           "legacy-workflow",
		PromptTemplate: "legacy",
	})
	require.NoError(t, err)
	handler := NewWorkflowHandler(workflowSvc, runSvc, orchestrator, nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	ctx.Params = gin.Params{{Key: "workflow_slug", Value: "legacy-workflow"}}
	ctx.Set("tenant", &middleware.TenantContext{
		OrganizationID: 7,
		UserID:         9,
	})

	handler.TriggerWorkflow(ctx)

	require.Equal(t, http.StatusConflict, recorder.Code)
	var response apierr.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, apierr.WORKFLOW_RESOURCE_APPLY_REQUIRED, response.Code)
}
