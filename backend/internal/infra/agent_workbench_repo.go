package infra

import (
	"context"
	"errors"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentworkbench"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ agentworkbench.PersistenceRepository = (*agentWorkbenchRepository)(nil)

type agentWorkbenchRepository struct {
	db *gorm.DB
}

func NewAgentWorkbenchRepository(db *gorm.DB) agentworkbench.PersistenceRepository {
	return &agentWorkbenchRepository{db: db}
}

func (repo *agentWorkbenchRepository) Append(
	ctx context.Context,
	request agentworkbench.AppendRequest,
) (agentworkbench.AppendResult, error) {
	if err := validateAgentWorkbenchAppend(request); err != nil {
		return agentworkbench.AppendResult{}, err
	}
	result := agentworkbench.AppendResult{}
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockAgentWorkbenchSession(tx, request.SessionID); err != nil {
			return err
		}
		replayed, err := inspectAgentWorkbenchSources(tx, request.Sources)
		if err != nil {
			return err
		}
		if replayed {
			return nil
		}
		current, err := loadAgentWorkbenchState(tx, request.SessionID)
		if err != nil {
			return err
		}
		if err := validateAgentWorkbenchCurrent(current, request); err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := insertAgentWorkbenchArtifactFiles(
			tx,
			request.ArtifactFiles,
		); err != nil {
			return err
		}
		for _, receipt := range request.Receipts {
			if _, err := putAgentWorkbenchReceiptTx(tx, receipt, now); err != nil {
				return err
			}
		}
		if err := insertAgentWorkbenchSources(tx, request.Sources, now); err != nil {
			return err
		}
		if err := insertAgentWorkbenchEvents(tx, request.Events); err != nil {
			return err
		}
		if err := persistAgentWorkbenchState(
			tx,
			current,
			request.ExpectedRevision,
			request.Projection,
			now,
		); err != nil {
			return err
		}
		result.Applied = true
		return nil
	})
	return result, err
}

func lockAgentWorkbenchSession(tx *gorm.DB, sessionID string) error {
	return tx.Exec(
		"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
		sessionID,
	).Error
}

func loadAgentWorkbenchState(
	tx *gorm.DB,
	sessionID string,
) (*agentWorkbenchStateRecord, error) {
	var record agentWorkbenchStateRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_id = ?", sessionID).
		Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func insertAgentWorkbenchEvents(
	tx *gorm.DB,
	events []agentworkbench.Event,
) error {
	records := make([]agentWorkbenchEventRecord, len(events))
	for index := range events {
		records[index] = agentWorkbenchEventRecordFromDomain(events[index])
	}
	if err := tx.Create(&records).Error; err != nil {
		if isUniqueViolation(err) {
			return agentworkbench.ErrEventConflict
		}
		return err
	}
	return nil
}

func persistAgentWorkbenchState(
	tx *gorm.DB,
	current *agentWorkbenchStateRecord,
	expectedRevision uint64,
	projection agentworkbench.SessionState,
	now time.Time,
) error {
	if current == nil {
		record := agentWorkbenchStateRecordFromDomain(projection, now)
		if err := tx.Create(&record).Error; err != nil {
			if isUniqueViolation(err) {
				return agentworkbench.ErrRevisionConflict
			}
			return err
		}
		return nil
	}
	result := tx.Model(&agentWorkbenchStateRecord{}).
		Where("session_id = ? AND revision = ?", projection.SessionID, expectedRevision).
		Updates(map[string]any{
			"stream_epoch":    projection.StreamEpoch,
			"revision":        projection.Revision,
			"latest_sequence": projection.LatestSequence,
			"projection":      projection.Projection,
			"digest":          projection.Digest,
			"updated_at":      now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return agentworkbench.ErrRevisionConflict
	}
	return nil
}
