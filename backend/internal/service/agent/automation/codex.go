package automation

import podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"

func init() {
	register("codex-cli", codexAdapter{})
	register("video-studio", codexAdapter{})
}

// codexApprovalModeKey is codex-cli's native approval CONFIG. Options are
// untrusted / on-request / never (see the builtin codex AgentFile).
const codexApprovalModeKey = "approval_mode"

// codexAdapter maps automation levels onto codex-cli's `approval_mode`.
// Autonomous forces ACP; the runner already pins approvalPolicy=never for the
// codex ACP transport, so full-auto behavior is preserved end-to-end.
type codexAdapter struct{}

func (codexAdapter) Apply(level string) Output {
	out := Output{ConfigOverrides: map[string]string{}}
	switch level {
	case podDomain.AutomationLevelInteractive:
		out.ConfigOverrides[codexApprovalModeKey] = "untrusted"
	case podDomain.AutomationLevelAutoEdit:
		out.ConfigOverrides[codexApprovalModeKey] = "on-request"
	case podDomain.AutomationLevelAutonomous:
		out.ConfigOverrides[codexApprovalModeKey] = "never"
		out.InteractionMode = podDomain.InteractionModeACP
	}
	return out
}
