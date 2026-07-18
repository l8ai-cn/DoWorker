package infra

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (repo *agentWorkbenchRepository) PutCommandReceipt(
	ctx context.Context,
	receipt agentworkbench.CommandReceipt,
) (*agentworkbench.CommandReceipt, error) {
	if err := validateAgentWorkbenchReceipt(receipt); err != nil {
		return nil, err
	}
	var stored *agentworkbench.CommandReceipt
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockAgentWorkbenchSession(tx, receipt.SessionID); err != nil {
			return err
		}
		var err error
		stored, err = putAgentWorkbenchReceiptTx(tx, receipt, time.Now().UTC())
		return err
	})
	return stored, err
}

func (repo *agentWorkbenchRepository) GetCommandReceipt(
	ctx context.Context,
	sessionID string,
	commandID string,
) (*agentworkbench.CommandReceipt, error) {
	if !validAgentWorkbenchText(sessionID, 100) ||
		!validAgentWorkbenchText(commandID, 100) {
		return nil, agentworkbench.ErrInvalidArgument
	}
	var record agentWorkbenchReceiptRecord
	err := repo.db.WithContext(ctx).
		Where("session_id = ? AND command_id = ?", sessionID, commandID).
		Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record.domain(), nil
}

func putAgentWorkbenchReceiptTx(
	tx *gorm.DB,
	receipt agentworkbench.CommandReceipt,
	now time.Time,
) (*agentworkbench.CommandReceipt, error) {
	current, err := loadAgentWorkbenchReceipt(tx, receipt.SessionID, receipt.CommandID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		record := agentWorkbenchReceiptRecordFromDomain(receipt, now)
		if err := tx.Create(&record).Error; err != nil {
			if isUniqueViolation(err) {
				return nil, agentworkbench.ErrReceiptConflict
			}
			return nil, err
		}
		return record.domain(), nil
	}
	if current.PayloadDigest != receipt.PayloadDigest {
		return nil, agentworkbench.ErrCommandIDConflict
	}
	nextState := int16(receipt.State)
	if current.State == nextState {
		if bytes.Equal(current.Receipt, receipt.Receipt) {
			return current.domain(), nil
		}
		if terminalAgentWorkbenchReceipt(receipt.State) {
			return nil, agentworkbench.ErrReceiptConflict
		}
	} else if !allowedAgentWorkbenchReceiptTransition(
		agentworkbench.ReceiptState(current.State),
		receipt.State,
	) {
		return nil, agentworkbench.ErrReceiptConflict
	}
	result := tx.Model(&agentWorkbenchReceiptRecord{}).
		Where("session_id = ? AND command_id = ?", receipt.SessionID, receipt.CommandID).
		Updates(map[string]any{
			"state":      nextState,
			"receipt":    receipt.Receipt,
			"updated_at": now,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, agentworkbench.ErrReceiptConflict
	}
	current.State = nextState
	current.Receipt = append([]byte(nil), receipt.Receipt...)
	current.UpdatedAt = now
	return current.domain(), nil
}

func loadAgentWorkbenchReceipt(
	tx *gorm.DB,
	sessionID string,
	commandID string,
) (*agentWorkbenchReceiptRecord, error) {
	var record agentWorkbenchReceiptRecord
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("session_id = ? AND command_id = ?", sessionID, commandID).
		Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func terminalAgentWorkbenchReceipt(state agentworkbench.ReceiptState) bool {
	switch state {
	case agentworkbench.ReceiptStateSucceeded,
		agentworkbench.ReceiptStateFailed,
		agentworkbench.ReceiptStateRejected,
		agentworkbench.ReceiptStateCancelled:
		return true
	default:
		return false
	}
}

func allowedAgentWorkbenchReceiptTransition(
	from agentworkbench.ReceiptState,
	to agentworkbench.ReceiptState,
) bool {
	switch from {
	case agentworkbench.ReceiptStateReceived:
		return to == agentworkbench.ReceiptStateAccepted ||
			to == agentworkbench.ReceiptStateFailed ||
			to == agentworkbench.ReceiptStateRejected
	case agentworkbench.ReceiptStateAccepted:
		return to == agentworkbench.ReceiptStateRunning ||
			to == agentworkbench.ReceiptStateSucceeded ||
			to == agentworkbench.ReceiptStateFailed ||
			to == agentworkbench.ReceiptStateCancelled
	case agentworkbench.ReceiptStateRunning:
		return to == agentworkbench.ReceiptStateSucceeded ||
			to == agentworkbench.ReceiptStateFailed ||
			to == agentworkbench.ReceiptStateCancelled
	default:
		return false
	}
}
