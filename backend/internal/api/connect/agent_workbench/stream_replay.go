package agentworkbenchconnect

import (
	"context"

	"connectrpc.com/connect"
	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

const (
	defaultReplayPageSize = 100
	maxReplayPageSize     = 1000
)

func (s *Server) replayTo(
	ctx context.Context,
	stream *connect.ServerStream[agentworkbenchv2.SessionDeltaBatch],
	position streamPosition,
	target *agentworkbenchv2.SessionSnapshot,
	pageSize int,
) (streamPosition, error) {
	if err := validateReplayTarget(position, target); err != nil {
		return position, err
	}
	if position.sequence == target.GetLatestSequence() {
		if position.revision != target.GetRevision() {
			return position, dataLoss("agent workbench replay revision is inconsistent")
		}
		return position, nil
	}
	scanned := position.sequence
	var pending *agentworkbenchv2.SessionDeltaBatch
	for scanned < target.GetLatestSequence() {
		previousScanned := scanned
		records, err := s.repository.ListAfter(
			ctx,
			position.sessionID,
			position.streamEpoch,
			scanned,
			pageSize,
		)
		if err != nil {
			return position, internalError(err)
		}
		if len(records) == 0 {
			return position, dataLoss("agent workbench replay has a missing event")
		}
		for _, record := range records {
			if record.Sequence > target.GetLatestSequence() {
				break
			}
			event, err := decodeEventRecord(record, position, scanned+1)
			if err != nil {
				return position, err
			}
			scanned = record.Sequence
			if pending == nil {
				pending, err = newReplayBatch(position, event)
				if err != nil {
					return position, err
				}
				continue
			}
			if event.GetEnvelope().GetRevision() == pending.GetRevision() {
				appendReplayEvent(pending, event)
				continue
			}
			position, err = sendReplayBatch(ctx, stream, position, pending)
			if err != nil {
				return position, err
			}
			pending, err = newReplayBatch(position, event)
			if err != nil {
				return position, err
			}
		}
		if scanned == previousScanned {
			return position, dataLoss("agent workbench replay has a missing event")
		}
	}
	if pending != nil {
		var err error
		position, err = sendReplayBatch(ctx, stream, position, pending)
		if err != nil {
			return position, err
		}
	}
	if position.revision != target.GetRevision() ||
		position.sequence != target.GetLatestSequence() {
		return position, dataLoss("agent workbench replay does not reach the snapshot")
	}
	return position, nil
}

func decodeEventRecord(
	record workbenchdomain.Event,
	position streamPosition,
	expectedSequence uint64,
) (*agentworkbenchv2.AgentEvent, error) {
	if record.SessionID != position.sessionID ||
		record.StreamEpoch != position.streamEpoch ||
		record.Sequence != expectedSequence {
		return nil, dataLoss("agent workbench event record metadata is invalid")
	}
	event := &agentworkbenchv2.AgentEvent{}
	if err := proto.Unmarshal(record.Payload, event); err != nil {
		return nil, dataLoss("agent workbench event protobuf is invalid")
	}
	envelope := event.GetEnvelope()
	if envelope == nil ||
		envelope.GetSessionId() != record.SessionID ||
		envelope.GetStreamEpoch() != record.StreamEpoch ||
		envelope.GetRevision() != record.Revision ||
		envelope.GetSequence() != record.Sequence ||
		event.GetEvent() == nil {
		return nil, dataLoss("agent workbench event metadata is inconsistent")
	}
	digest, err := protoDigest(event)
	if err != nil {
		return nil, internalError(err)
	}
	if digest != record.Digest {
		return nil, dataLoss("agent workbench event digest is invalid")
	}
	return event, nil
}

func newReplayBatch(
	position streamPosition,
	event *agentworkbenchv2.AgentEvent,
) (*agentworkbenchv2.SessionDeltaBatch, error) {
	envelope := event.GetEnvelope()
	if envelope.GetRevision() != position.revision+1 ||
		envelope.GetSequence() != position.sequence+1 {
		return nil, failedPrecondition(
			"agent workbench cursor is not on a revision boundary; resync required",
		)
	}
	return &agentworkbenchv2.SessionDeltaBatch{
		SessionId:     position.sessionID,
		StreamEpoch:   position.streamEpoch,
		BaseRevision:  position.revision,
		Revision:      envelope.GetRevision(),
		FirstSequence: envelope.GetSequence(),
		LastSequence:  envelope.GetSequence(),
		Events:        []*agentworkbenchv2.AgentEvent{event},
	}, nil
}

func appendReplayEvent(
	batch *agentworkbenchv2.SessionDeltaBatch,
	event *agentworkbenchv2.AgentEvent,
) {
	batch.Events = append(batch.Events, event)
	batch.LastSequence = event.GetEnvelope().GetSequence()
}

func sendReplayBatch(
	ctx context.Context,
	stream *connect.ServerStream[agentworkbenchv2.SessionDeltaBatch],
	position streamPosition,
	batch *agentworkbenchv2.SessionDeltaBatch,
) (streamPosition, error) {
	digest, err := deltaDigest(batch)
	if err != nil {
		return position, internalError(err)
	}
	batch.Digest = digest
	decorated, err := authorizedDelta(ctx, batch)
	if err != nil {
		return position, internalError(err)
	}
	if err := stream.Send(decorated); err != nil {
		if ctx.Err() != nil {
			return position, canceled(ctx.Err())
		}
		return position, err
	}
	return streamPosition{
		sessionID:   position.sessionID,
		streamEpoch: position.streamEpoch,
		revision:    batch.GetRevision(),
		sequence:    batch.GetLastSequence(),
	}, nil
}

func replayPageSize(requested uint32) int {
	if requested == 0 {
		return defaultReplayPageSize
	}
	if requested > maxReplayPageSize {
		return maxReplayPageSize
	}
	return int(requested)
}
