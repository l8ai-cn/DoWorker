package infra

import (
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
)

type agentWorkbenchStateRecord struct {
	SessionID      string    `gorm:"column:session_id;primaryKey"`
	StreamEpoch    string    `gorm:"column:stream_epoch"`
	Revision       uint64    `gorm:"column:revision;type:numeric(20,0)"`
	LatestSequence uint64    `gorm:"column:latest_sequence;type:numeric(20,0)"`
	Projection     []byte    `gorm:"column:projection"`
	Digest         string    `gorm:"column:digest"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (agentWorkbenchStateRecord) TableName() string {
	return "agent_workbench_session_states"
}

func (record agentWorkbenchStateRecord) domain() *agentworkbench.SessionState {
	return &agentworkbench.SessionState{
		SessionID: record.SessionID, StreamEpoch: record.StreamEpoch,
		Revision: record.Revision, LatestSequence: record.LatestSequence,
		Projection: append([]byte(nil), record.Projection...), Digest: record.Digest,
		CreatedAt: record.CreatedAt.UTC(), UpdatedAt: record.UpdatedAt.UTC(),
	}
}

func agentWorkbenchStateRecordFromDomain(
	state agentworkbench.SessionState,
	now time.Time,
) agentWorkbenchStateRecord {
	return agentWorkbenchStateRecord{
		SessionID: state.SessionID, StreamEpoch: state.StreamEpoch,
		Revision: state.Revision, LatestSequence: state.LatestSequence,
		Projection: append([]byte(nil), state.Projection...), Digest: state.Digest,
		CreatedAt: now, UpdatedAt: now,
	}
}

type agentWorkbenchEventRecord struct {
	SessionID          string    `gorm:"column:session_id;primaryKey"`
	StreamEpoch        string    `gorm:"column:stream_epoch;primaryKey"`
	Sequence           uint64    `gorm:"column:sequence;primaryKey;type:numeric(20,0)"`
	Revision           uint64    `gorm:"column:revision;type:numeric(20,0)"`
	Payload            []byte    `gorm:"column:payload"`
	Digest             string    `gorm:"column:digest"`
	CausationCommandID *string   `gorm:"column:causation_command_id"`
	CreatedAt          time.Time `gorm:"column:created_at"`
}

func (agentWorkbenchEventRecord) TableName() string {
	return "agent_workbench_events"
}

func (record agentWorkbenchEventRecord) domain() agentworkbench.Event {
	return agentworkbench.Event{
		SessionID: record.SessionID, StreamEpoch: record.StreamEpoch,
		Revision: record.Revision, Sequence: record.Sequence,
		Payload: append([]byte(nil), record.Payload...), Digest: record.Digest,
		CausationCommandID: record.CausationCommandID,
		CreatedAt:          record.CreatedAt.UTC(),
	}
}

func agentWorkbenchEventRecordFromDomain(
	event agentworkbench.Event,
) agentWorkbenchEventRecord {
	return agentWorkbenchEventRecord{
		SessionID: event.SessionID, StreamEpoch: event.StreamEpoch,
		Revision: event.Revision, Sequence: event.Sequence,
		Payload: append([]byte(nil), event.Payload...), Digest: event.Digest,
		CausationCommandID: event.CausationCommandID,
		CreatedAt:          event.CreatedAt.UTC(),
	}
}
