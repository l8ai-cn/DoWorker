package agentworkbench

import (
	"errors"
	"math"

	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

var ErrInvalidBatch = errors.New("invalid runner workbench batch")

type ProjectedBatch struct {
	Snapshot *agentworkbenchv2.SessionSnapshot
	Delta    *agentworkbenchv2.SessionDeltaBatch
}

func ProjectRunnerBatch(
	current *agentworkbenchv2.SessionSnapshot,
	sessionID string,
	streamEpoch string,
	batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
) (*ProjectedBatch, error) {
	if err := validateProjectionInput(current, sessionID, streamEpoch, batch); err != nil {
		return nil, err
	}
	snapshot := initialSnapshot(sessionID, streamEpoch)
	if current != nil {
		snapshot = proto.Clone(current).(*agentworkbenchv2.SessionSnapshot)
	}
	if snapshot.Revision == math.MaxUint64 ||
		uint64(len(batch.Mutations)) > math.MaxUint64-snapshot.LatestSequence {
		return nil, ErrInvalidBatch
	}
	revision := snapshot.Revision + 1
	firstSequence := snapshot.LatestSequence + 1
	events := make([]*agentworkbenchv2.AgentEvent, 0, len(batch.Mutations))
	for index, mutation := range batch.Mutations {
		event, err := projectMutation(
			sessionID,
			streamEpoch,
			revision,
			firstSequence+uint64(index),
			mutation,
		)
		if err != nil {
			return nil, err
		}
		if err := applyProjectedEvent(snapshot, event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := validateProjectedArtifactReferences(snapshot); err != nil {
		return nil, err
	}
	lastSequence := firstSequence + uint64(len(events)) - 1
	delta := &agentworkbenchv2.SessionDeltaBatch{
		SessionId:     sessionID,
		StreamEpoch:   streamEpoch,
		BaseRevision:  snapshot.Revision,
		Revision:      revision,
		FirstSequence: firstSequence,
		LastSequence:  lastSequence,
		Events:        events,
	}
	deltaDigest, err := deterministicDigest(delta)
	if err != nil {
		return nil, err
	}
	delta.Digest = deltaDigest
	snapshot.Revision = revision
	snapshot.LatestSequence = lastSequence
	snapshot.Digest = nil
	snapshotDigest, err := deterministicDigest(snapshot)
	if err != nil {
		return nil, err
	}
	snapshot.Digest = stringPointer(snapshotDigest)
	return &ProjectedBatch{Snapshot: snapshot, Delta: delta}, nil
}

func initialSnapshot(
	sessionID string,
	streamEpoch string,
) *agentworkbenchv2.SessionSnapshot {
	return &agentworkbenchv2.SessionSnapshot{
		SessionId:   sessionID,
		StreamEpoch: streamEpoch,
		Status:      agentworkbenchv2.SessionStatus_SESSION_STATUS_IDLE,
	}
}

func stringPointer(value string) *string {
	return &value
}

func validateProjectionInput(
	current *agentworkbenchv2.SessionSnapshot,
	sessionID string,
	streamEpoch string,
	batch *agentworkbenchv2.RunnerWorkbenchEventBatch,
) error {
	if sessionID == "" || streamEpoch == "" || batch == nil ||
		batch.PodKey == "" || batch.AdapterId == "" ||
		batch.SourceProtocolVersion == "" || batch.RunnerSessionEpoch == "" ||
		len(batch.Mutations) == 0 {
		return ErrInvalidBatch
	}
	if current != nil &&
		(current.SessionId != sessionID || current.StreamEpoch != streamEpoch) {
		return ErrInvalidBatch
	}
	seen := make(map[string]struct{}, len(batch.Mutations))
	var previous uint64
	for _, mutation := range batch.Mutations {
		source := mutation.GetSource()
		if source == nil || source.StableEventId == "" ||
			source.SourceSequence == 0 || source.OccurredAt == "" ||
			source.SourceSequence <= previous {
			return ErrInvalidBatch
		}
		key := batch.RunnerSessionEpoch + ":" + source.StableEventId
		if _, exists := seen[key]; exists {
			return ErrInvalidBatch
		}
		seen[key] = struct{}{}
		previous = source.SourceSequence
	}
	return nil
}
