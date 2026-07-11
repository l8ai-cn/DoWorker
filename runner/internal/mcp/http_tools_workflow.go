package mcp

import (
	"context"
	"fmt"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
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
		Description: "Create a new automated workflow (repeatable AI agent task) after clarifying it with the user. " +
			"Follow the looper methodology: (1) workflow-worthiness gate — only create a workflow when fresh observations can change the next action across runs; recommend a one-time task otherwise; " +
			"(2) pick the smallest trigger — omit cron_expression for on-demand workflows, set it only when work truly arrives on a schedule; " +
			"(3) clarify goal, acceptance criteria and schedule with the user BEFORE calling this tool; " +
			"(4) workflows are created disabled by default — pass enabled=true only after the user explicitly confirms. " +
			"The prompt_template should state the goal, concrete acceptance criteria, and a clean idle exit (what to do when there is no work).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Human-readable workflow name (slug is derived from it)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "One-sentence purpose of the workflow",
				},
				"prompt_template": map[string]interface{}{
					"type":        "string",
					"description": "Instructions executed on every run: goal, acceptance criteria, verification steps, and idle exit behaviour",
				},
				"agent_slug": map[string]interface{}{
					"type":        "string",
					"description": "Agent image to run (defaults to the current pod's agent)",
				},
				"cron_expression": map[string]interface{}{
					"type":        "string",
					"description": "Standard 5-field cron (minute hour day month weekday). Omit for on-demand workflows",
				},
				"execution_mode": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"autopilot", "direct"},
					"description": "autopilot iterates until done (default); direct runs the prompt once per trigger",
				},
				"sandbox_strategy": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"persistent", "fresh"},
					"description": "persistent keeps the workspace between runs (default); fresh starts clean",
				},
				"timeout_minutes": map[string]interface{}{
					"type":        "integer",
					"description": "Per-run budget in minutes (default: 60)",
				},
				"session_persistence": map[string]interface{}{
					"type":        "boolean",
					"description": "Keep agent conversation history across runs",
				},
				"enabled": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable immediately. Only set true after explicit user confirmation (default: false)",
				},
			},
			"required": []string{"name", "prompt_template"},
		},
		Handler: func(ctx context.Context, client tools.CollaborationClient, args map[string]interface{}) (interface{}, error) {
			req := &tools.WorkflowCreateRequest{
				Name:               getStringArg(args, "name"),
				Description:        getStringArg(args, "description"),
				PromptTemplate:     getStringArg(args, "prompt_template"),
				AgentSlug:          getStringArg(args, "agent_slug"),
				CronExpression:     getStringArg(args, "cron_expression"),
				ExecutionMode:      getStringArg(args, "execution_mode"),
				SandboxStrategy:    getStringArg(args, "sandbox_strategy"),
				TimeoutMinutes:     getIntArg(args, "timeout_minutes"),
				SessionPersistence: getBoolArg(args, "session_persistence"),
				Enabled:            getBoolArg(args, "enabled"),
			}
			if req.Name == "" || req.PromptTemplate == "" {
				return nil, fmt.Errorf("name and prompt_template are required")
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
