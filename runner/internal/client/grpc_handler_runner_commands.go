package client

import (
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

func (c *GRPCConnection) handleInitializeResult(result *runnerv1.InitializeResult) {
	logger.GRPC().Debug("Received initialize_result", "version", result.ServerInfo.Version)
	select {
	case c.initResultCh <- result:
	default:
		logger.GRPC().Warn("Initialize result channel full, dropping")
	}
}

func (c *GRPCConnection) handleCreatePod(cmd *runnerv1.CreatePodCommand) {
	log := logger.GRPC()
	log.Info("Received create_pod", "pod_key", cmd.PodKey)
	if c.handler == nil {
		log.Warn("No handler set, ignoring create_pod")
		return
	}

	if err := c.handler.OnCreatePod(cmd); err != nil {
		log.Error("Failed to create pod", "pod_key", cmd.PodKey, "error", err)
		c.sendError(cmd.PodKey, "create_pod_failed", err.Error())
	}
}

func (c *GRPCConnection) handleTerminatePod(cmd *runnerv1.TerminatePodCommand) {
	log := logger.GRPC()
	log.Info("Received terminate_pod", "pod_key", cmd.PodKey, "force", cmd.Force)
	if c.handler == nil {
		log.Warn("No handler set, ignoring terminate_pod")
		return
	}

	req := TerminatePodRequest{PodKey: cmd.PodKey, DeleteBranch: cmd.GetDeleteBranch()}
	if err := c.handler.OnTerminatePod(req); err != nil {
		log.Error("Failed to terminate pod", "pod_key", cmd.PodKey, "error", err)
	}
}

func (c *GRPCConnection) handleUpgradeRunner(cmd *runnerv1.UpgradeRunnerCommand) {
	log := logger.GRPC()
	log.Info("Received upgrade_runner", "request_id", cmd.RequestId, "target_version", cmd.TargetVersion)
	if c.handler == nil {
		log.Warn("No handler set, ignoring upgrade_runner")
		return
	}

	if err := c.handler.OnUpgradeRunner(cmd); err != nil {
		log.Error("Failed to handle upgrade runner", "request_id", cmd.RequestId, "error", err)
	}
}

func (c *GRPCConnection) handleUpdatePodPerpetual(cmd *runnerv1.UpdatePodPerpetualCommand) {
	log := logger.GRPC()
	log.Info("Received update_pod_perpetual", "pod_key", cmd.PodKey, "perpetual", cmd.Perpetual)
	if c.handler == nil {
		log.Warn("No handler set, ignoring update_pod_perpetual")
		return
	}
	if err := c.handler.OnUpdatePodPerpetual(cmd); err != nil {
		log.Error("Failed to update pod perpetual", "pod_key", cmd.PodKey, "error", err)
	}
}

func (c *GRPCConnection) handleUpdatePodPolicyRules(cmd *runnerv1.UpdatePodPolicyRulesCommand) {
	log := logger.GRPC()
	if c.handler == nil {
		log.Warn("No handler set, ignoring update_pod_policy_rules")
		return
	}
	if err := c.handler.OnUpdatePodPolicyRules(cmd); err != nil {
		log.Error("Failed to update pod policy rules", "pod_key", cmd.GetPodKey(), "error", err)
	}
}

func (c *GRPCConnection) handleAcpRelay(cmd *runnerv1.AcpRelayCommand) {
	log := logger.GRPC()
	if c.handler == nil {
		log.Warn("No handler set, ignoring acp_relay")
		return
	}
	if err := c.handler.OnAcpRelay(cmd); err != nil {
		log.Error("Failed to handle acp relay", "pod_key", cmd.PodKey, "error", err)
	}
}
