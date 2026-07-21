package grpc

import (
	"context"
	"log/slog"
	"testing"

	workflowDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workflow"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	workflowService "github.com/l8ai-cn/agentcloud/backend/internal/service/workflow"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestMcpCreateWorkflowRequiresResourceApply(t *testing.T) {
	adapter := &GRPCRunnerAdapter{}
	tenant := &middleware.TenantContext{OrganizationID: 1, UserID: 1}

	_, err := adapter.mcpCreateWorkflow(
		context.Background(),
		tenant,
		"",
		[]byte(`{"name":"Daily Review","prompt_template":"review changed files"}`),
	)

	require.NotNil(t, err)
	require.Equal(t, int32(409), err.code)
	require.Contains(t, err.message, "validate-plan-apply")
}

func TestMcpTriggerWorkflowReportsResourceApplyPrecondition(t *testing.T) {
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
		OrganizationID: 1,
		CreatedByID:    1,
		Name:           "Legacy workflow",
		Slug:           "legacy-workflow",
		PromptTemplate: "legacy",
		ExecutionMode:  workflowDomain.ExecutionModeDirect,
	})
	require.NoError(t, err)
	adapter := &GRPCRunnerAdapter{
		workflowService:      workflowSvc,
		workflowRunService:   runSvc,
		workflowOrchestrator: orchestrator,
	}

	_, mcpErr := adapter.mcpTriggerWorkflow(
		context.Background(),
		&middleware.TenantContext{OrganizationID: 1, UserID: 1},
		[]byte(`{"workflow_slug":"legacy-workflow"}`),
	)

	require.NotNil(t, mcpErr)
	require.Equal(t, int32(409), mcpErr.code)
	require.Contains(t, mcpErr.message, "resource binding")
}
