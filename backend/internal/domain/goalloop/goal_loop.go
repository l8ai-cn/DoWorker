package goalloop

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

var ErrNotFound = errors.New("goal loop not found")

const (
	StatusDraft     = "draft"
	StatusActive    = "active"
	StatusPaused    = "paused"
	StatusVerifying = "verifying"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

const (
	EscalationPause = "pause"
	EscalationFail  = "fail"
)

type GoalLoop struct {
	ID                          int64           `gorm:"primaryKey" json:"id"`
	OrganizationID              int64           `gorm:"not null;index" json:"organization_id"`
	CreatedByID                 int64           `gorm:"not null" json:"created_by_id"`
	Name                        string          `gorm:"size:255;not null" json:"name"`
	Slug                        string          `gorm:"size:100;not null;uniqueIndex:uq_goal_loops_organization_slug" json:"slug"`
	Description                 *string         `gorm:"type:text" json:"description,omitempty"`
	WorkerSpecSnapshotID        int64           `gorm:"not null" json:"worker_spec_snapshot_id"`
	Objective                   string          `gorm:"type:text;not null" json:"objective"`
	AcceptanceCriteria          json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"acceptance_criteria"`
	VerificationCommand         string          `gorm:"type:text;not null" json:"verification_command"`
	Status                      string          `gorm:"size:32;not null;default:'draft';index" json:"status"`
	PodKey                      *string         `gorm:"size:100" json:"pod_key,omitempty"`
	AutopilotControllerKey      *string         `gorm:"size:255" json:"autopilot_controller_key,omitempty"`
	MaxIterations               int             `gorm:"not null;default:10" json:"max_iterations"`
	CurrentIteration            int             `gorm:"not null;default:0" json:"current_iteration"`
	NoProgressCount             int             `gorm:"not null;default:0" json:"no_progress_count"`
	SameErrorCount              int             `gorm:"not null;default:0" json:"same_error_count"`
	LastProgressFingerprint     *string         `gorm:"size:64" json:"-"`
	LastErrorFingerprint        *string         `gorm:"size:64" json:"-"`
	RetryPromptCommandID        *string         `gorm:"size:64" json:"-"`
	RetryPromptCreatedAt        *time.Time      `json:"-"`
	TokenBudget                 *int64          `json:"token_budget,omitempty"`
	TimeoutMinutes              int             `gorm:"not null;default:60" json:"timeout_minutes"`
	NoProgressLimit             int             `gorm:"not null;default:3" json:"no_progress_limit"`
	SameErrorLimit              int             `gorm:"not null;default:2" json:"same_error_limit"`
	EscalationPolicy            string          `gorm:"size:20;not null;default:'pause'" json:"escalation_policy"`
	VerificationRequestID       *string         `gorm:"size:100" json:"verification_request_id,omitempty"`
	VerificationExitCode        *int            `json:"verification_exit_code,omitempty"`
	VerificationOutput          *string         `gorm:"type:text" json:"verification_output,omitempty"`
	VerificationOutputTruncated bool            `gorm:"not null;default:false" json:"verification_output_truncated"`
	VerificationError           *string         `gorm:"type:text" json:"verification_error,omitempty"`
	StartedAt                   *time.Time      `json:"started_at,omitempty"`
	VerifiedAt                  *time.Time      `json:"verified_at,omitempty"`
	CompletedAt                 *time.Time      `json:"completed_at,omitempty"`
	CreatedAt                   time.Time       `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt                   time.Time       `gorm:"not null;default:now()" json:"updated_at"`

	OrchestrationResourceID       *int64 `json:"orchestration_resource_id,omitempty"`
	OrchestrationResourceRevision *int64 `json:"orchestration_resource_revision,omitempty"`
}

func (GoalLoop) TableName() string {
	return "goal_loops"
}

func (l *GoalLoop) BeforeSave() error {
	return slugkit.ValidateIdentifier("goal_loops.slug", l.Slug)
}

func (l *GoalLoop) IsTerminal() bool {
	return l.Status == StatusCompleted || l.Status == StatusFailed || l.Status == StatusCancelled
}

func (l *GoalLoop) HasCompleteResourceBinding() bool {
	return l.OrchestrationResourceID != nil &&
		*l.OrchestrationResourceID > 0 &&
		l.OrchestrationResourceRevision != nil &&
		*l.OrchestrationResourceRevision > 0
}
