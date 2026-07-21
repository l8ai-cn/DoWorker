package mcp

import (
	"context"
	"encoding/json"

	"github.com/l8ai-cn/agentcloud/runner/internal/mcp/tools"
)

// ==================== LoopClient ====================

// ListWorkflows lists workflows for the pod's organization.
func (c *GRPCCollaborationClient) ListWorkflows(ctx context.Context, status, query string, limit, offset int) ([]tools.WorkflowSummary, error) {
	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}
	if status != "" {
		params["status"] = status
	}
	if query != "" {
		params["query"] = query
	}
	var result struct {
		Loops []tools.WorkflowSummary `json:"workflows"`
	}
	if err := c.call(ctx, "list_workflows", params, &result); err != nil {
		return nil, err
	}
	return result.Loops, nil
}

// TriggerWorkflow triggers a workflow run by slug.
func (c *GRPCCollaborationClient) TriggerWorkflow(ctx context.Context, workflowSlug string, variables map[string]interface{}) (*tools.WorkflowTriggerResult, error) {
	params := map[string]interface{}{
		"workflow_slug": workflowSlug,
	}
	if len(variables) > 0 {
		varsJSON, err := json.Marshal(variables)
		if err != nil {
			return nil, err
		}
		params["variables"] = json.RawMessage(varsJSON)
	}

	var result tools.WorkflowTriggerResult
	if err := c.call(ctx, "trigger_workflow", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateWorkflow persists a new workflow for the pod's organization.
func (c *GRPCCollaborationClient) CreateWorkflow(ctx context.Context, req *tools.WorkflowCreateRequest) (*tools.WorkflowCreateResult, error) {
	var result tools.WorkflowCreateResult
	if err := c.call(ctx, "create_workflow", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
