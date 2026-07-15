package expert

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

var ErrNotFound = errors.New("expert not found")

const (
	InteractionModePTY = "pty"
	InteractionModeACP = "acp"
)

const (
	AutomationLevelInteractive = "interactive"
	AutomationLevelAutoEdit    = "auto_edit"
	AutomationLevelAutonomous  = "autonomous"
	AutomationLevelDefault     = AutomationLevelAutonomous
)

// NormalizeAutomationLevel maps empty/unknown values to the autonomous default
// so experts always launch automatable Workers unless downgraded.
func NormalizeAutomationLevel(level string) string {
	switch level {
	case AutomationLevelInteractive, AutomationLevelAutoEdit, AutomationLevelAutonomous:
		return level
	default:
		return AutomationLevelDefault
	}
}

type KnowledgeMount struct {
	Slug string `json:"slug"`
	Mode string `json:"mode,omitempty"`
}

type Expert struct {
	ID             int64   `gorm:"primaryKey" json:"id"`
	OrganizationID int64   `gorm:"not null;index" json:"organization_id"`
	Slug           string  `gorm:"size:100;not null;uniqueIndex:idx_experts_org_slug" json:"slug"`
	Name           string  `gorm:"size:255;not null" json:"name"`
	Description    *string `gorm:"type:text" json:"description,omitempty"`

	AgentSlug    string  `gorm:"size:100;not null;column:agent_slug" json:"agent_slug"`
	RunnerID     *int64  `json:"runner_id,omitempty"`
	RepositoryID *int64  `json:"repository_id,omitempty"`
	BranchName   *string `gorm:"size:255" json:"branch_name,omitempty"`

	Prompt          *string `gorm:"type:text" json:"prompt,omitempty"`
	InteractionMode string  `gorm:"size:20;not null;default:pty" json:"interaction_mode"`
	// AutomationLevel is the unified permission/automation tier this expert
	// launches its Workers with (interactive/auto_edit/autonomous).
	AutomationLevel string `gorm:"size:20;not null;default:autonomous;column:automation_level" json:"automation_level"`
	Perpetual       bool   `gorm:"not null;default:false" json:"perpetual"`

	UsedEnvBundles  pq.StringArray  `gorm:"type:text[];column:used_env_bundles;not null;default:'{}'" json:"used_env_bundles"`
	SkillSlugs      pq.StringArray  `gorm:"type:text[];column:skill_slugs;not null;default:'{}'" json:"skill_slugs"`
	KnowledgeMounts json.RawMessage `gorm:"type:jsonb;not null;default:'[]'" json:"knowledge_mounts"`
	ConfigOverrides json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"config_overrides"`
	AgentfileLayer  *string         `gorm:"type:text" json:"agentfile_layer,omitempty"`

	SourcePodKey                  *string `gorm:"size:100" json:"source_pod_key,omitempty"`
	WorkerSpecSnapshotID          *int64  `json:"worker_spec_snapshot_id,omitempty"`
	OrchestrationResourceID       *int64  `json:"orchestration_resource_id,omitempty"`
	OrchestrationResourceRevision *int64  `json:"orchestration_resource_revision,omitempty"`

	// Git-backing columns index editable metadata. Worker execution is bound to
	// WorkerSpecSnapshotID and never reconstructed from these cached fields.
	GitRepoPath   *string         `gorm:"size:255;column:git_repo_path" json:"git_repo_path,omitempty"`
	DefaultBranch string          `gorm:"size:255;not null;default:main;column:default_branch" json:"default_branch"`
	HTTPCloneURL  *string         `gorm:"size:1000;column:http_clone_url" json:"http_clone_url,omitempty"`
	Metadata      json.RawMessage `gorm:"type:jsonb;not null;default:'{}';column:metadata" json:"metadata"`

	CreatedByID int64      `gorm:"not null" json:"created_by_id"`
	RunCount    int        `gorm:"not null;default:0" json:"run_count"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (Expert) TableName() string { return "experts" }

func (e *Expert) BeforeSave(_ *gorm.DB) error {
	return slugkit.ValidateIdentifier("experts.slug", e.Slug)
}

func ParseKnowledgeMounts(raw json.RawMessage) []KnowledgeMount {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var mounts []KnowledgeMount
	if err := json.Unmarshal(raw, &mounts); err != nil {
		return nil
	}
	return mounts
}

type Repository interface {
	Create(ctx context.Context, expert *Expert) error
	Update(ctx context.Context, expert *Expert) error
	Delete(ctx context.Context, orgID, id int64) error
	GetByID(ctx context.Context, orgID, id int64) (*Expert, error)
	GetBySlug(ctx context.Context, orgID int64, slug string) (*Expert, error)
	SlugExists(ctx context.Context, orgID int64, slug string, excludeID int64) (bool, error)
	List(ctx context.Context, orgID int64, limit, offset int) ([]Expert, int64, error)
	RecordRun(ctx context.Context, orgID, id int64, at time.Time) error
}
