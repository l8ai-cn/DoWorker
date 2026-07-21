package agentworkbench

import (
	"errors"
	"math"
	"time"

	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

var ErrInvalidCommand = errors.New("invalid agent workbench command")

func ProjectAcceptedCommand(
	current *agentworkbenchv2.SessionSnapshot,
	command *agentworkbenchv2.CommandEnvelope,
	receipt *agentworkbenchv2.CommandReceipt,
) (*ProjectedBatch, error) {
	if err := validateAcceptedCommand(current, command, receipt); err != nil {
		return nil, err
	}
	snapshot := proto.Clone(current).(*agentworkbenchv2.SessionSnapshot)
	revision := snapshot.Revision + 1
	eventCount := uint64(1)
	if command.GetSendPrompt() != nil {
		eventCount++
	}
	if eventCount > math.MaxUint64-snapshot.LatestSequence {
		return nil, ErrInvalidCommand
	}
	firstSequence := snapshot.LatestSequence + 1
	events := make([]*agentworkbenchv2.AgentEvent, 0, eventCount)
	if prompt := command.GetSendPrompt(); prompt != nil {
		events = append(events, userMessageEvent(
			command,
			revision,
			firstSequence,
			prompt,
		))
	}
	events = append(events, receiptEvent(
		command,
		receipt,
		revision,
		firstSequence+uint64(len(events)),
	))
	for _, event := range events {
		if err := applyProjectedEvent(snapshot, event); err != nil {
			return nil, err
		}
	}
	lastSequence := firstSequence + uint64(len(events)) - 1
	delta := &agentworkbenchv2.SessionDeltaBatch{
		SessionId: current.SessionId, StreamEpoch: current.StreamEpoch,
		BaseRevision: current.Revision, Revision: revision,
		FirstSequence: firstSequence, LastSequence: lastSequence,
		Events: events,
	}
	deltaDigest, err := deterministicDigest(delta)
	if err != nil {
		return nil, err
	}
	delta.Digest = deltaDigest
	snapshot.Revision = revision
	snapshot.LatestSequence = lastSequence
	snapshot.ActiveTurnId = stringPointer(command.CommandId)
	snapshot.Digest = nil
	snapshotDigest, err := deterministicDigest(snapshot)
	if err != nil {
		return nil, err
	}
	snapshot.Digest = stringPointer(snapshotDigest)
	return &ProjectedBatch{Snapshot: snapshot, Delta: delta}, nil
}

func validateAcceptedCommand(
	current *agentworkbenchv2.SessionSnapshot,
	command *agentworkbenchv2.CommandEnvelope,
	receipt *agentworkbenchv2.CommandReceipt,
) error {
	if current == nil || command == nil || receipt == nil ||
		current.Revision == math.MaxUint64 ||
		command.SessionId != current.SessionId ||
		command.StreamEpoch != current.StreamEpoch ||
		command.CommandId == "" || command.PayloadDigest == "" ||
		command.IssuedAt == "" || command.Command == nil ||
		receipt.SessionId != command.SessionId ||
		receipt.CommandId != command.CommandId ||
		receipt.PayloadDigest != command.PayloadDigest ||
		receipt.State != agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_ACCEPTED {
		return ErrInvalidCommand
	}
	if command.ExpectedRevision != nil && *command.ExpectedRevision != current.Revision {
		return ErrInvalidCommand
	}
	if _, err := time.Parse(time.RFC3339Nano, command.IssuedAt); err != nil {
		return ErrInvalidCommand
	}
	return nil
}

func userMessageEvent(
	command *agentworkbenchv2.CommandEnvelope,
	revision uint64,
	sequence uint64,
	prompt *agentworkbenchv2.SendPromptCommand,
) *agentworkbenchv2.AgentEvent {
	content := make([]*agentworkbenchv2.ContentBlock, 0, len(prompt.Attachments)+1)
	if prompt.Text != "" {
		content = append(content, &agentworkbenchv2.ContentBlock{
			ContentId: command.CommandId + ":text",
			Identity: &agentworkbenchv2.ContentIdentity{
				Namespace: "agentcloud", SemanticKey: "text", SchemaVersion: "1",
			},
			Content: &agentworkbenchv2.ContentBlock_Text{
				Text: &agentworkbenchv2.TextContent{Text: prompt.Text},
			},
		})
	}
	for _, attachment := range prompt.Attachments {
		content = append(content, proto.Clone(attachment).(*agentworkbenchv2.ContentBlock))
	}
	return &agentworkbenchv2.AgentEvent{
		Envelope: commandEnvelope(command, revision, sequence, "message:"+command.CommandId),
		Event: &agentworkbenchv2.AgentEvent_TimelineItemAppended{
			TimelineItemAppended: &agentworkbenchv2.TimelineItemAppended{
				Content: &agentworkbenchv2.TimelineItemContent{
					Content: &agentworkbenchv2.TimelineItemContent_Message{
						Message: &agentworkbenchv2.MessageTimelineItem{
							Role:    agentworkbenchv2.MessageRole_MESSAGE_ROLE_USER,
							Content: content,
							Status:  agentworkbenchv2.TimelineItemStatus_TIMELINE_ITEM_STATUS_COMPLETED,
						},
					},
				},
			},
		},
	}
}

func receiptEvent(
	command *agentworkbenchv2.CommandEnvelope,
	receipt *agentworkbenchv2.CommandReceipt,
	revision uint64,
	sequence uint64,
) *agentworkbenchv2.AgentEvent {
	return &agentworkbenchv2.AgentEvent{
		Envelope: commandEnvelope(command, revision, sequence, "receipt:"+command.CommandId),
		Event: &agentworkbenchv2.AgentEvent_CommandReceiptChanged{
			CommandReceiptChanged: &agentworkbenchv2.CommandReceiptChanged{
				Receipt: proto.Clone(receipt).(*agentworkbenchv2.CommandReceipt),
			},
		},
	}
}

func commandEnvelope(
	command *agentworkbenchv2.CommandEnvelope,
	revision uint64,
	sequence uint64,
	itemID string,
) *agentworkbenchv2.EventEnvelope {
	return &agentworkbenchv2.EventEnvelope{
		SessionId: command.SessionId, StreamEpoch: command.StreamEpoch,
		Revision: revision, Sequence: sequence, ItemId: itemID,
		TurnId:             stringPointer(command.CommandId),
		CausationCommandId: stringPointer(command.CommandId),
		CreatedAt:          command.IssuedAt,
	}
}
