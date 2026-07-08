package agentpod

import (
	"errors"
	"time"
)

const (
	CommandTypeCreatePod  = "create_pod"
	CommandTypeSendPrompt = "send_prompt"

	ErrCodeQueueExpired = "QUEUE_EXPIRED"
)

var (
	ErrQueueFull          = errors.New("pending command queue full")
	ErrDuplicateCommand   = errors.New("duplicate pending command")
)

type PendingCommand struct {
	ID             int64     `gorm:"primaryKey"`
	OrganizationID int64     `gorm:"not null"`
	RunnerID       int64     `gorm:"not null;index:idx_pending_cmds_runner_fifo,priority:1"`
	PodKey         string    `gorm:"size:100;not null"`
	CommandType    string    `gorm:"size:20;not null"`
	CommandID      string    `gorm:"size:64;not null;uniqueIndex:uq_pending_cmds_command"`
	Payload        []byte    `gorm:"not null"`
	ExpiresAt      time.Time `gorm:"not null;index:idx_pending_cmds_expiry"`
	CreatedAt      time.Time `gorm:"not null;default:now()"`
}

func (PendingCommand) TableName() string {
	return "pending_runner_commands"
}
