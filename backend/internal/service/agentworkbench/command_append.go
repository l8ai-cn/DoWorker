package agentworkbench

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainworkbench "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/l8ai-cn/agentcloud/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

const commandAppendAttempts = 8

func (dispatcher *CommandDispatcher) appendAccepted(
	ctx context.Context,
	command *agentworkbenchv2.CommandEnvelope,
) (*agentworkbenchv2.CommandReceipt, error) {
	for attempt := 0; attempt < commandAppendAttempts; attempt++ {
		stored, err := dispatcher.repository.GetSnapshot(ctx, command.SessionId)
		if err != nil {
			return nil, err
		}
		snapshot, err := decodeStoredSnapshot(command.SessionId, stored)
		if err != nil {
			return nil, err
		}
		now := dispatcher.now().UTC().Format(timeFormat)
		receipt := &agentworkbenchv2.CommandReceipt{
			SessionId: command.SessionId, CommandId: command.CommandId,
			State:             agentworkbenchv2.CommandReceiptState_COMMAND_RECEIPT_STATE_ACCEPTED,
			PayloadDigest:     command.PayloadDigest,
			ResultingRevision: uint64Pointer(snapshot.Revision + 1),
			UpdatedAt:         stringPointer(now),
		}
		projectionCommand := cloneCommand(command)
		projectionCommand.ExpectedRevision = nil
		projected, err := ProjectAcceptedCommand(snapshot, projectionCommand, receipt)
		if err != nil {
			return nil, err
		}
		request, err := commandAppendRequest(stored, receipt, projected)
		if err != nil {
			return nil, err
		}
		result, err := dispatcher.repository.Append(ctx, request)
		if errors.Is(err, domainworkbench.ErrRevisionConflict) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("append accepted command: %w", err)
		}
		if result.Applied {
			dispatcher.publisher.Publish(command.SessionId, projected.Delta)
		}
		return receipt, nil
	}
	return nil, domainworkbench.ErrRevisionConflict
}

func commandAppendRequest(
	stored *domainworkbench.SessionState,
	receipt *agentworkbenchv2.CommandReceipt,
	projected *ProjectedBatch,
) (domainworkbench.AppendRequest, error) {
	projection, err := marshalDeterministic(projected.Snapshot)
	if err != nil {
		return domainworkbench.AppendRequest{}, err
	}
	domainReceipt, err := domainReceipt(receipt)
	if err != nil {
		return domainworkbench.AppendRequest{}, err
	}
	request := domainworkbench.AppendRequest{
		SessionID: projected.Snapshot.SessionId, ExpectedRevision: stored.Revision,
		Projection: domainworkbench.SessionState{
			SessionID:      projected.Snapshot.SessionId,
			StreamEpoch:    projected.Snapshot.StreamEpoch,
			Revision:       projected.Snapshot.Revision,
			LatestSequence: projected.Snapshot.LatestSequence,
			Projection:     projection, Digest: projected.Snapshot.GetDigest(),
		},
		Receipts: []domainworkbench.CommandReceipt{domainReceipt},
		Events:   make([]domainworkbench.Event, len(projected.Delta.Events)),
	}
	for index, event := range projected.Delta.Events {
		payload, err := marshalDeterministic(event)
		if err != nil {
			return domainworkbench.AppendRequest{}, err
		}
		digest, err := deterministicDigest(event)
		if err != nil {
			return domainworkbench.AppendRequest{}, err
		}
		createdAt, err := time.Parse(time.RFC3339Nano, event.GetEnvelope().GetCreatedAt())
		if err != nil {
			return domainworkbench.AppendRequest{}, ErrInvalidCommand
		}
		request.Events[index] = domainworkbench.Event{
			SessionID:   projected.Snapshot.SessionId,
			StreamEpoch: projected.Snapshot.StreamEpoch,
			Revision:    projected.Snapshot.Revision,
			Sequence:    event.GetEnvelope().GetSequence(),
			Payload:     payload, Digest: digest,
			CausationCommandID: event.GetEnvelope().CausationCommandId,
			CreatedAt:          createdAt,
		}
	}
	return request, nil
}

func decodeStoredSnapshot(
	sessionID string,
	stored *domainworkbench.SessionState,
) (*agentworkbenchv2.SessionSnapshot, error) {
	if stored == nil {
		return nil, ErrCommandUnavailable
	}
	snapshot := &agentworkbenchv2.SessionSnapshot{}
	if err := proto.Unmarshal(stored.Projection, snapshot); err != nil {
		return nil, err
	}
	if stored.SessionID != sessionID || snapshot.SessionId != sessionID ||
		stored.StreamEpoch != snapshot.StreamEpoch ||
		stored.Revision != snapshot.Revision ||
		stored.LatestSequence != snapshot.LatestSequence ||
		stored.Digest != snapshot.GetDigest() {
		return nil, ErrInvalidCommand
	}
	return snapshot, nil
}

func uint64Pointer(value uint64) *uint64 {
	return &value
}

func cloneCommand(
	command *agentworkbenchv2.CommandEnvelope,
) *agentworkbenchv2.CommandEnvelope {
	return proto.Clone(command).(*agentworkbenchv2.CommandEnvelope)
}
