// Package client provides gRPC connection management for Runner.
package client

import (
	"context"
	"fmt"
	"io"
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
	"github.com/anthropics/agentsmesh/runner/internal/safego"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// readLoop reads messages from the gRPC stream.
// The done channel is closed when the loop exits to notify other goroutines.
func (c *GRPCConnection) readLoop(ctx context.Context, done chan<- struct{}) {
	defer close(done) // Signal exit to other goroutines
	log := logger.GRPC()
	log.InfoContext(ctx, "Read loop starting")
	for {
		msg, err := c.stream.Recv()
		if err != nil {
			// Don't update lastRecvTime on error — only track successful receives
			if err == io.EOF {
				log.Info("Stream ended (EOF)")
				return
			}
			if status.Code(err) == codes.Canceled {
				logger.GRPCTrace().Trace("Stream cancelled")
			} else if fatal, hint := isFatalStreamError(err); fatal {
				log.Error("Fatal stream error (will not retry)", "error", err)
				log.Error(hint)
				c.setFatalError(fmt.Errorf("%s", hint))
			} else {
				log.Error("Stream error", "error", err)
			}
			return
		}
		// Record successful recv for liveness tracking and diagnostics
		c.lastRecvTime.Store(time.Now().UnixNano())
		c.handleServerMessage(ctx, msg)
	}
}

// handleServerMessage dispatches received server messages to appropriate handlers.
// Heavy operations (CreatePod, SubscribePod, CreateAutopilot) are dispatched
// asynchronously via goroutines to avoid blocking the readLoop.
// Lightweight operations remain synchronous to preserve message ordering.
func (c *GRPCConnection) handleServerMessage(ctx context.Context, msg *runnerv1.ServerMessage) {
	msgType := extractServerMessageType(msg)
	if !isHighFrequencyServerMessage(msgType) {
		var span trace.Span
		ctx, span = startMessageSpan(ctx, msgType)
		defer span.End()
	}
	_ = ctx
	switch payload := msg.Payload.(type) {
	case *runnerv1.ServerMessage_InitializeResult:
		c.handleInitializeResult(payload.InitializeResult)

	// Heavy operations - dispatched via per-pod command queue.
	// Same pod's commands execute sequentially (create_pod before create_autopilot).
	// Different pods execute concurrently. Tracked by handlerWg for clean shutdown.
	case *runnerv1.ServerMessage_CreatePod:
		c.handlerWg.Add(1)
		c.podQueue.Enqueue(payload.CreatePod.PodKey, func() {
			defer c.handlerWg.Done()
			c.handleCreatePod(payload.CreatePod)
		})

	case *runnerv1.ServerMessage_TerminatePod:
		c.handlerWg.Add(1)
		c.podQueue.Enqueue(payload.TerminatePod.PodKey, func() {
			defer c.handlerWg.Done()
			c.handleTerminatePod(payload.TerminatePod)
			c.podQueue.Remove(payload.TerminatePod.PodKey)
		})

	case *runnerv1.ServerMessage_SubscribePod:
		c.handlerWg.Add(1)
		go func() {
			defer c.handlerWg.Done()
			c.handleSubscribePod(payload.SubscribePod)
		}()

	case *runnerv1.ServerMessage_CreateAutopilot:
		c.handlerWg.Add(1)
		c.podQueue.Enqueue(payload.CreateAutopilot.PodKey, func() {
			defer c.handlerWg.Done()
			c.handleCreateAutopilot(payload.CreateAutopilot)
		})

	case *runnerv1.ServerMessage_RunVerification:
		c.handlerWg.Add(1)
		c.podQueue.Enqueue(payload.RunVerification.PodKey, func() {
			defer c.handlerWg.Done()
			c.handleRunVerification(payload.RunVerification)
		})

	// Lightweight operations - synchronous to preserve ordering
	case *runnerv1.ServerMessage_PodInput:
		c.handlePodInput(payload.PodInput)

	case *runnerv1.ServerMessage_SendPrompt:
		c.handlerWg.Add(1)
		c.podQueue.Enqueue(payload.SendPrompt.PodKey, func() {
			defer c.handlerWg.Done()
			c.handleSendPrompt(payload.SendPrompt)
		})

	case *runnerv1.ServerMessage_UnsubscribePod:
		c.handleUnsubscribePod(payload.UnsubscribePod)

	case *runnerv1.ServerMessage_QuerySandboxes:
		c.handleQuerySandboxes(payload.QuerySandboxes)

	case *runnerv1.ServerMessage_ObservePod:
		c.handleObservePod(payload.ObservePod)

	case *runnerv1.ServerMessage_AutopilotControl:
		c.handleAutopilotControl(payload.AutopilotControl)

	case *runnerv1.ServerMessage_McpResponse:
		c.handleMcpResponse(payload.McpResponse)

	case *runnerv1.ServerMessage_Ping:
		c.handlePing(payload.Ping)

	case *runnerv1.ServerMessage_HeartbeatAck:
		c.handleHeartbeatAck(payload.HeartbeatAck)

	case *runnerv1.ServerMessage_UpgradeRunner:
		c.handlerWg.Add(1)
		safego.Go("handle-upgrade-runner", func() {
			defer c.handlerWg.Done()
			c.handleUpgradeRunner(payload.UpgradeRunner)
		})

	case *runnerv1.ServerMessage_UploadLogs:
		c.handlerWg.Add(1)
		safego.Go("handle-upload-logs", func() {
			defer c.handlerWg.Done()
			c.handleUploadLogs(payload.UploadLogs)
		})

	case *runnerv1.ServerMessage_UpdatePodPerpetual:
		c.handleUpdatePodPerpetual(payload.UpdatePodPerpetual)

	case *runnerv1.ServerMessage_UpdatePodPolicyRules:
		c.handleUpdatePodPolicyRules(payload.UpdatePodPolicyRules)

	case *runnerv1.ServerMessage_AcpRelay:
		c.handleAcpRelay(payload.AcpRelay)

	case *runnerv1.ServerMessage_SandboxFs:
		c.handlerWg.Add(1)
		c.podQueue.Enqueue(payload.SandboxFs.PodKey, func() {
			defer c.handlerWg.Done()
			c.handleSandboxFs(payload.SandboxFs)
		})

	case *runnerv1.ServerMessage_ConnectTunnel:
		c.handleConnectTunnel(payload.ConnectTunnel)

	default:
		logger.GRPC().Warn("Unknown server message type")
	}
}
