package mcp

import (
	"context"
	"fmt"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func (s *HTTPServer) createCreatePodTool() *MCPTool {
	return &MCPTool{
		Name:        "create_pod",
		Description: "Consume a previously validated Worker plan to create a pod. Runtime, model, prompt, tool, knowledge, permission, repository, and placement settings come from the immutable plan. The new pod is automatically bound to the creator for pod interaction.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"plan_id": map[string]interface{}{
					"type":        "string",
					"description": "Canonical UUID returned by the Worker plan operation.",
				},
			},
			"required":             []string{"plan_id"},
			"additionalProperties": false,
		},
		Handler: func(ctx context.Context, client tools.CollaborationClient, args map[string]interface{}) (interface{}, error) {
			rawPlanID, ok := args["plan_id"]
			if !ok {
				return nil, fmt.Errorf("plan_id is required")
			}
			if len(args) != 1 {
				return nil, fmt.Errorf("create_pod accepts only plan_id")
			}
			planID, ok := rawPlanID.(string)
			if !ok || planID == "" {
				return nil, fmt.Errorf("plan_id is required")
			}

			resp, err := client.CreatePod(ctx, &tools.PodCreateRequest{PlanID: planID})
			if err != nil {
				return nil, err
			}

			scopes := []tools.BindingScope{tools.ScopePodRead, tools.ScopePodWrite}
			binding, err := client.RequestBinding(ctx, resp.PodKey, scopes)
			if err != nil {
				return nil, fmt.Errorf("pod %s was created but automatic binding failed", resp.PodKey)
			}
			binding, err := client.RequestBinding(
				ctx,
				resp.PodKey,
				[]tools.BindingScope{
					tools.ScopePodRead,
					tools.ScopePodWrite,
				},
			)
			if err != nil {
				return nil, fmt.Errorf("bind created pod %s: %w", resp.PodKey, err)
			}
			return fmt.Sprintf(
				"Pod: %s | Status: %s | Binding: #%d (%s)%s",
				resp.PodKey,
				resp.Status,
				binding.ID,
				binding.Status,
				appliedResourceSuffix(resp.Resource),
			), nil
		},
	}
}

func appliedResourceSuffix(resource *tools.AppliedResourceSummary) string {
	text := resource.FormatText()
	if text == "" {
		return ""
	}
	return " | " + text
}
