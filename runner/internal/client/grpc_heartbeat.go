package client

import (
	"context"
	"time"

	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
	"github.com/l8ai-cn/agentcloud/runner/internal/logger"
)

func (c *GRPCConnection) heartbeatLoop(ctx context.Context, done <-chan struct{}) {
	ticker := time.NewTicker(c.heartbeatInterval)
	defer ticker.Stop()

	c.sendHeartbeat()
	if c.heartbeatMonitor != nil {
		c.heartbeatMonitor.OnSent()
	}

	for {
		select {
		case <-c.stopCh:
			return
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.sendHeartbeat()
			if c.heartbeatMonitor != nil {
				c.heartbeatMonitor.OnSent()
			}
		}
	}
}

func (c *GRPCConnection) sendHeartbeat() {
	var pods []*runnerv1.PodInfo
	var relayConnections []*runnerv1.RelayConnectionInfo

	if c.handler != nil {
		for _, pod := range c.handler.OnListPods() {
			pods = append(pods, &runnerv1.PodInfo{
				PodKey:      pod.PodKey,
				Status:      pod.Status,
				AgentStatus: pod.AgentStatus,
			})
		}
		for _, connection := range c.handler.OnListRelayConnections() {
			relayConnections = append(relayConnections, &runnerv1.RelayConnectionInfo{
				PodKey:      connection.PodKey,
				RelayUrl:    connection.RelayURL,
				Connected:   connection.Connected,
				ConnectedAt: connection.ConnectedAt,
			})
		}
	}

	var agentVersions []*runnerv1.AgentVersionInfo
	if c.agentProbe != nil {
		agentVersions = c.agentProbe.ProbeAndDiff()
		if len(agentVersions) > 0 {
			c.mu.Lock()
			c.availableAgents = c.agentProbe.GetAvailableAgents()
			c.mu.Unlock()
		}
	}

	msg := &runnerv1.RunnerMessage{
		Payload: &runnerv1.RunnerMessage_Heartbeat{
			Heartbeat: &runnerv1.HeartbeatData{
				NodeId:           c.nodeID,
				Pods:             pods,
				RelayConnections: relayConnections,
				AgentVersions:    agentVersions,
			},
		},
		Timestamp: time.Now().UnixMilli(),
	}

	logger.GRPC().Debug(
		"Sending heartbeat",
		"pods", len(pods),
		"relay_connections", len(relayConnections),
		"version_changes", len(agentVersions),
	)
	if err := c.sendControl(msg); err != nil {
		logger.GRPC().Error("Failed to send heartbeat", "error", err)
	}
}
