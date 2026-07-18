package mcp

import (
	"context"
	"fmt"

	"github.com/anthropics/agentsmesh/runner/internal/mcp/tools"
)

func (s *HTTPServer) createCreatePodTool() *MCPTool {
	return &MCPTool{
		Name:        "create_pod",
		Description: "Validate, plan, and apply an agentsmesh.io/v1alpha1 Worker resource. The new pod is bound to the creator with pod:read and pod:write permissions.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"resource": map[string]interface{}{
					"type":        "object",
					"description": "Complete Worker resource manifest with apiVersion, kind, metadata, and spec.",
				},
			},
			"required":             []string{"resource"},
			"additionalProperties": false,
		},
		Handler: func(
			ctx context.Context,
			client tools.CollaborationClient,
			args map[string]interface{},
		) (interface{}, error) {
			resource, err := resourceManifestArgument(args)
			if err != nil {
				return nil, err
			}
			resp, err := client.CreatePod(ctx, &tools.PodCreateRequest{
				Resource: resource,
			})
			if err != nil {
				return nil, err
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
