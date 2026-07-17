package agentworkbenchconnect

import (
	"crypto/sha256"
	"encoding/hex"

	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"google.golang.org/protobuf/proto"
)

func newSessionState(
	sessionID string,
	streamEpoch string,
) (workbenchdomain.SessionState, error) {
	snapshot := &agentworkbenchv2.SessionSnapshot{
		SessionId:   sessionID,
		StreamEpoch: streamEpoch,
		Status:      agentworkbenchv2.SessionStatus_SESSION_STATUS_IDLE,
	}
	digest, err := snapshotDigest(snapshot)
	if err != nil {
		return workbenchdomain.SessionState{}, err
	}
	snapshot.Digest = &digest
	projection, err := marshalDeterministic(snapshot)
	if err != nil {
		return workbenchdomain.SessionState{}, err
	}
	return workbenchdomain.SessionState{
		SessionID:   sessionID,
		StreamEpoch: streamEpoch,
		Projection:  projection,
		Digest:      digest,
	}, nil
}

func decodeSessionState(
	sessionID string,
	state *workbenchdomain.SessionState,
) (*agentworkbenchv2.SessionSnapshot, error) {
	if state == nil || len(state.Projection) == 0 {
		return nil, dataLoss("agent workbench snapshot is missing")
	}
	snapshot := &agentworkbenchv2.SessionSnapshot{}
	if err := proto.Unmarshal(state.Projection, snapshot); err != nil {
		return nil, dataLoss("agent workbench snapshot protobuf is invalid")
	}
	if state.SessionID != sessionID ||
		snapshot.GetSessionId() != state.SessionID ||
		snapshot.GetStreamEpoch() != state.StreamEpoch ||
		snapshot.GetRevision() != state.Revision ||
		snapshot.GetLatestSequence() != state.LatestSequence ||
		snapshot.GetDigest() != state.Digest {
		return nil, dataLoss("agent workbench snapshot metadata is inconsistent")
	}
	if snapshot.GetRevision() > snapshot.GetLatestSequence() ||
		(snapshot.GetRevision() == 0) != (snapshot.GetLatestSequence() == 0) {
		return nil, dataLoss("agent workbench snapshot position is invalid")
	}
	digest, err := snapshotDigest(snapshot)
	if err != nil {
		return nil, internalError(err)
	}
	if digest != state.Digest {
		return nil, dataLoss("agent workbench snapshot digest is invalid")
	}
	return snapshot, nil
}

func snapshotDigest(snapshot *agentworkbenchv2.SessionSnapshot) (string, error) {
	cloned := proto.Clone(snapshot).(*agentworkbenchv2.SessionSnapshot)
	cloned.Digest = nil
	return protoDigest(cloned)
}

func deltaDigest(delta *agentworkbenchv2.SessionDeltaBatch) (string, error) {
	cloned := proto.Clone(delta).(*agentworkbenchv2.SessionDeltaBatch)
	cloned.Digest = ""
	return protoDigest(cloned)
}

func protoDigest(message proto.Message) (string, error) {
	data, err := marshalDeterministic(message)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func marshalDeterministic(message proto.Message) ([]byte, error) {
	return (proto.MarshalOptions{Deterministic: true}).Marshal(message)
}
