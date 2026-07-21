package coordinator

import "time"

// Execution status lifecycle. Ported from auto-harness coordinator ExecutionStatus
// but collapsed to the states Agent Cloud drives through PodOrchestrator.
const (
	ExecutionStatusPending        = "pending"
	ExecutionStatusClaimed        = "claimed"
	ExecutionStatusRunning        = "running"
	ExecutionStatusSucceeded      = "succeeded"
	ExecutionStatusFailed         = "failed"
	ExecutionStatusCancelled      = "cancelled"
	ExecutionStatusFeedbackFailed = "feedback_failed"
)

const (
	FeedbackStatusPending = "pending"
	FeedbackStatusPosted  = "posted"
	FeedbackStatusFailed  = "failed"
)

func IsTerminalStatus(status string) bool {
	switch status {
	case ExecutionStatusSucceeded, ExecutionStatusFailed, ExecutionStatusCancelled, ExecutionStatusFeedbackFailed:
		return true
	default:
		return false
	}
}

// Execution records one claim→dispatch→feedback cycle, linking a coordinator
// project to the Agent Cloud ticket it materialized and the pod that ran it.
type Execution struct {
	ID             int64 `gorm:"primaryKey" json:"id"`
	OrganizationID int64 `gorm:"not null;index" json:"organization_id"`
	ProjectID      int64 `gorm:"not null;index" json:"project_id"`
	TicketID       int64 `gorm:"not null;index" json:"ticket_id"`

	PodID  *int64  `json:"pod_id,omitempty"`
	PodKey *string `gorm:"size:100;index" json:"pod_key,omitempty"`

	Status string `gorm:"size:32;not null;default:'pending';index" json:"status"`
	Stage  string `gorm:"size:64" json:"stage,omitempty"`

	ClaimMarker string `gorm:"type:text" json:"claim_marker,omitempty"`
	ExternalID  string `gorm:"size:255;index" json:"external_id"`

	Summary        string `gorm:"type:text" json:"summary,omitempty"`
	FeedbackStatus string `gorm:"size:32" json:"feedback_status,omitempty"`
	Error          string `gorm:"type:text" json:"error,omitempty"`

	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (Execution) TableName() string { return "coordinator_executions" }
