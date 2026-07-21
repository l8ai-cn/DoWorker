// Package client provides gRPC connection management for Runner.
package client

import (
	"time"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

// SendPodCreated sends a pod_created event to the server (control message).
func (c *GRPCConnection) SendPodCreated(podKey string, pid int32, sandboxPath, branchName string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PodCreated{
			PodCreated: &runnerv1.PodCreatedEvent{
				PodKey:      podKey,
				Pid:         pid,
				SandboxPath: sandboxPath,
				BranchName:  branchName,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// SendPodTerminated sends a pod_terminated event to the server (control message).
func (c *GRPCConnection) SendPodTerminated(podKey string, exitCode int32, errorMsg string, status string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PodTerminated{
			PodTerminated: &runnerv1.PodTerminatedEvent{
				PodKey:       podKey,
				ExitCode:     exitCode,
				ErrorMessage: errorMsg,
				Status:       status,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	if c.queuePodTerminationUntilInitialized(podKey, msg) {
		return nil
	}
	if err := c.sendControl(msg); err != nil {
		c.retainPodTermination(podKey, msg)
		logger.GRPC().Warn("Queued pod termination after control send failed",
			"pod_key", podKey, "error", err)
	}
	return nil
}

// SendPodInitProgress sends a pod initialization progress event to the server (control message).
func (c *GRPCConnection) SendPodInitProgress(podKey, phase string, progress int32, message string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PodInitProgress{
			PodInitProgress: &runnerv1.PodInitProgressEvent{
				PodKey:   podKey,
				Phase:    phase,
				Progress: progress,
				Message:  message,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}

// SendPodRestarting sends a pod_restarting event when a perpetual pod auto-restarts.
func (c *GRPCConnection) SendPodRestarting(podKey string, exitCode, restartCount, newPID int32) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PodRestarting{
			PodRestarting: &runnerv1.PodRestartingEvent{
				PodKey:       podKey,
				ExitCode:     exitCode,
				RestartCount: restartCount,
				NewPid:       newPID,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendControl(msg)
}
