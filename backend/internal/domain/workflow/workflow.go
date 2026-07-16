package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrWorkflowDisabled = errors.New("workflow is disabled")
	ErrHasActiveRuns    = errors.New("workflow has active runs")
)

const (
	StatusEnabled  = "enabled"
	StatusDisabled = "disabled"
	StatusArchived = "archived"
)

const (
	ExecutionModeAutopilot = "autopilot"
	ExecutionModeDirect    = "direct"
)

const (
	SandboxStrategyPersistent = "persistent"
	SandboxStrategyFresh      = "fresh"
)

const (
	ConcurrencyPolicySkip    = "skip"
	ConcurrencyPolicyQueue   = "queue"
	ConcurrencyPolicyReplace = "replace"
)

type Workflow struct {
	ID             int64 `gorm:"primaryKey" json:"id"`
	OrganizationID int64 `gorm:"not null;index" json:"organization_id"`

	Name        string  `gorm:"size:255;not null" json:"name"`
	Slug        string  `gorm:"size:100;not null;uniqueIndex:idx_workflows_org_slug" json:"slug"`
	Description *string `gorm:"type:text" json:"description,omitempty"`

	AgentSlug      string `gorm:"size:100;column:agent_slug" json:"agent_slug,omitempty"`
	PermissionMode string `gorm:"size:50;not null;default:'bypassPermissions'" json:"permission_mode"`

	PromptTemplate string `gorm:"type:text;not null" json:"prompt_template"`

	RepositoryID    *int64  `json:"repository_id,omitempty"`
	RunnerID        *int64  `json:"runner_id,omitempty"`
	BranchName      *string `gorm:"size:255" json:"branch_name,omitempty"`
	TicketID        *int64  `json:"ticket_id,omitempty"`
	ModelResourceID *int64  `gorm:"column:model_resource_id" json:"model_resource_id,omitempty"`
	// UsedEnvBundles is the ordered list of EnvBundle names to attach to
	// every run. Each name becomes a `USE_ENV_BUNDLE "<name>"` line in the
	// generated AgentFile layer, emitted in array order; later entries
	// override earlier ones on conflicting env keys (mirrors Pod creation).
	// Empty slice = no runtime preferences. Names are stable across bundle
	// rename/recreate, unlike IDs.
	UsedEnvBundles pq.StringArray `gorm:"type:text[];column:used_env_bundles;not null;default:'{}'" json:"used_env_bundles"`

	ConfigOverrides json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"config_overrides"`

	PromptVariables json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"prompt_variables"`

	ExecutionMode  string  `gorm:"size:20;not null;default:'autopilot'" json:"execution_mode"`
	CronExpression *string `gorm:"size:100" json:"cron_expression,omitempty"`

	AutopilotConfig json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"autopilot_config"`

	CallbackURL *string `gorm:"size:500" json:"callback_url,omitempty"`

	Status             string `gorm:"size:20;not null;default:'enabled';index" json:"status"`
	SandboxStrategy    string `gorm:"size:20;not null;default:'persistent'" json:"sandbox_strategy"`
	SessionPersistence bool   `gorm:"not null;default:true" json:"session_persistence"`
	ConcurrencyPolicy  string `gorm:"size:20;not null;default:'skip'" json:"concurrency_policy"`
	MaxConcurrentRuns  int    `gorm:"not null;default:1" json:"max_concurrent_runs"`
	MaxRetainedRuns    int    `gorm:"not null;default:0" json:"max_retained_runs"` // 0 = unlimited
	TimeoutMinutes     int    `gorm:"not null;default:60" json:"timeout_minutes"`
	IdleTimeoutSec     int    `gorm:"not null;default:30" json:"idle_timeout_sec"`

	// SandboxPath is informational only — the actual sandbox resume mechanism
	// uses LastPodKey → Pod.sandbox_path chain via PodOrchestrator.CreatePod(source_pod_key).
	// This field is NOT required for Resume to work; it serves as a diagnostic reference.
	SandboxPath *string `gorm:"size:500" json:"sandbox_path,omitempty"`
	LastPodKey  *string `gorm:"size:100" json:"last_pod_key,omitempty"`

	OrchestrationResourceID       *int64 `json:"orchestration_resource_id,omitempty"`
	OrchestrationResourceRevision *int64 `json:"orchestration_resource_revision,omitempty"`
	WorkerSpecSnapshotID          *int64 `json:"worker_spec_snapshot_id,omitempty"`

	CreatedByID int64 `gorm:"not null" json:"created_by_id"`

	TotalRuns      int `gorm:"not null;default:0" json:"total_runs"`
	SuccessfulRuns int `gorm:"not null;default:0" json:"successful_runs"`
	FailedRuns     int `gorm:"not null;default:0" json:"failed_runs"`

	LastRunAt *time.Time `json:"last_run_at,omitempty"`
	NextRunAt *time.Time `json:"next_run_at,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	ActiveRunCount int      `json:"active_run_count" gorm:"-"`
	AvgDurationSec *float64 `json:"avg_duration_sec,omitempty" gorm:"-"`
}

