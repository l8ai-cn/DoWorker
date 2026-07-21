package runner

import (
	"context"
	"time"

	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

// buildACPPod creates a pod configured for ACP (Agent Communication Protocol) interaction.
// The ACPClient is NOT created here — it will be created by wireAndStartACPPod()
// in the MessageHandler, which wires Relay-forwarding event callbacks.
func (b *PodBuilder) buildACPPod(
	_ context.Context,
	sandboxRoot, workingDir, branchName string,
	resolvedArgs []string,
	envVars map[string]string,
	launchCommand string,
	workspace *sandboxWorkspace,
) (*Pod, error) {
	b.sendProgress("starting_acp", 80, "Preparing ACP agent...")
	unsetEnv := gitProcessIsolationUnsetEnv(b.cmd.GetSandboxConfig().GetCredentialType())

	pod := &Pod{
		ID:              b.cmd.PodKey,
		PodKey:          b.cmd.PodKey,
		Agent:           launchCommand,
		InteractionMode: InteractionModeACP,
		Branch:          branchName,
		SandboxPath:     sandboxRoot,
		LaunchCommand:   launchCommand,
		LaunchArgs:      resolvedArgs,
		WorkDir:         workingDir,
		LaunchEnv:       buildMergedEnvWithout(envVars, unsetEnv),
		Perpetual:       b.cmd.Perpetual,
		PolicyRules:     policyRulesFromProto(b.cmd.PolicyRules),
		StartedAt:       time.Now(),
		Status:          PodStatusInitializing,
		workspace:       workspace,
		// ACPClient, IO are set by wireAndStartACPPod()
		// PTY fields (Terminal, VirtualTerminal, Aggregator, PTYLogger) left nil
	}

	logger.Pod().Info("Pod built (ACP)", "pod_key", b.cmd.PodKey, "working_dir", workingDir)
	b.sendProgress("acp_ready", 100, "ACP agent ready")

	return pod, nil
}
