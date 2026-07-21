package automation

import (
	agentDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agent"
	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
)

func init() { register("loopal", loopalAdapter{}) }

// loopalAdapter maps automation levels onto loopal's `permission_mode`
// (supervised / auto / bypass). Autonomous forces ACP for non-interactive runs.
type loopalAdapter struct{}

func (loopalAdapter) Apply(level string) Output {
	out := Output{ConfigOverrides: map[string]string{}}
	switch level {
	case podDomain.AutomationLevelInteractive:
		out.ConfigOverrides[agentDomain.ConfigKeyPermissionMode] = "supervised"
	case podDomain.AutomationLevelAutoEdit:
		out.ConfigOverrides[agentDomain.ConfigKeyPermissionMode] = "auto"
	case podDomain.AutomationLevelAutonomous:
		out.ConfigOverrides[agentDomain.ConfigKeyPermissionMode] = "bypass"
		out.InteractionMode = podDomain.InteractionModeACP
	}
	return out
}
