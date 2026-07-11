package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

const (
	runnerReadyProtocolVersion = 3
	runnerReadyResultTimeout   = 25 * time.Second
)

func (a *GRPCRunnerAdapter) SendSubscribePod(
	ctx context.Context,
	runnerID int64,
	podKey, relayURL, runnerToken string,
	includeSnapshot bool,
	snapshotHistory int32,
) error {
	return a.sendAndWaitReady(ctx, runnerID, relaySubscriptionReady, func(commandID string) *runnerv1.ServerMessage {
		return &runnerv1.ServerMessage{
			Payload: &runnerv1.ServerMessage_SubscribePod{
				SubscribePod: &runnerv1.SubscribePodCommand{
					CommandId:       commandID,
					PodKey:          podKey,
					RelayUrl:        relayURL,
					RunnerToken:     runnerToken,
					IncludeSnapshot: includeSnapshot,
					SnapshotHistory: snapshotHistory,
				},
			},
		}
	})
}

func (a *GRPCRunnerAdapter) SendConnectTunnel(
	ctx context.Context,
	runnerID int64,
	gatewayURL, tunnelToken string,
) error {
	return a.sendAndWaitReady(ctx, runnerID, tunnelConnectionReady, func(commandID string) *runnerv1.ServerMessage {
		return &runnerv1.ServerMessage{
			Payload: &runnerv1.ServerMessage_ConnectTunnel{
				ConnectTunnel: &runnerv1.ConnectTunnelCommand{
					CommandId:   commandID,
					GatewayUrl:  gatewayURL,
					TunnelToken: tunnelToken,
				},
			},
		}
	})
}

func (a *GRPCRunnerAdapter) sendAndWaitReady(
	ctx context.Context,
	runnerID int64,
	kind runnerReadyResultKind,
	buildMessage func(commandID string) *runnerv1.ServerMessage,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	conn := a.connManager.GetConnection(runnerID)
	if conn == nil {
		return status.Errorf(codes.NotFound, "runner %d not connected", runnerID)
	}
	if conn.GetProtocolVersion() < runnerReadyProtocolVersion {
		return status.Errorf(
			codes.FailedPrecondition,
			"runner %d protocol version %d does not support readiness confirmation; upgrade to protocol %d",
			runnerID,
			conn.GetProtocolVersion(),
			runnerReadyProtocolVersion,
		)
	}

	commandID := uuid.NewString()
	result, unregister := a.readyResults.register(runnerID, conn.Generation, commandID, kind)
	defer unregister()

	msg := buildMessage(commandID)
	msg.Timestamp = time.Now().UnixMilli()
	if err := conn.SendMessage(msg); err != nil {
		return err
	}

	waitCtx, cancel := context.WithTimeout(ctx, runnerReadyResultTimeout)
	defer cancel()

	select {
	case ready := <-result:
		if ready.success {
			return nil
		}
		return fmt.Errorf("runner readiness failed: %s: %s", ready.errorCode, ready.message)
	case <-conn.CloseChan():
		return status.Error(codes.Unavailable, "runner disconnected before readiness confirmation")
	case <-waitCtx.Done():
		return fmt.Errorf("runner readiness confirmation: %w", waitCtx.Err())
	}
}
