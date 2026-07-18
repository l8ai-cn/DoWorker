package runner

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"
	"google.golang.org/protobuf/proto"
)

func (d *PendingCommandDrainer) dispatchOne(
	ctx context.Context,
	runnerID int64,
	row *agentpod.PendingCommand,
) (processed bool, stop bool) {
	payload, err := d.payloadCipher.decrypt(row.Payload)
	if err != nil {
		d.logger.Error("cannot decrypt pending command", "id", row.ID, "error", err)
		return false, true
	}
	var msg runnerv1.ServerMessage
	if err := proto.Unmarshal(payload, &msg); err != nil {
		d.logger.Error("dropping corrupt pending command", "id", row.ID, "error", err)
		_ = d.repo.Delete(ctx, row.ID)
		return true, false
	}
	if !pendingMessageMatchesRow(row, &msg) {
		d.logger.Error("pending command envelope does not match row metadata", "id", row.ID)
		return false, true
	}
	switch row.CommandType {
	case agentpod.CommandTypeCreatePod:
		return d.dispatchCreatePod(ctx, runnerID, row, msg.GetCreatePod())
	case agentpod.CommandTypeSendPrompt:
		return d.dispatchSendPrompt(ctx, runnerID, row, msg.GetSendPrompt())
	default:
		_ = d.repo.Delete(ctx, row.ID)
		return true, false
	}
}

func pendingMessageMatchesRow(row *agentpod.PendingCommand, msg *runnerv1.ServerMessage) bool {
	switch row.CommandType {
	case agentpod.CommandTypeCreatePod:
		command := msg.GetCreatePod()
		return command != nil && command.GetPodKey() == row.PodKey
	case agentpod.CommandTypeSendPrompt:
		command := msg.GetSendPrompt()
		return command != nil &&
			command.GetPodKey() == row.PodKey &&
			command.GetCommandId() == row.CommandID
	default:
		return false
	}
}
