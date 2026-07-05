package client

import (
	"time"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (c *GRPCConnection) SendAcpSessionEvent(podKey, eventType, jsonPayload string) error {
	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_AcpSession{
			AcpSession: &runnerv1.AcpSessionEvent{
				PodKey:       podKey,
				EventType:    eventType,
				JsonPayload:  jsonPayload,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return c.sendTerminal(msg)
}
