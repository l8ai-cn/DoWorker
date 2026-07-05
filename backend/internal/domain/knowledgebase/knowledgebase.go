package knowledgebase

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("knowledgebase: not found")
	ErrSlugExists = errors.New("knowledgebase: slug already exists")
)

const (
	SourceTypeGit      = "git"
	SourceTypeFeishu   = "feishu"
	SourceTypeDingtalk = "dingtalk"
	SourceTypeGoogle   = "google"
)

const (
	SyncStatusIdle    = "idle"
	SyncStatusSyncing = "syncing"
	SyncStatusSynced  = "synced"
	SyncStatusFailed  = "failed"
)

const (
	MountModeReadOnly  = "ro"
	MountModeReadWrite = "rw"
)

func ValidSourceType(t string) bool {
	switch t {
	case SourceTypeGit, SourceTypeFeishu, SourceTypeDingtalk, SourceTypeGoogle:
		return true
	}
	return false
}

func ValidMountMode(m string) bool {
	return m == MountModeReadOnly || m == MountModeReadWrite
}

// KnowledgeBase is an org-scoped llm-wiki: one git repository laid out as
// llms.txt (index) + AGENTS.md (schema) + raw/ (immutable sources) + wiki/
// (LLM-maintained pages). Non-git source types sync one-way into raw/ so the
// pod mount pipeline stays git-only.
type KnowledgeBase struct {
	ID             int64 `gorm:"primaryKey" json:"id"`
	OrganizationID int64 `gorm:"not null;index" json:"organization_id"`

	Slug        string `gorm:"size:100;not null;uniqueIndex:idx_knowledge_bases_org_slug" json:"slug"`
	Name        string `gorm:"size:255;not null" json:"name"`
	Description string `gorm:"not null;default:''" json:"description"`

	GitRepoPath   string `gorm:"size:255;not null" json:"git_repo_path"`
	HTTPCloneURL  string `gorm:"column:http_clone_url;size:1000;not null" json:"http_clone_url"`
	DefaultBranch string `gorm:"size:255;not null;default:'main'" json:"default_branch"`

	SourceType   string          `gorm:"size:32;not null;default:'git'" json:"source_type"`
	SourceConfig json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"source_config"`

	SyncStatus   string     `gorm:"size:32;not null;default:'idle'" json:"sync_status"`
	SyncError    *string    `json:"sync_error,omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`

	CreatedByUserID int64 `gorm:"not null" json:"created_by_user_id"`

	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (KnowledgeBase) TableName() string { return "knowledge_bases" }

// AgentMount is a default KB→agent attachment applied at pod create time.
type AgentMount struct {
	ID              int64     `gorm:"primaryKey" json:"id"`
	OrganizationID  int64     `gorm:"not null;index" json:"organization_id"`
	KnowledgeBaseID int64     `gorm:"not null" json:"knowledge_base_id"`
	AgentSlug       string    `gorm:"size:100;not null" json:"agent_slug"`
	Mode            string    `gorm:"size:8;not null;default:'ro'" json:"mode"`
	CreatedAt       time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (AgentMount) TableName() string { return "knowledge_base_agent_mounts" }
