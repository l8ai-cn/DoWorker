package grpc

import (
	"context"
	"log/slog"

	"github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

type GRPCCommandSender struct {
	adapter *GRPCRunnerAdapter
}

func NewGRPCCommandSender(adapter *GRPCRunnerAdapter) *GRPCCommandSender {
	return &GRPCCommandSender{adapter: adapter}
}

func (s *GRPCCommandSender) SendCreatePod(ctx context.Context, runnerID int64, cmd *runnerv1.CreatePodCommand) error {
	slog.InfoContext(ctx, "sending create_pod command",
		"runner_id", runnerID,
		"pod_key", cmd.GetPodKey(),
		"prompt_set", cmd.GetPrompt() != "",
		"prompt_position", cmd.GetPromptPosition(),
	)
	if err := s.adapter.SendCreatePod(runnerID, cmd); err != nil {
		slog.ErrorContext(ctx, "failed to send create_pod command", "runner_id", runnerID, "pod_key", cmd.GetPodKey(), "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendTerminatePod(ctx context.Context, runnerID int64, podKey string, deleteBranch bool) error {
	slog.InfoContext(ctx, "sending terminate_pod command", "runner_id", runnerID, "pod_key", podKey, "delete_branch", deleteBranch)
	if err := s.adapter.SendTerminatePod(runnerID, podKey, false, deleteBranch); err != nil {
		slog.ErrorContext(ctx, "failed to send terminate_pod command", "runner_id", runnerID, "pod_key", podKey, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendPodInput(ctx context.Context, runnerID int64, podKey string, data []byte) error {
	if err := s.adapter.SendPodInput(runnerID, podKey, data); err != nil {
		slog.ErrorContext(ctx, "failed to send pod_input command", "runner_id", runnerID, "pod_key", podKey, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendServerMessage(ctx context.Context, runnerID int64, msg *runnerv1.ServerMessage) error {
	if err := s.adapter.SendServerMessage(runnerID, msg); err != nil {
		slog.ErrorContext(ctx, "failed to send server message", "runner_id", runnerID, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendPrompt(ctx context.Context, runnerID int64, podKey, prompt string) error {
	slog.InfoContext(ctx, "sending prompt command", "runner_id", runnerID, "pod_key", podKey)
	if err := s.adapter.SendPrompt(runnerID, podKey, prompt); err != nil {
		slog.ErrorContext(ctx, "failed to send prompt command", "runner_id", runnerID, "pod_key", podKey, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendSubscribePod(ctx context.Context, runnerID int64, podKey, relayURL, runnerToken string, includeSnapshot bool, snapshotHistory int32) error {
	slog.InfoContext(ctx, "sending subscribe_pod command", "runner_id", runnerID, "pod_key", podKey)
	if err := s.adapter.SendSubscribePod(ctx, runnerID, podKey, relayURL, runnerToken, includeSnapshot, snapshotHistory); err != nil {
		slog.ErrorContext(ctx, "failed to send subscribe_pod command", "runner_id", runnerID, "pod_key", podKey, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendUnsubscribePod(ctx context.Context, runnerID int64, podKey string) error {
	slog.InfoContext(ctx, "sending unsubscribe_pod command", "runner_id", runnerID, "pod_key", podKey)
	if err := s.adapter.SendUnsubscribePod(runnerID, podKey); err != nil {
		slog.ErrorContext(ctx, "failed to send unsubscribe_pod command", "runner_id", runnerID, "pod_key", podKey, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendCreateAutopilot(runnerID int64, cmd *runnerv1.CreateAutopilotCommand) error {
	slog.Info("sending create_autopilot command", "runner_id", runnerID, "autopilot_key", cmd.GetAutopilotKey())
	if err := s.adapter.SendCreateAutopilot(runnerID, cmd); err != nil {
		slog.Error("failed to send create_autopilot command", "runner_id", runnerID, "autopilot_key", cmd.GetAutopilotKey(), "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendAutopilotControl(runnerID int64, cmd *runnerv1.AutopilotControlCommand) error {
	slog.Info("sending autopilot_control command", "runner_id", runnerID, "autopilot_key", cmd.GetAutopilotKey())
	if err := s.adapter.SendAutopilotControl(runnerID, cmd); err != nil {
		slog.Error("failed to send autopilot_control command", "runner_id", runnerID, "autopilot_key", cmd.GetAutopilotKey(), "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendUpdatePodPerpetual(ctx context.Context, runnerID int64, podKey string, perpetual bool) error {
	slog.InfoContext(ctx, "sending update_pod_perpetual command", "runner_id", runnerID, "pod_key", podKey, "perpetual", perpetual)
	if err := s.adapter.SendUpdatePodPerpetual(runnerID, podKey, perpetual); err != nil {
		slog.ErrorContext(ctx, "failed to send update_pod_perpetual command", "runner_id", runnerID, "pod_key", podKey, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendUpdatePodPolicyRules(ctx context.Context, runnerID int64, podKey string, rules []*runnerv1.PolicyRuleSnapshot) error {
	if err := s.adapter.SendUpdatePodPolicyRules(runnerID, podKey, rules); err != nil {
		slog.ErrorContext(ctx, "failed to send update_pod_policy_rules", "runner_id", runnerID, "pod_key", podKey, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendAcpRelay(ctx context.Context, runnerID int64, podKey, payloadJSON string) error {
	if err := s.adapter.SendAcpRelay(runnerID, podKey, payloadJSON); err != nil {
		slog.ErrorContext(ctx, "failed to send acp relay", "runner_id", runnerID, "pod_key", podKey, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendConnectTunnel(ctx context.Context, runnerID int64, gatewayURL, tunnelToken string) error {
	slog.InfoContext(ctx, "sending connect_tunnel command", "runner_id", runnerID, "gateway_url", gatewayURL)
	if err := s.adapter.SendConnectTunnel(ctx, runnerID, gatewayURL, tunnelToken); err != nil {
		slog.ErrorContext(ctx, "failed to send connect_tunnel command", "runner_id", runnerID, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendQuerySandboxes(runnerID int64, requestID string, podKeys []string) error {
	slog.Info("sending query_sandboxes command", "runner_id", runnerID, "request_id", requestID, "pod_count", len(podKeys))
	if err := s.adapter.SendQuerySandboxes(runnerID, requestID, podKeys); err != nil {
		slog.Error("failed to send query_sandboxes command", "runner_id", runnerID, "request_id", requestID, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendSandboxFs(runnerID int64, cmd *runnerv1.SandboxFsCommand) error {
	if err := s.adapter.SendSandboxFs(runnerID, cmd); err != nil {
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendObservePod(ctx context.Context, runnerID int64, requestID, podKey string, lines int32, includeScreen bool) error {
	slog.InfoContext(ctx, "sending observe_pod command", "runner_id", runnerID, "request_id", requestID, "pod_key", podKey)
	if err := s.adapter.SendObservePod(runnerID, requestID, podKey, lines, includeScreen); err != nil {
		slog.ErrorContext(ctx, "failed to send observe_pod command", "runner_id", runnerID, "pod_key", podKey, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) IsConnected(runnerID int64) bool {
	return s.adapter.IsConnected(runnerID)
}

func (s *GRPCCommandSender) SendUpgradeRunner(runnerID int64, requestID, targetVersion string, force bool) error {
	slog.Info("sending upgrade_runner command", "runner_id", runnerID, "request_id", requestID, "target_version", targetVersion, "force", force)
	if err := s.adapter.SendUpgradeRunner(runnerID, requestID, targetVersion, force); err != nil {
		slog.Error("failed to send upgrade_runner command", "runner_id", runnerID, "request_id", requestID, "error", err)
		return err
	}
	return nil
}

func (s *GRPCCommandSender) SendUploadLogs(runnerID int64, requestID, presignedURL string, urlExpiresAt int64) error {
	slog.Info("sending upload_logs command", "runner_id", runnerID, "request_id", requestID)
	if err := s.adapter.SendUploadLogs(runnerID, requestID, presignedURL, urlExpiresAt); err != nil {
		slog.Error("failed to send upload_logs command", "runner_id", runnerID, "request_id", requestID, "error", err)
		return err
	}
	return nil
}

var _ runner.RunnerCommandSender = (*GRPCCommandSender)(nil)

var _ runner.SandboxQuerySender = (*GRPCCommandSender)(nil)

var _ runner.UpgradeCommandSender = (*GRPCCommandSender)(nil)

var _ runner.LogUploadCommandSender = (*GRPCCommandSender)(nil)

var _ runner.ServerMessageSender = (*GRPCCommandSender)(nil)
