package agentpod

import (
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/agent"
	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/domain/runner"
	"github.com/anthropics/agentsmesh/backend/internal/domain/ticket"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
)

// Pod status constants
const (
	StatusInitializing = "initializing"
	StatusRunning      = "running"
	StatusPaused       = "paused"
	StatusDisconnected = "disconnected" // User closed browser
	StatusOrphaned     = "orphaned"     // Lost due to runner restart
	StatusCompleted    = "completed"
	StatusTerminated   = "terminated"
	StatusError        = "error"
	StatusQueued       = "queued"
)

// Agent status constants
const (
	AgentStatusExecuting = "executing"
	AgentStatusWaiting   = "waiting"
	AgentStatusIdle      = "idle"
)

// Permission mode for Claude Code (maps to --permission-mode flag)
const (
	PermissionModeDefault     = "default"
	PermissionModePlan        = "plan"
	PermissionModeAcceptEdits = "acceptEdits"
	PermissionModeDontAsk     = "dontAsk"
	PermissionModeBypass      = "bypassPermissions"
)

// Interaction mode constants
const (
	InteractionModePTY = "pty"
	InteractionModeACP = "acp"
)

// AutomationLevel is the unified, cross-agent worker automation/permission
// tier configured at creation. A per-agent adapter translates it into that
// agent's native permission mechanism (permission_mode / approval_mode / ...).
const (
	// AutomationLevelInteractive — human-in-the-loop, approval on every step.
	AutomationLevelInteractive = "interactive"
	// AutomationLevelAutoEdit — auto-approve read/edit, ask only on risky ops.
	AutomationLevelAutoEdit = "auto_edit"
	// AutomationLevelAutonomous — fully autonomous, non-interactive; forces ACP
	// (a PTY cannot be programmatically approved).
	AutomationLevelAutonomous = "autonomous"

	// AutomationLevelDefault is applied when the level is empty/unknown so every
	// Worker is automatable by default.
	AutomationLevelDefault = AutomationLevelAutonomous
)

// NormalizeAutomationLevel maps empty/unknown values to the default so callers
// never persist an invalid tier.
func NormalizeAutomationLevel(level string) string {
	switch level {
	case AutomationLevelInteractive, AutomationLevelAutoEdit, AutomationLevelAutonomous:
		return level
	default:
		return AutomationLevelDefault
	}
}

