package mcp

import (
	"context"
	"fmt"

	"github.com/l8ai-cn/agentcloud/runner/internal/mcp/tools"
)

// Workflow Tools

func (s *HTTPServer) createListWorkflowsTool() *MCPTool {
	return &MCPTool{
		Name:        "list_workflows",
		Description: "List automated workflows in the organization. Loops are repeatable tasks that can be triggered manually or via cron.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"status": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"enabled", "disabled", "archived"},
					"description": "Filter by workflow status",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query for workflow name",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum results (default: 20)",
				},
				"offset": map[string]interface{}{
					"type":        "integer",
					"description": "Pagination offset (default: 0)",
				},
			},
		},
		Handler: func(ctx context.Context, client tools.CollaborationClient, args map[string]interface{}) (interface{}, error) {
			status := getStringArg(args, "status")
			query := getStringArg(args, "query")

			limit := getIntArg(args, "limit")
			if limit == 0 {
				limit = 20
			}
			offset := getIntArg(args, "offset")

			result, err := client.ListWorkflows(ctx, status, query, limit, offset)
			if err != nil {
				return nil, err
			}
			return tools.WorkflowSummaryList(result), nil
		},
	}
}

func (s *HTTPServer) createCreateWorkflowTool() *MCPTool {
	return &MCPTool{
		Name: "create_workflow",
		Description: "Validate, plan, and apply an agentcloud.io/v1alpha1 Workflow resource after clarifying it with the user. " +
			"Follow the looper methodology: (1) workflow-worthiness gate — only create a workflow when fresh observations can change the next action across runs; recommend a one-time task otherwise; " +
			"(2) pick the smallest trigger in the Workflow spec; " +
			"(3) clarify goal, acceptance criteria and schedule with the user BEFORE calling this tool; " +
			"(4) workflows are created disabled by default — pass enabled=true only after the user explicitly confirms.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"resource": map[string]interface{}{
					"type":        "object",
					"description": "Complete Workflow resource manifest with apiVersion, kind, metadata, and spec.",
				},
				"enabled": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable immediately. Only set true after explicit user confirmation (default: false)",
				},
			},
			"required":             []string{"resource"},
			"additionalProperties": false,
		},
		Handler: func(ctx context.Context, client tools.CollaborationClient, args map[string]interface{}) (interface{}, error) {
			resource, err := resourceManifestArgument(args)
			if err != nil {
				return nil, err
			}
			req := &tools.WorkflowCreateRequest{
				Resource: resource,
				Enabled:  getBoolArg(args, "enabled"),
			}
			return client.CreateWorkflow(ctx, req)
		},
	}
}

func (s *HTTPServer) createTriggerWorkflowTool() *MCPTool {
	return &MCPTool{
		Name:        "trigger_workflow",
		Description: "Manually trigger a workflow run. Optionally pass runtime variables to override the workflow's default prompt variables.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"workflow_slug": map[string]interface{}{
					"type":        "string",
					"description": "The slug of the workflow to trigger. Use list_workflows to find available workflows.",
				},
				"variables": map[string]interface{}{
					"type":        "object",
					"description": "Runtime variables to override prompt template placeholders (optional)",
				},
			},
			"required": []string{"workflow_slug"},
		},
		Handler: func(ctx context.Context, client tools.CollaborationClient, args map[string]interface{}) (interface{}, error) {
			workflowSlug := getStringArg(args, "workflow_slug")
			if workflowSlug == "" {
				return nil, fmt.Errorf("workflow_slug is required")
			}
			variables := getMapArg(args, "variables")

			return client.TriggerWorkflow(ctx, workflowSlug, variables)
		},
	}
}