func (Workflow) TableName() string {
	return "workflows"
}

// ListWorkflowsFilter represents filters for listing workflows
type ListWorkflowsFilter struct {
	OrganizationID int64
	Status         string
	ExecutionMode  string
	CronEnabled    *bool // true=has cron, false=no cron, nil=any
	Query          string
	Limit          int
	Offset         int
}

// WorkflowRepository defines the interface for workflow data access
type WorkflowRepository interface {
	Create(ctx context.Context, workflow *Workflow) error
	GetByID(ctx context.Context, id int64) (*Workflow, error)
	GetBySlug(ctx context.Context, orgID int64, slug string) (*Workflow, error)
	List(ctx context.Context, filter *ListWorkflowsFilter) ([]*Workflow, int64, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	Delete(ctx context.Context, orgID int64, slug string) (int64, error)
	// GetDueCronWorkflows returns enabled cron workflows that are due for execution.
	// orgIDs filters to specific organizations; nil means all orgs (single-instance mode).
	GetDueCronWorkflows(ctx context.Context, orgIDs []int64) ([]*Workflow, error)

	// ClaimCronWorkflow atomically claims a cron workflow with SKIP LOCKED and advances next_run_at.
	// Returns true if claimed, false if skipped or no longer due.
	ClaimCronWorkflow(ctx context.Context, workflowID int64, nextRunAt *time.Time) (bool, error)

	// FindWorkflowsNeedingNextRun returns enabled cron workflows with next_run_at IS NULL.
	// orgIDs filters to specific organizations; nil means all orgs (single-instance mode).
	FindWorkflowsNeedingNextRun(ctx context.Context, orgIDs []int64) ([]*Workflow, error)

	// IncrementRunStats atomically increments run statistics counters.
	IncrementRunStats(ctx context.Context, workflowID int64, status string, lastRunAt time.Time) error
}

// IsEnabled returns true if the workflow is active
func (l *Workflow) IsEnabled() bool {
	return l.Status == StatusEnabled
}

// HasCron returns true if the workflow has cron scheduling enabled
func (l *Workflow) HasCron() bool {
	return l.CronExpression != nil && *l.CronExpression != ""
}

// IsAutopilot returns true if the workflow uses autopilot execution mode
func (l *Workflow) IsAutopilot() bool {
	return l.ExecutionMode == ExecutionModeAutopilot
}

// IsPersistent returns true if sandbox is persistent across runs.
// When persistent, both sandbox files and agent session (conversation history) are preserved.
func (l *Workflow) IsPersistent() bool {
	return l.SandboxStrategy == SandboxStrategyPersistent
}

func (l *Workflow) IsResourceManaged() bool {
	return l.OrchestrationResourceID != nil ||
		l.OrchestrationResourceRevision != nil ||
		l.WorkerSpecSnapshotID != nil
}

// SuccessRate returns the success rate as a percentage (0-100)
func (l *Workflow) SuccessRate() float64 {
	if l.TotalRuns == 0 {
		return 0
	}
	return float64(l.SuccessfulRuns) / float64(l.TotalRuns) * 100
}
