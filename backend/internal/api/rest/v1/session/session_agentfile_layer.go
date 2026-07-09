package sessionapi

import "strings"

func acpAgentfileLayer(extra ...string) *string {
	parts := []string{"MODE acp"}
	for _, line := range extra {
		if t := strings.TrimSpace(line); t != "" {
			parts = append(parts, t)
		}
	}
	out := strings.Join(parts, "\n")
	return &out
}

func sessionAgentfileLayer(agentID string, ptyOnly bool, extra ...string) *string {
	var parts []string
	if ptyOnly {
		parts = append(parts, "MODE pty")
		if agentID == "codex-cli" {
			parts = append(parts, `CONFIG approval_mode = "never"`)
		}
	} else {
		parts = append(parts, "MODE acp")
	}
	for _, line := range extra {
		if t := strings.TrimSpace(line); t != "" {
			parts = append(parts, t)
		}
	}
	out := strings.Join(parts, "\n")
	return &out
}
