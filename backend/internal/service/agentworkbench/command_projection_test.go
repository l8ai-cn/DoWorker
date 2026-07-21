package agentworkbench

import (
	"testing"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"github.com/stretchr/testify/require"
)

func TestProjectAcceptedCommandAddsUserMessageAndReceipt(t *testing.T) {
	current := initialSnapshot("conv_1", "stream-1")
	command := &agentworkbenchv2.CommandEnvelope{
		SessionId: "conv_1", StreamEpoch: "stream-1",
		CommandId: "command-1", PayloadDigest: validDigest("1"),
		IssuedAt: "2026-07-16T10:00:00Z",
		Command: &agentworkbenchv2.CommandEnvelope_SendPrompt{
			SendPrompt: &agentworkbenchv2.SendPromptCommand{Text: "生成一段演示视频"},
		},
	}
	receipt := acceptedReceipt(command)

	projected, err := ProjectAcceptedCommand(current, command, receipt)
	require.NoError(t, err)
	require.Equal(t, uint64(1), projected.Snapshot.Revision)
	require.Equal(t, uint64(2), projected.Snapshot.LatestSequence)
	require.Len(t, projected.Snapshot.History, 1)
	require.Equal(
		t,
		agentworkbenchv2.MessageRole_MESSAGE_ROLE_USER,
		projected.Snapshot.History[0].GetContent().GetMessage().GetRole(),
	)
	require.Equal(
		t,
		"生成一段演示视频",
		projected.Snapshot.History[0].GetContent().GetMessage().GetContent()[0].GetText().GetText(),
	)
	require.Len(t, projected.Snapshot.CommandReceipts, 1)
	require.Equal(t, "command-1", projected.Snapshot.CommandReceipts[0].CommandId)
	require.Len(t, projected.Delta.Events, 2)
}

func TestProjectAcceptedCommandRejectsMismatchedReceipt(t *testing.T) {
	current := initialSnapshot("conv_1", "stream-1")
	command := &agentworkbenchv2.CommandEnvelope{
		SessionId: "conv_1", StreamEpoch: "stream-1",
		CommandId: "command-1", PayloadDigest: validDigest("1"),
		IssuedAt: "2026-07-16T10:00:00Z",
		Command: &agentworkbenchv2.CommandEnvelope_Interrupt{
			Interrupt: &agentworkbenchv2.InterruptCommand{},
		},
	}
	receipt := acceptedReceipt(command)
	receipt.PayloadDigest = validDigest("2")

	_, err := ProjectAcceptedCommand(current, command, receipt)
	require.ErrorIs(t, err, ErrInvalidCommand)
}

func acceptedReceipt(
	command *agentworkbenchv2.CommandEnvelope,
) *agentworkbenchv2.CommandReceipt {
	return &agentworkbenchv2.CommandReceipt{
		SessionId: command.SessionId, CommandId: command.CommandId,
		State:         agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_ACCEPTED,
		PayloadDigest: command.PayloadDigest,
		ReceivedAt:    stringPointer("2026-07-16T10:00:00Z"),
		UpdatedAt:     stringPointer("2026-07-16T10:00:01Z"),
	}
}

func validDigest(fill string) string {
	value := ""
	for len(value) < 64 {
		value += fill
	}
	return "sha256:" + value[:64]
}
