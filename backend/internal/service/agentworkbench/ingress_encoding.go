package agentworkbench

import (
	"fmt"
	"time"

	domainworkbench "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func appendRequest(
	sessionID string,
	stored *domainworkbench.SessionState,
	batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
	projected *ProjectedBatch,
) (domainworkbench.AppendRequest, error) {
	projection, err := marshalDeterministic(projected.Snapshot)
	if err != nil {
		return domainworkbench.AppendRequest{}, err
	}
	expectedRevision := uint64(0)
	if stored != nil {
		expectedRevision = stored.Revision
	}
	request := domainworkbench.AppendRequest{
		SessionID:        sessionID,
		ExpectedRevision: expectedRevision,
		Projection: domainworkbench.SessionState{
			SessionID: sessionID, StreamEpoch: projected.Snapshot.StreamEpoch,
			Revision:       projected.Snapshot.Revision,
			LatestSequence: projected.Snapshot.LatestSequence,
			Projection:     projection, Digest: projected.Snapshot.GetDigest(),
		},
		Sources: make([]domainworkbench.SourceEvent, len(batch.Mutations)),
		Events:  make([]domainworkbench.Event, len(projected.Delta.Events)),
	}
	for index, mutation := range batch.Mutations {
		source := mutation.GetSource()
		createdAt, err := time.Parse(time.RFC3339Nano, source.GetOccurredAt())
		if err != nil {
			return domainworkbench.AppendRequest{}, ErrInvalidBatch
		}
		sourceDigest, err := deterministicDigest(source)
		if err != nil {
			return domainworkbench.AppendRequest{}, err
		}
		event := projected.Delta.Events[index]
		payload, err := marshalDeterministic(event)
		if err != nil {
			return domainworkbench.AppendRequest{}, err
		}
		eventDigest, err := deterministicDigest(event)
		if err != nil {
			return domainworkbench.AppendRequest{}, err
		}
		request.Sources[index] = domainworkbench.SourceEvent{
			SessionID: sessionID, StableEventID: source.StableEventId,
			RunnerSessionEpoch: batch.RunnerSessionEpoch,
			SourceSequence:     source.SourceSequence,
			PayloadDigest:      sourceDigest,
		}
		request.Events[index] = domainworkbench.Event{
			SessionID: sessionID, StreamEpoch: projected.Delta.StreamEpoch,
			Revision: projected.Delta.Revision, Sequence: event.GetEnvelope().GetSequence(),
			Payload: payload, Digest: eventDigest,
			CausationCommandID: event.GetEnvelope().CausationCommandId,
			CreatedAt:          createdAt,
		}
	}
	return request, nil
}

func marshalDeterministic(message proto.Message) ([]byte, error) {
	data, err := (proto.MarshalOptions{Deterministic: true}).Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("encode agent workbench protobuf: %w", err)
	}
	return data, nil
}
