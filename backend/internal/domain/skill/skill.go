// Package skill holds the domain model for the unified skill catalog.
// Every skill — platform-authored or imported from an external git repo —
// is backed by its own internal git repo (namespace am-skills); Git is the
// source of truth for content, the Skill row is the catalog index carrying
// the packaged-artifact pointers plus upstream provenance for imports.
package skill

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

// ErrNotFound is returned when a skill catalog row does not exist.
var ErrNotFound = errors.New("skill not found")

// install_source values for catalog rows.
const (
	SourceGitops = "gitops" // authored in-platform
	SourceImport = "import" // imported from an external git repo
)

// Skill is a unified catalog row for a git-backed skill.
type Skill struct {
	ID             int64          `gorm:"primaryKey" json:"id"`
	OrganizationID *int64         `gorm:"column:organization_id" json:"organization_id"` // NULL = platform-level
	Slug           string         `gorm:"size:100;not null" json:"slug"`
	DisplayName    string         `gorm:"size:255;not null;default:''" json:"display_name"`
	Description    string         `gorm:"not null;default:''" json:"description"`
	License        string         `gorm:"size:100;not null;default:''" json:"license"`
	Category       string         `gorm:"size:50;not null;default:''" json:"category,omitempty"`
	Compatibility  string         `gorm:"size:500;not null;default:''" json:"compatibility,omitempty"`
	AllowedTools   string         `gorm:"not null;default:''" json:"allowed_tools,omitempty"`
	Tags           pq.StringArray `gorm:"type:text[];not null;default:'{}'" json:"tags"`

	// AgentFilter whitelists agent slugs; empty means all agents.
	AgentFilter json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"agent_filter,omitempty"`
	IsActive    bool            `gorm:"not null;default:true" json:"is_active"`

	// GitRepoPath is "am-skills/org<ID>-<slug>"; the repo is authoritative.
	GitRepoPath   string  `gorm:"size:255;not null;column:git_repo_path" json:"git_repo_path"`
	DefaultBranch string  `gorm:"size:255;not null;default:main;column:default_branch" json:"default_branch"`
	HTTPCloneURL  *string `gorm:"size:1000;column:http_clone_url" json:"http_clone_url,omitempty"`

	// Upstream provenance — set for imported skills, empty for authored ones.
	UpstreamURL       string `gorm:"size:500;not null;default:'';column:upstream_url" json:"upstream_url,omitempty"`
	UpstreamSubdir    string `gorm:"size:255;not null;default:'';column:upstream_subdir" json:"upstream_subdir,omitempty"`
	UpstreamCommitSha string `gorm:"size:40;not null;default:'';column:upstream_commit_sha" json:"upstream_commit_sha,omitempty"`

	// Packaged-artifact pointers, produced by the extension packager bridge.
	InstallSource string `gorm:"size:20;not null;default:gitops;column:install_source" json:"install_source"`
	ContentSha    string `gorm:"size:64;not null;default:'';column:content_sha" json:"content_sha"`
	StorageKey    string `gorm:"size:500;not null;default:'';column:storage_key" json:"storage_key"`
	PackageSize   int64  `gorm:"not null;default:0;column:package_size" json:"package_size"`
	Version       int    `gorm:"not null;default:1" json:"version"`

	CreatedByID *int64    `gorm:"column:created_by_id" json:"created_by_id,omitempty"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (Skill) TableName() string { return "skills" }

func NormalizeTags(tags []string) pq.StringArray {
	unique := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" {
			unique[tag] = struct{}{}
		}
	}
	normalized := make(pq.StringArray, 0, len(unique))
	for tag := range unique {
		normalized = append(normalized, tag)
	}
	sort.Strings(normalized)
	return normalized
}

// IsPlatformLevel reports whether this skill is visible to every org.
func (s *Skill) IsPlatformLevel() bool { return s.OrganizationID == nil }

// GetAgentFilter parses agent_filter; nil means all agents are allowed.
func (s *Skill) GetAgentFilter() []string {
	if len(s.AgentFilter) == 0 {
		return nil
	}
	var filter []string
	if err := json.Unmarshal(s.AgentFilter, &filter); err != nil {
		return nil
	}
	return filter
}

// VisibleTo reports whether an org may see/install this skill.
func (s *Skill) VisibleTo(orgID int64) bool {
	return s.OrganizationID == nil || *s.OrganizationID == orgID
}

// Repository owns the skills catalog rows.
type Repository interface {
	Create(ctx context.Context, s *Skill) error
	Update(ctx context.Context, s *Skill) error
	UpdateIfVersion(ctx context.Context, s *Skill, expectedVersion int) (bool, error)
	WithMutationLock(ctx context.Context, id int64, mutate func(Repository) error) error
	WithPackageLock(ctx context.Context, storageKey string, mutate func(Repository) error) error
	IsPackageReferenced(ctx context.Context, storageKey string) (bool, error)
	Delete(ctx context.Context, orgID, id int64) error
	GetByID(ctx context.Context, orgID, id int64) (*Skill, error)
	// GetAnyByID fetches without org scoping — callers must check VisibleTo.
	GetAnyByID(ctx context.Context, id int64) (*Skill, error)
	GetBySlug(ctx context.Context, orgID int64, slug string) (*Skill, error)
	SlugExists(ctx context.Context, orgID int64, slug string, excludeID int64) (bool, error)
	FindByUpstream(ctx context.Context, orgID int64, upstreamURL, upstreamSubdir string) (*Skill, error)
	List(ctx context.Context, orgID int64, limit, offset int) ([]Skill, int64, error)
	// ListCatalog returns active org-level + platform-level skills for
	// marketplace browsing, optionally filtered by search query/category.
	ListCatalog(ctx context.Context, orgID int64, query, category string) ([]Skill, error)
}
