package runner

import (
	"github.com/anthropics/agentsmesh/runner/internal/acp"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

func (h *RunnerMessageHandler) abortACPPodStartup(
	podKey string,
	acpClient *acp.ACPClient,
	sandboxPath string,
) {
	if pod := h.podStore.Delete(podKey); pod != nil {
		pod.closeWorkspace()
	}
	if acpClient != nil {
		acpClient.Stop()
	}
	if sandboxPath != "" {
		h.removePodSandbox(sandboxPath)
	}
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
