package infra

import (
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func inspectAgentWorkbenchSources(
	tx *gorm.DB,
	sources []agentworkbench.SourceEvent,
) (bool, error) {
	if len(sources) == 0 {
		return false, nil
	}
	existing := 0
	for _, source := range sources {
		records, err := findAgentWorkbenchSourceIdentity(tx, source)
		if err != nil {
			return false, err
		}
		if len(records) == 0 {
			continue
		}
		if len(records) != 1 || !sameAgentWorkbenchSource(records[0], source) {
			return false, agentworkbench.ErrSourceEventConflict
		}
		existing++
	}
	if existing == len(sources) {
		return true, nil
	}
	if existing != 0 {
		return false, agentworkbench.ErrSourceEventConflict
	}
	return false, nil
}

func findAgentWorkbenchSourceIdentity(
	tx *gorm.DB,
	source agentworkbench.SourceEvent,
) ([]agentWorkbenchSourceEventRecord, error) {
	var records []agentWorkbenchSourceEventRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			`session_id = ? AND (
				stable_event_id = ?
				OR (runner_session_epoch = ? AND source_sequence = ?)
			)`,
			source.SessionID,
			source.StableEventID,
			source.RunnerSessionEpoch,
			source.SourceSequence,
		).
		Find(&records).Error
	return records, err
}

func sameAgentWorkbenchSource(
	record agentWorkbenchSourceEventRecord,
	source agentworkbench.SourceEvent,
) bool {
	return record.SessionID == source.SessionID &&
		record.StableEventID == source.StableEventID &&
		record.RunnerSessionEpoch == source.RunnerSessionEpoch &&
		record.SourceSequence == source.SourceSequence &&
		record.PayloadDigest == source.PayloadDigest
}

func insertAgentWorkbenchSources(
	tx *gorm.DB,
	sources []agentworkbench.SourceEvent,
	now time.Time,
) error {
	if len(sources) == 0 {
		return nil
	}
	records := make([]agentWorkbenchSourceEventRecord, len(sources))
	for index := range sources {
		records[index] = agentWorkbenchSourceEventRecordFromDomain(sources[index], now)
	}
	if err := tx.Create(&records).Error; err != nil {
		if isUniqueViolation(err) {
			return agentworkbench.ErrSourceEventConflict
		}
		return err
	}
	return nil
}
