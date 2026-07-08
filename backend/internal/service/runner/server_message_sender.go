package runner

import (
	"context"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

type ConnectionChecker interface {
	IsConnected(runnerID int64) bool
}

type ServerMessageSender interface {
	SendServerMessage(ctx context.Context, runnerID int64, msg *runnerv1.ServerMessage) error
}
