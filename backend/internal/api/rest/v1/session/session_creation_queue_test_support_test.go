package sessionapi

import (
	"testing"
	"time"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type sqlitePendingCommand struct {
	ID             int64     `gorm:"primaryKey"`
	OrganizationID int64     `gorm:"not null"`
	RunnerID       int64     `gorm:"not null;index:idx_pending_cmds_runner_fifo,priority:1"`
	PodKey         string    `gorm:"size:100;not null"`
	CommandType    string    `gorm:"size:20;not null"`
	CommandID      string    `gorm:"size:64;not null;uniqueIndex:uq_pending_cmds_command"`
	Payload        []byte    `gorm:"not null"`
	ExpiresAt      time.Time `gorm:"not null;index:idx_pending_cmds_expiry"`
	CreatedAt      time.Time `gorm:"not null"`
}

func (sqlitePendingCommand) TableName() string {
	return "pending_runner_commands"
}

func pendingCommandCount(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&podDomain.PendingCommand{}).Count(&count).Error)
	return count
}

type recordingSessionDispatchQueue struct {
	ttl           time.Duration
	enabled       bool
	connected     bool
	maxPerRunner  int
	triggers      []int64
	beforeTrigger func(int64)
}

func (q *recordingSessionDispatchQueue) AllowsDurableCommand(_ int64) bool {
	return q.enabled || q.connected
}

func (q *recordingSessionDispatchQueue) MaxPerRunner() int {
	return q.maxPerRunner
}

func (q *recordingSessionDispatchQueue) SendPromptTTL() time.Duration {
	return q.ttl
}

func (q *recordingSessionDispatchQueue) TriggerDrain(runnerID int64) {
	if q.beforeTrigger != nil {
		q.beforeTrigger(runnerID)
	}
	q.triggers = append(q.triggers, runnerID)
}
