package grpc

import (
	"context"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func (a *GRPCRunnerAdapter) persistTunnelConnection(
	ctx context.Context,
	runnerID int64,
	result *runnerv1.TunnelConnectionResultEvent,
) error {
	if a.runnerService == nil {
		return nil
	}
	if err := a.runnerService.UpdateTunnelConnection(
		ctx,
		runnerID,
		result.GetSuccess(),
		result.GetErrorCode(),
	); err != nil {
		a.logger.Error("failed to persist tunnel connection state",
			"runner_id", runnerID,
			"success", result.GetSuccess(),
			"error_code", result.GetErrorCode(),
			"error", err,
		)
		return err
	}
	return nil
}
