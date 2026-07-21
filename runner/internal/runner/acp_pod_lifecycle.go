package runner

import (
	"fmt"

	"github.com/l8ai-cn/agentcloud/runner/internal/acp"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

func (h *RunnerMessageHandler) abortACPPodStartup(
	podKey string,
	acpClient *acp.ACPClient,
	sandboxPath string,
) error {
	if pod := h.podStore.Delete(podKey); pod != nil {
		pod.closeWorkspace()
	}
	if acpClient != nil {
		acpClient.Stop()
	}
	if sandboxPath != "" {
		if err := h.removePodSandbox(podKey, sandboxPath); err != nil {
			return fmt.Errorf("remove failed ACP sandbox: %w", err)
		}
	}
	return nil
}

func (h *RunnerMessageHandler) handleACPExit(podKey string, exitCode int) {
	if pod, ok := h.podStore.Get(podKey); ok && pod != nil {
		if acpIO, ok := pod.IO.(*ACPPodIO); ok {
			acpIO.ForceIdleIfBusy()
		}
	}
	logger.Pod().Info(
		"ACP process exited",
		"pod_key",
		podKey,
		"exit_code",
		exitCode,
	)
	h.cleanupPodExit(podKey, exitCode, false)
}
