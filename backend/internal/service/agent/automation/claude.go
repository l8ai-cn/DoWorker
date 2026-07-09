package automation

import (
	agentDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
)

func init() { register("claude-code", claudeAdapter{}) }

// claudeAdapter maps automation levels onto Claude Code's `permission_mode`
// CONFIG (default / acceptEdits / bypassPermissions). Autonomous forces ACP so
// tool approvals can be granted programmatically.
type claudeAdapter struct{}

func (claudeAdapter) Apply(level string) Output {
	out := Output{ConfigOverrides: map[string]string{}}
	switch level {
	case podDomain.AutomationLevelInteractive:
		out.ConfigOverrides[agentDomain.ConfigKeyPermissionMode] = podDomain.PermissionModeDefault
	case podDomain.AutomationLevelAutoEdit:
		out.ConfigOverrides[agentDomain.ConfigKeyPermissionMode] = podDomain.PermissionModeAcceptEdits
	case podDomain.AutomationLevelAutonomous:
		out.ConfigOverrides[agentDomain.ConfigKeyPermissionMode] = podDomain.PermissionModeBypass
		out.InteractionMode = podDomain.InteractionModeACP
	}
	return out
}
