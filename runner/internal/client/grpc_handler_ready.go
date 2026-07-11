package client

import (
	"errors"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

var errReadyHandlerUnavailable = errors.New("runner message handler is not configured")

func (c *GRPCConnection) handleSubscribePod(cmd *runnerv1.SubscribePodCommand) {
	log := logger.GRPC()
	log.Info("Received subscribe_pod", "pod_key", cmd.GetPodKey(), "relay_url", cmd.GetRelayUrl())

	err := c.subscribePod(cmd)
	result := &runnerv1.RelaySubscriptionResultEvent{
		CommandId: cmd.GetCommandId(),
		Success:   err == nil,
	}
	if err != nil {
		result.ErrorCode = readyErrorCode(err, "relay_subscription_failed")
		result.Message = err.Error()
		log.Error("Failed to subscribe pod", "pod_key", cmd.GetPodKey(), "error", err)
	}
	if sendErr := c.sendReadyResult(&runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_RelaySubscriptionResult{
			RelaySubscriptionResult: result,
		},
	}); sendErr != nil {
		log.Error("Failed to report relay subscription result",
			"command_id", cmd.GetCommandId(), "error", sendErr)
	}
}

func (c *GRPCConnection) subscribePod(cmd *runnerv1.SubscribePodCommand) error {
	if c.handler == nil {
		return errReadyHandlerUnavailable
	}
	return c.handler.OnSubscribePod(SubscribePodRequest{
		CommandID:       cmd.GetCommandId(),
		PodKey:          cmd.GetPodKey(),
		RelayURL:        cmd.GetRelayUrl(),
		RunnerToken:     cmd.GetRunnerToken(),
		IncludeSnapshot: cmd.GetIncludeSnapshot(),
		SnapshotHistory: cmd.GetSnapshotHistory(),
	})
}

func (c *GRPCConnection) handleConnectTunnel(cmd *runnerv1.ConnectTunnelCommand) {
	log := logger.GRPC()
	log.Info("Received connect_tunnel", "gateway_url", cmd.GetGatewayUrl())

	err := c.connectTunnel(cmd)
	result := &runnerv1.TunnelConnectionResultEvent{
		CommandId: cmd.GetCommandId(),
		Success:   err == nil,
	}
	if err != nil {
		result.ErrorCode = readyErrorCode(err, "tunnel_connection_failed")
		result.Message = err.Error()
		log.Error("Failed to connect tunnel", "gateway_url", cmd.GetGatewayUrl(), "error", err)
	}
	if sendErr := c.sendReadyResult(&runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_TunnelConnectionResult{
			TunnelConnectionResult: result,
		},
	}); sendErr != nil {
		log.Error("Failed to report tunnel connection result",
			"command_id", cmd.GetCommandId(), "error", sendErr)
	}
}

func (c *GRPCConnection) connectTunnel(cmd *runnerv1.ConnectTunnelCommand) error {
	if c.handler == nil {
		return errReadyHandlerUnavailable
	}
	return c.handler.OnConnectTunnel(ConnectTunnelRequest{
		CommandID:   cmd.GetCommandId(),
		GatewayURL:  cmd.GetGatewayUrl(),
		TunnelToken: cmd.GetTunnelToken(),
	})
}

func readyErrorCode(err error, operationCode string) string {
	if errors.Is(err, errReadyHandlerUnavailable) {
		return "handler_unavailable"
	}
	return operationCode
}