// Pod represents an AI coding pod (AgentPod instance)
type Pod struct {
	ID             int64 `gorm:"primaryKey" json:"id"`
	OrganizationID int64 `gorm:"not null;index" json:"organization_id"`

	PodKey   string `gorm:"size:100;not null;uniqueIndex" json:"pod_key"`
	RunnerID int64  `gorm:"not null;index" json:"runner_id"`

	AgentSlug string `gorm:"size:100;column:agent_slug" json:"agent_slug,omitempty"`

	RepositoryID *int64 `json:"repository_id,omitempty"`
	TicketID     *int64 `json:"ticket_id,omitempty"`
	CreatedByID  int64  `gorm:"not null" json:"created_by_id"`

	TerminalPID *int   `gorm:"column:pty_pid" json:"pty_pid,omitempty"`
	Status      string `gorm:"size:50;not null;default:'initializing';index" json:"status"`
	AgentStatus string `gorm:"size:50;not null;default:'idle'" json:"agent_status"`
	AgentPID    *int   `gorm:"column:agent_pid" json:"agent_pid,omitempty"` // Claude/Agent process PID

	StartedAt         *time.Time `json:"started_at,omitempty"`
	FinishedAt        *time.Time `json:"finished_at,omitempty"`
	LastActivity      *time.Time `json:"last_activity,omitempty"`
	AgentWaitingSince *time.Time `json:"-"`

	// Prompt and configuration
	Prompt      string  `gorm:"column:prompt;type:text" json:"prompt,omitempty"`
	BranchName  *string `gorm:"size:255" json:"branch_name,omitempty"`
	SandboxPath *string `gorm:"column:sandbox_path;size:500" json:"sandbox_path,omitempty"`

	// Agent configuration used for this pod
	Model           *string `gorm:"size:50" json:"model,omitempty"`           // opus/sonnet/haiku
	PermissionMode  *string `gorm:"size:50" json:"permission_mode,omitempty"` // default/plan/acceptEdits/dontAsk/bypassPermissions
	InteractionMode string  `gorm:"column:interaction_mode;type:varchar(10);default:pty;not null" json:"interaction_mode"`
	// AutomationLevel is the unified permission/automation tier this pod was
	// created with (interactive/auto_edit/autonomous). The per-agent adapter
	// already translated it into agentfile CONFIG/MODE at create time; this
	// column preserves the requested tier for display and resume.
	AutomationLevel string `gorm:"column:automation_level;type:varchar(20);default:autonomous;not null" json:"automation_level"`
	// Error details from Runner (e.g., git clone auth failure)
	ErrorCode    *string `gorm:"size:100" json:"error_code,omitempty"`
	ErrorMessage *string `gorm:"type:text" json:"error_message,omitempty"`

	// Terminal title from OSC 0/2 escape sequences
	Title *string `gorm:"size:255" json:"title,omitempty"`

	// User-assigned alias for display purposes
	Alias *string `gorm:"size:100" json:"alias,omitempty"`

	// Session ID for agent session management (e.g., Claude Code --session-id)
	// Used for resume functionality - allows agents to restore conversation context
	SessionID *string `gorm:"size:36" json:"session_id,omitempty"`

	ExternalSessionID *string `gorm:"size:200" json:"external_session_id,omitempty"`

	// SourcePodKey tracks the original pod when this pod was created via resume
	// Enables tracking the chain of resumed sessions
	SourcePodKey *string `gorm:"size:100" json:"source_pod_key,omitempty"`

	VirtualAPIKeyID      *int64 `gorm:"column:virtual_api_key_id" json:"virtual_api_key_id,omitempty"`
	ModelResourceID      *int64 `gorm:"column:model_resource_id" json:"model_resource_id,omitempty"`
	WorkerSpecSnapshotID *int64 `gorm:"column:worker_spec_snapshot_id" json:"worker_spec_snapshot_id,omitempty"`

	PodLifecycleMetadata `gorm:"embedded"`

	// Preview: HTTP data-plane exposure of a loopback service inside the pod.
	// PreviewPort==0 means preview is disabled. PreviewPath is an optional
	// default sub-path the preview entrypoint lands on.
	PreviewPort int    `gorm:"column:preview_port;not null;default:0" json:"preview_port"`
	PreviewPath string `gorm:"column:preview_path;size:255;not null;default:''" json:"preview_path,omitempty"`

	// Perpetual mode: Runner auto-restarts the agent process on clean exit.
	// pod_key stays the same across restarts (service identity).
	Perpetual     bool       `gorm:"not null;default:false" json:"perpetual"`
	RestartCount  int        `gorm:"not null;default:0" json:"restart_count"`
	LastRestartAt *time.Time `json:"last_restart_at,omitempty"`

	// ResolvedConfig is the resolved pod-level AgentFile CONFIG snapshot
	// (model / permission_mode / approval_mode / ...) captured at pod create
	// time. Resume reads it to preserve the source pod's runtime configuration.
	// DB column and JSON tag remain "config_overrides" for back-compat.
	ResolvedConfig agent.ConfigValues `gorm:"column:config_overrides;type:jsonb;default:'{}'" json:"config_overrides,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`

	// Associations
	Runner     *runner.Runner          `gorm:"foreignKey:RunnerID" json:"runner,omitempty"`
	Agent      *agent.Agent            `gorm:"foreignKey:AgentSlug;references:Slug" json:"agent,omitempty"`
	Repository *gitprovider.Repository `gorm:"foreignKey:RepositoryID" json:"repository,omitempty"`
	Ticket     *ticket.Ticket          `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	CreatedBy  *user.User              `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	// Virtual field: populated by service layer via loop_runs join, not a DB column
	Loop *PodLoopInfo `gorm:"-" json:"loop,omitempty"`
}

// PodLoopInfo contains minimal loop information for pod display.
type PodLoopInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (Pod) TableName() string {
	return "pods"
}

// IsActive returns true if pod is active
func (p *Pod) IsActive() bool {
	return IsPodStatusActive(p.Status)
}

// IsTerminal returns true if pod is in a terminal state
func (p *Pod) IsTerminal() bool {
	return IsPodStatusTerminal(p.Status)
}

// CanReconnect returns true if pod can be reconnected
func (p *Pod) CanReconnect() bool {
	return p.Status == StatusDisconnected
}

// IsACPMode returns true if the pod uses ACP interaction mode.
func (p *Pod) IsACPMode() bool {
	return p.InteractionMode == InteractionModeACP
}

// GetOrganizationID returns the organization ID (implements middleware.PodGetter)
func (p *Pod) GetOrganizationID() int64 {
	return p.OrganizationID
}

// GetPodKey returns the pod key (implements middleware.PodGetter)
func (p *Pod) GetPodKey() string {
	return p.PodKey
}
