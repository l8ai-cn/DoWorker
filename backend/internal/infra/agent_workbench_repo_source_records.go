package infra

import (
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
)

type agentWorkbenchSourceEventRecord struct {
	SessionID          string    `gorm:"column:session_id;primaryKey"`
	StableEventID      string    `gorm:"column:stable_event_id;primaryKey"`
	RunnerSessionEpoch string    `gorm:"column:runner_session_epoch"`
	SourceSequence     uint64    `gorm:"column:source_sequence;type:numeric(20,0)"`
	PayloadDigest      string    `gorm:"column:payload_digest"`
	CreatedAt          time.Time `gorm:"column:created_at"`
}

func (agentWorkbenchSourceEventRecord) TableName() string {
	return "agent_workbench_source_events"
}

func agentWorkbenchSourceEventRecordFromDomain(
	source agentworkbench.SourceEvent,
	now time.Time,
) agentWorkbenchSourceEventRecord {
	return agentWorkbenchSourceEventRecord{
		SessionID: source.SessionID, StableEventID: source.StableEventID,
		RunnerSessionEpoch: source.RunnerSessionEpoch,
		SourceSequence:     source.SourceSequence,
		PayloadDigest:      source.PayloadDigest,
		CreatedAt:          now,
	}
}
