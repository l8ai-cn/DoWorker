package client

import (
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"github.com/anthropics/agentsmesh/runner/internal/logger"
)

func (c *GRPCConnection) queuePodTerminationUntilInitialized(
	podKey string,
	msg *runnerv1.RunnerMessage,
) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stream != nil && c.initialized {
		return false
	}
	if c.pendingPodTerminations == nil {
		c.pendingPodTerminations = make(map[string]*runnerv1.RunnerMessage)
	}
	c.pendingPodTerminations[podKey] = msg
	logger.GRPC().Info("Queued pod termination until connection is initialized", "pod_key", podKey)
	return true
}

func (c *GRPCConnection) retainPodTermination(
	podKey string,
	msg *runnerv1.RunnerMessage,
) {
	c.mu.Lock()
	if c.pendingPodTerminations == nil {
		c.pendingPodTerminations = make(map[string]*runnerv1.RunnerMessage)
	}
	c.pendingPodTerminations[podKey] = msg
	c.mu.Unlock()
}

func (c *GRPCConnection) flushPendingPodTerminations() {
	c.mu.Lock()
	if c.stream == nil || !c.initialized || len(c.pendingPodTerminations) == 0 {
		c.mu.Unlock()
		return
	}
	pending := c.pendingPodTerminations
	c.pendingPodTerminations = make(map[string]*runnerv1.RunnerMessage)
	c.mu.Unlock()

	for podKey, msg := range pending {
		if err := c.sendControl(msg); err != nil {
			c.retainPodTermination(podKey, msg)
			logger.GRPC().Warn("Failed to flush pending pod termination",
				"pod_key", podKey, "error", err)
			continue
		}
		logger.GRPC().Info("Flushed pending pod termination", "pod_key", podKey)
	}
}
