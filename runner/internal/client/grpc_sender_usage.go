package client

import (
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

func (c *GRPCConnection) SendPodUsageEvent(podKey, model string, in, out, cacheRead, cacheCreate int64) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_PodUsage{
			PodUsage: &runnerv1.PodUsageEvent{
				PodKey: podKey, Model: model,
				InputTokens: in, OutputTokens: out,
				CacheReadTokens: cacheRead, CacheCreationTokens: cacheCreate,
			},
		},
	}
	return c.SendMessage(msg)
}

func (c *GRPCConnection) SendExternalSessionCaptured(podKey, externalID string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_ExternalSessionCaptured{
			ExternalSessionCaptured: &runnerv1.ExternalSessionCapturedEvent{
				PodKey: podKey, ExternalSessionId: externalID,
			},
		},
	}
	return c.SendMessage(msg)
}
