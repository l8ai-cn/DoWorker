package agentworkbenchconnect

import (
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
)

type streamPosition struct {
	sessionID   string
	streamEpoch string
	revision    uint64
	sequence    uint64
}

func validateCursor(
	cursor *agentworkbenchv2.SessionCursor,
	snapshot *agentworkbenchv2.SessionSnapshot,
) (streamPosition, error) {
	if cursor.GetSessionId() == "" || cursor.GetStreamEpoch() == "" {
		return streamPosition{}, invalidArgument(
			"cursor.session_id and cursor.stream_epoch are required",
		)
	}
	if cursor.GetSessionId() != snapshot.GetSessionId() ||
		cursor.GetStreamEpoch() != snapshot.GetStreamEpoch() {
		return streamPosition{}, failedPrecondition(
			"agent workbench cursor does not match the current snapshot; resync required",
		)
	}
	if cursor.GetRevision() > snapshot.GetRevision() ||
		cursor.GetSequence() > snapshot.GetLatestSequence() ||
		cursor.GetRevision() > cursor.GetSequence() ||
		(cursor.GetRevision() == 0) != (cursor.GetSequence() == 0) {
		return streamPosition{}, failedPrecondition(
			"agent workbench cursor is ahead of the current snapshot; resync required",
		)
	}
	if cursor.GetRevision() == snapshot.GetRevision() &&
		cursor.GetSequence() != snapshot.GetLatestSequence() {
		return streamPosition{}, failedPrecondition(
			"agent workbench cursor is not on a revision boundary; resync required",
		)
	}
	return streamPosition{
		sessionID:   cursor.GetSessionId(),
		streamEpoch: cursor.GetStreamEpoch(),
		revision:    cursor.GetRevision(),
		sequence:    cursor.GetSequence(),
	}, nil
}

func validateReplayTarget(
	position streamPosition,
	snapshot *agentworkbenchv2.SessionSnapshot,
) error {
	if snapshot.GetSessionId() != position.sessionID ||
		snapshot.GetStreamEpoch() != position.streamEpoch ||
		snapshot.GetRevision() < position.revision ||
		snapshot.GetLatestSequence() < position.sequence {
		return failedPrecondition(
			"agent workbench snapshot changed; resync required",
		)
	}
	return nil
}

func validateLiveDelta(
	delta *agentworkbenchv2.SessionDeltaBatch,
	position streamPosition,
) (bool, streamPosition, error) {
	if delta == nil ||
		delta.GetSessionId() != position.sessionID ||
		delta.GetStreamEpoch() != position.streamEpoch {
		return false, position, failedPrecondition(
			"agent workbench live stream changed; resync required",
		)
	}
	if err := validateDeltaDigest(delta); err != nil {
		return false, position, err
	}
	if delta.GetLastSequence() <= position.sequence {
		if delta.GetRevision() <= position.revision {
			return true, position, nil
		}
		return false, position, dataLoss("agent workbench live delta order is invalid")
	}
	if delta.GetFirstSequence() <= position.sequence {
		return false, position, failedPrecondition(
			"agent workbench live delta overlaps the cursor; resync required",
		)
	}
	if delta.GetBaseRevision() != position.revision ||
		delta.GetRevision() != position.revision+1 ||
		delta.GetFirstSequence() != position.sequence+1 {
		return false, position, failedPrecondition(
			"agent workbench live delta has a gap; resync required",
		)
	}
	if err := validateDeltaEvents(delta); err != nil {
		return false, position, err
	}
	return false, streamPosition{
		sessionID:   position.sessionID,
		streamEpoch: position.streamEpoch,
		revision:    delta.GetRevision(),
		sequence:    delta.GetLastSequence(),
	}, nil
}

func validateDeltaDigest(delta *agentworkbenchv2.SessionDeltaBatch) error {
	digest, err := deltaDigest(delta)
	if err != nil {
		return internalError(err)
	}
	if digest != delta.GetDigest() {
		return dataLoss("agent workbench delta digest is invalid")
	}
	return nil
}

func validateDeltaEvents(delta *agentworkbenchv2.SessionDeltaBatch) error {
	if delta.GetFirstSequence() == 0 ||
		delta.GetLastSequence() < delta.GetFirstSequence() ||
		uint64(len(delta.GetEvents())) !=
			delta.GetLastSequence()-delta.GetFirstSequence()+1 {
		return dataLoss("agent workbench delta range is invalid")
	}
	for index, event := range delta.GetEvents() {
		envelope := event.GetEnvelope()
		if envelope == nil ||
			envelope.GetSessionId() != delta.GetSessionId() ||
			envelope.GetStreamEpoch() != delta.GetStreamEpoch() ||
			envelope.GetRevision() != delta.GetRevision() ||
			envelope.GetSequence() != delta.GetFirstSequence()+uint64(index) {
			return dataLoss("agent workbench delta event metadata is invalid")
		}
	}
	return nil
}
