package infra

import (
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
)

type agentWorkbenchReceiptRecord struct {
	SessionID     string    `gorm:"column:session_id;primaryKey"`
	CommandID     string    `gorm:"column:command_id;primaryKey"`
	PayloadDigest string    `gorm:"column:payload_digest"`
	State         int16     `gorm:"column:state"`
	Receipt       []byte    `gorm:"column:receipt"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (agentWorkbenchReceiptRecord) TableName() string {
	return "agent_workbench_command_receipts"
}

func (record agentWorkbenchReceiptRecord) domain() *agentworkbench.CommandReceipt {
	return &agentworkbench.CommandReceipt{
		SessionID: record.SessionID, CommandID: record.CommandID,
		PayloadDigest: record.PayloadDigest,
		State:         agentworkbench.ReceiptState(record.State),
		Receipt:       append([]byte(nil), record.Receipt...),
		CreatedAt:     record.CreatedAt.UTC(),
		UpdatedAt:     record.UpdatedAt.UTC(),
	}
}

func agentWorkbenchReceiptRecordFromDomain(
	receipt agentworkbench.CommandReceipt,
	now time.Time,
) agentWorkbenchReceiptRecord {
	return agentWorkbenchReceiptRecord{
		SessionID: receipt.SessionID, CommandID: receipt.CommandID,
		PayloadDigest: receipt.PayloadDigest, State: int16(receipt.State),
		Receipt:   append([]byte(nil), receipt.Receipt...),
		CreatedAt: now, UpdatedAt: now,
	}
}
