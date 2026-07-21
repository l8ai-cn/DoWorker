// Package client provides gRPC connection management for Runner.
package client

import (
	"context"
	"time"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

// writeLoop sends messages to the gRPC stream with priority scheduling.
// Readiness results have highest priority, followed by control messages, reliable
// Workbench events, and terminal output.
// This is the ONLY goroutine that calls stream.Send() to ensure thread-safety.
// Includes stuck detection: triggers reconnect if no successful send for 30 seconds.
func (c *GRPCConnection) writeLoop(ctx context.Context, done <-chan struct{}) {
	log := logger.GRPC()
	log.InfoContext(ctx, "Write loop starting")
	defer log.InfoContext(ctx, "Write loop exited")

	stuckTicker := time.NewTicker(10 * time.Second)
	defer stuckTicker.Stop()

	// Initialize last send time
	c.lastSendTime.Store(time.Now().UnixNano())

	for {
		select {
		case <-c.stopCh:
			return
		case <-done:
			return
		case <-ctx.Done():
			return

		case <-stuckTicker.C:
			// Stuck detection: if no successful send for 2*heartbeatInterval, trigger reconnect.
			// Using 2x avoids false positives when heartbeat just happens to align with check.
			stuckThreshold := 2 * c.heartbeatInterval
			lastSend := time.Unix(0, c.lastSendTime.Load())
			if time.Since(lastSend) > stuckThreshold {
				log.Error("WriteLoop stuck, triggering reconnect",
					"threshold", stuckThreshold, "last_send_ago", time.Since(lastSend))
				c.triggerReconnect()
				return
			}

		case msg := <-c.readyCh:
			c.sendAndRecord(msg)

		default:
			// No readiness result pending - use nested select for priority
			select {
			case <-c.stopCh:
				return
			case <-done:
				return
			case <-ctx.Done():
				return
			case msg := <-c.readyCh:
				c.sendAndRecord(msg)
			case msg := <-c.controlCh:
				c.sendAndRecord(msg)
			case msg := <-c.workbenchCh:
				c.sendAndRecord(msg)
			case msg := <-c.terminalCh:
				// Process terminal messages when no control messages pending
				logger.GRPCTrace().Trace("writeLoop: sending terminal message",
					"queue_len", len(c.terminalCh))
				c.sendAndRecord(msg)
			}
		}
	}
}

// sendAndRecord sends a message with a hard timeout to prevent writeLoop from blocking forever.
// If stream.Send() doesn't complete within sendTimeout, the message is abandoned and reconnect
// is triggered. The orphaned goroutine will exit when the stream is closed during reconnection.
//
// Key insight: gRPC stream.Send() can block indefinitely due to flow control.
// We cannot cancel it, but closing the stream (during reconnection) unblocks it.
//
// Goroutine leak prevention: after timeout, we immediately close the stream reference.
// This causes the blocked stream.Send() to return an error, allowing the goroutine to exit.
// Without this, the goroutine stays alive until the next reconnection cycle clears the stream.
func (c *GRPCConnection) sendAndRecord(msg *runnerv1.RunnerMessage) {
	c.mu.Lock()
	stream := c.stream
	c.mu.Unlock()

	if stream == nil {
		logger.GRPC().Warn("sendAndRecord: stream is nil, dropping message")
		return
	}

	const sendTimeout = 5 * time.Second

	type sendResult struct {
		err     error
		elapsed time.Duration
	}

	resultCh := make(chan sendResult, 1)
	start := time.Now()

	go func() {
		err := stream.Send(msg)
		resultCh <- sendResult{err: err, elapsed: time.Since(start)}
	}()

	select {
	case result := <-resultCh:
		// Send completed (success or failure)
		if result.err != nil {
			logger.GRPC().Error("Failed to send message", "error", result.err, "elapsed", result.elapsed)
			return
		}

		// Log slow sends for diagnosis
		if result.elapsed > 100*time.Millisecond {
			logger.GRPC().Warn("Slow stream.Send()", "elapsed", result.elapsed,
				"terminal_queue", len(c.terminalCh))
		}

		// Update last successful send time
		c.lastSendTime.Store(time.Now().UnixNano())

	case <-time.After(sendTimeout):
		// Send timed out — the goroutine is blocked on stream.Send().
		// Clear the stream reference and trigger reconnect. This causes:
		// 1. writeLoop to stop sending new messages (stream == nil check)
		// 2. Reconnection flow to close the gRPC conn, which unblocks stream.Send()
		// 3. The orphaned goroutine receives an error from Send() and exits
		logger.GRPC().Error("stream.Send() timed out, clearing stream and triggering reconnect",
			"timeout", sendTimeout, "terminal_queue", len(c.terminalCh))

		c.mu.Lock()
		c.stream = nil
		c.mu.Unlock()

		// Trigger reconnect to recover from degraded connection
		c.triggerReconnect()
	}
}
