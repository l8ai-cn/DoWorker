package workflowconnect

import (
	"context"
	"log/slog"
	"testing"

	"connectrpc.com/connect"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	workflowsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/workflow"
	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	workflowv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/workflow/v1"
	"github.com/stretchr/testify/require"
)

type loopOrg struct{}

func (loopOrg) GetID() int64    { return 7 }
func (loopOrg) GetSlug() string { return "acme" }
func (loopOrg) GetName() string { return "Acme" }

type loopOrgService struct{}

func (loopOrgService) GetBySlug(context.Context, string) (middleware.OrganizationGetter, error) {
	return loopOrg{}, nil
}

func (loopOrgService) IsMember(context.Context, int64, int64) (bool, error) {
	return true, nil
}

func (loopOrgService) GetMemberRole(context.Context, int64, int64) (string, error) {
	return "owner", nil
}

func loopContext() context.Context {
	return middleware.SetTenant(context.Background(), &middleware.TenantContext{UserID: 9})
}

func TestTriggerWorkflowReportsResourceApplyPrecondition(t *testing.T) {
	db := testkit.SetupTestDB(t)
	workflowService := workflowsvc.NewWorkflowService(
		infra.NewWorkflowRepository(db),
	)
	runService := workflowsvc.NewWorkflowRunService(
		infra.NewWorkflowRunRepository(db),
	)
	bus := eventbus.NewEventBus(nil, slog.Default())
	t.Cleanup(func() { bus.Close() })
	orchestrator := workflowsvc.NewWorkflowOrchestrator(
		workflowService,
		runService,
		bus,
		slog.Default(),
	)
	_, err := workflowService.Create(context.Background(), &workflowsvc.CreateWorkflowRequest{
		OrganizationID: 7,
		CreatedByID:    9,
		Name:           "Legacy workflow",
		Slug:           "legacy-workflow",
		PromptTemplate: "legacy",
	})
	require.NoError(t, err)
	server := NewServer(
		workflowService,
		runService,
		orchestrator,
		loopOrgService{},
		nil,
	)

	_, err = server.TriggerWorkflow(
		loopContext(),
		connect.NewRequest(&workflowv1.TriggerWorkflowRequest{
			OrgSlug:      "acme",
			WorkflowSlug: "legacy-workflow",
		}),
	)

	require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	require.Contains(t, err.Error(), "resource binding")
}
