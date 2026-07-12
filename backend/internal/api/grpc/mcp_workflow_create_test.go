package grpc

import (
	"context"
	"testing"

	loopDomain "github.com/anthropics/agentsmesh/backend/internal/domain/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workflowService "github.com/anthropics/agentsmesh/backend/internal/service/workflow"
	"github.com/anthropics/agentsmesh/backend/internal/testkit"
)

// adapterWithWorkflowService builds a bare adapter backed by a real workflow service on
// an in-memory DB. podService is nil, so applyCallingPodDefaults early-returns
// and agent_slug must be supplied in the payload.
func adapterWithWorkflowService(t *testing.T) *GRPCRunnerAdapter {
	db := testkit.SetupTestDB(t)
	return &GRPCRunnerAdapter{
		workflowService: workflowService.NewWorkflowService(infra.NewWorkflowRepository(db)),
	}
}

func TestMcpCreateWorkflow_CreatesDisabledByDefault(t *testing.T) {
	a := adapterWithWorkflowService(t)
	tc := &middleware.TenantContext{OrganizationID: 1, UserID: 1}
	payload := []byte(`{"name":"Daily Review","prompt_template":"review changed files","agent_slug":"claude"}`)

	result, mcpErr := a.mcpCreateWorkflow(context.Background(), tc, "", payload)
	if mcpErr != nil {
		t.Fatalf("unexpected error: %v", mcpErr)
	}
	resp, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map: %T", result)
	}
	summary, ok := resp["workflow"].(*mcpWorkflowSummary)
	if !ok {
		t.Fatalf("workflow is not *mcpWorkflowSummary: %T", resp["workflow"])
	}
	if summary.Status != loopDomain.StatusDisabled {
		t.Errorf("status = %q, want disabled (lowest autonomy by default)", summary.Status)
	}
	if summary.Slug != "daily-review" {
		t.Errorf("slug = %q, want daily-review", summary.Slug)
	}
}

func TestMcpCreateWorkflow_EnabledWhenConfirmed(t *testing.T) {
	a := adapterWithWorkflowService(t)
	tc := &middleware.TenantContext{OrganizationID: 1, UserID: 1}
	payload := []byte(`{"name":"Nightly Sync","prompt_template":"sync","agent_slug":"claude","enabled":true}`)

	result, mcpErr := a.mcpCreateWorkflow(context.Background(), tc, "", payload)
	if mcpErr != nil {
		t.Fatalf("unexpected error: %v", mcpErr)
	}
	summary := result.(map[string]interface{})["workflow"].(*mcpWorkflowSummary)
	if summary.Status != loopDomain.StatusEnabled {
		t.Errorf("status = %q, want enabled when enabled=true", summary.Status)
	}
}

func TestMcpCreateWorkflow_RejectsInvalidCron(t *testing.T) {
	a := adapterWithWorkflowService(t)
	tc := &middleware.TenantContext{OrganizationID: 1, UserID: 1}
	payload := []byte(`{"name":"Bad Cron","prompt_template":"x","agent_slug":"claude","cron_expression":"not a cron"}`)

	_, mcpErr := a.mcpCreateWorkflow(context.Background(), tc, "", payload)
	if mcpErr == nil || mcpErr.code != 400 {
		t.Fatalf("expected 400 for invalid cron, got %v", mcpErr)
	}
}
