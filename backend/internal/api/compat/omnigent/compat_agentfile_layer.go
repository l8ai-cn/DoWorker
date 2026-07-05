package omnigent

import "strings"

// compatAgentfileLayer builds an Agentfile layer for Omnigent compat sessions.
// ACP mode is required so runner emits AcpSessionEvent → EventBridge → items/SSE.
func compatAgentfileLayer(extra ...string) *string {
	parts := []string{"MODE acp"}
	for _, line := range extra {
		if t := strings.TrimSpace(line); t != "" {
			parts = append(parts, t)
		}
	}
	out := strings.Join(parts, "\n")
	return &out
}

func compatPTYLayer(extra ...string) *string {
	parts := []string{"MODE pty"}
	for _, line := range extra {
		if t := strings.TrimSpace(line); t != "" {
			parts = append(parts, t)
		}
	}
	out := strings.Join(parts, "\n")
	return &out
}
