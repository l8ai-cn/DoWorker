package runner

import (
	"context"
	"fmt"
	"os"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/client"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

func (h *RunnerMessageHandler) OnTerminatePod(req client.TerminatePodRequest) error {
	log := logger.Pod()
	log.Info("Terminating pod", "pod_key", req.PodKey, "delete_branch", req.DeleteBranch)

	pod, ok := h.podStore.Get(req.PodKey)
	if !ok {
		log.Warn("Pod not found for termination", "pod_key", req.PodKey)
		return fmt.Errorf("pod not found: %s", req.PodKey)
	}
	if req.DeleteBranch && pod.SandboxPath != "" {
		h.removePodSandbox(pod.SandboxPath)
	}

	h.cleanupPodExit(req.PodKey, -1, true)
	return nil
}

func (h *RunnerMessageHandler) removePodSandbox(path string) {
	if runner, ok := h.runner.(*Runner); ok && runner.workspace != nil {
		if err := runner.workspace.RemoveWorktree(context.Background(), path); err != nil {
			logger.Pod().Warn("worktree remove failed", "path", path, "error", err)
		}
		return
	}
	_ = os.RemoveAll(path)
}

func (h *RunnerMessageHandler) OnUpdatePodPerpetual(
	cmd *runnerv1.UpdatePodPerpetualCommand,
) error {
	log := logger.Pod()
	pod, ok := h.podStore.Get(cmd.PodKey)
	if !ok {
		log.Warn("Pod not found for perpetual update", "pod_key", cmd.PodKey)
		return fmt.Errorf("pod not found: %s", cmd.PodKey)
	}
	pod.Perpetual = cmd.Perpetual
	h.podStore.Put(cmd.PodKey, pod)
	log.Info("Pod perpetual mode updated", "pod_key", cmd.PodKey, "perpetual", cmd.Perpetual)
	return nil
}

func (h *RunnerMessageHandler) OnUpdatePodPolicyRules(
	cmd *runnerv1.UpdatePodPolicyRulesCommand,
) error {
	log := logger.Pod()
	podKey := cmd.GetPodKey()
	pod, ok := h.podStore.Get(podKey)
	if !ok {
		log.Warn("Pod not found for policy update", "pod_key", podKey)
		return fmt.Errorf("pod not found: %s", podKey)
	}
	pod.PolicyRules = policyRulesFromProto(cmd.GetPolicyRules())
	h.podStore.Put(podKey, pod)
	log.Info("Pod policy rules updated", "pod_key", podKey, "rules", len(pod.PolicyRules))
	return nil
}

func (h *RunnerMessageHandler) OnListPods() []client.PodInfo {
	pods := h.podStore.All()
	result := make([]client.PodInfo, 0, len(pods))
	for _, pod := range pods {
		info := client.PodInfo{
			PodKey:      pod.PodKey,
			Status:      pod.GetStatus(),
			AgentStatus: h.getAgentStatusFromDetector(pod),
		}
		if pod.IO != nil {
			info.Pid = pod.IO.GetPID()
		}
		result = append(result, info)
	}
	return result
}

func (h *RunnerMessageHandler) getAgentStatusFromDetector(pod *Pod) string {
	if pod.IO != nil {
		return pod.IO.GetAgentStatus()
	}
	return "idle"
}
