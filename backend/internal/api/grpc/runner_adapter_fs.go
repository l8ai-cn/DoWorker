package grpc

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
)

func (a *GRPCRunnerAdapter) SendSandboxFs(runnerID int64, cmd *runnerv1.SandboxFsCommand) error {
	conn := a.connManager.GetConnection(runnerID)
	if conn == nil {
		return status.Errorf(codes.NotFound, "runner %d not connected", runnerID)
	}
	msg := &runnerv1.ServerMessage{
		Payload: &runnerv1.ServerMessage_SandboxFs{
			SandboxFs: cmd,
		},
		Timestamp: time.Now().UnixMilli(),
	}
	return conn.SendMessage(msg)
}
