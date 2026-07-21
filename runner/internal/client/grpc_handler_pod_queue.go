package client

import (
	"fmt"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

const runnerQueueFullCode = "runner_queue_full"

type podCommandFailureReporter func(message string) error

func createAutopilotPodKey(cmd *runnerv1.CreateAutopilotCommand) string {
	if cmd.PodKey != "" {
		return cmd.PodKey
	}
	if cmd.PodConfig != nil {
		return cmd.PodConfig.PodKey
	}
	return ""
}

func (c *GRPCConnection) enqueuePodCommand(
	podKey string,
	operation string,
	run func(),
	reportFailure podCommandFailureReporter,
) {
	c.handlerWg.Add(1)
	err := c.podQueue.Enqueue(podKey, func() {
		defer c.handlerWg.Done()
		run()
	})
	if err == nil {
		return
	}

	c.handlerWg.Done()
	message := fmt.Sprintf("runner command queue rejected %s: %v", operation, err)
	log := logger.GRPC()
	log.Error("Pod command queue rejected command",
		"pod_key", podKey,
		"operation", operation,
		"error", err)
	if reportErr := reportFailure(message); reportErr != nil {
		log.Error("Failed to report pod command queue rejection",
			"pod_key", podKey,
			"operation", operation,
			"error", reportErr)
	}
}

func (c *GRPCConnection) podQueueFailureReporter(podKey string) podCommandFailureReporter {
	return func(message string) error {
		return c.SendError(podKey, runnerQueueFullCode, message)
	}
}

func (c *GRPCConnection) sandboxFsQueueFailureReporter(
	cmd *runnerv1.SandboxFsCommand,
) podCommandFailureReporter {
	return func(message string) error {
		return c.SendSandboxFsResult(&runnerv1.SandboxFsResultEvent{
			RequestId: cmd.RequestId,
			PodKey:    cmd.PodKey,
			Error:     message,
		})
	}
}

func (c *GRPCConnection) verificationQueueFailureReporter(
	cmd *runnerv1.RunVerificationCommand,
) podCommandFailureReporter {
	return func(message string) error {
		return c.SendVerificationResult(&runnerv1.VerificationResultEvent{
			RequestId: cmd.RequestId,
			PodKey:    cmd.PodKey,
			ExitCode:  -1,
			Error:     message,
		})
	}
}
