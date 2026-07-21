package acp

import "fmt"

// BuildMCPServersConfig creates the pod-scoped MCP configuration passed to session/new.
func BuildMCPServersConfig(mcpPort int, podKey string) map[string]any {
	return map[string]any{
		"agentcloud": map[string]any{
			"type": "http",
			"url":  fmt.Sprintf("http://127.0.0.1:%d/mcp", mcpPort),
			"headers": []map[string]string{{
				"name": "X-Pod-Key", "value": podKey,
			}},
		},
	}
}
