// Package skill holds the domain model for platform-authored, git-backed
// skills (namespace am-skills). Git is the source of truth for a skill's
// content (SKILL.md + skill.json in its repo); the AuthoredSkill row is a
// DB cache/index that backs List/Get and points at the packaged artifact in
// object storage. This coexists additively with the external-import /
// marketplace skill flow (domain/extension), which is untouched.
package skill

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when an authored skill row does not exist.
var ErrNotFound = errors.New("authored skill not found")

// AuthoredSkill is the DB cache row for a git-backed authored skill.
type AuthoredSkill struct {
	ID             int64  `gorm:"primaryKey" json:"id"`
	OrganizationID int64  `gorm:"column:organization_id;not null" json:"organization_id"`
	Slug           string `gorm:"size:100;not null" json:"slug"`
	DisplayName    string `gorm:"size:255;not null;default:''" json:"display_name"`
	Description    string `gorm:"not null;default:''" json:"description"`
	License        string `gorm:"size:100;not null;default:''" json:"license"`

	// GitRepoPath is "am-skills/org<ID>-<slug>"; the repo is authoritative.
	GitRepoPath   string  `gorm:"size:255;not null;column:git_repo_path" json:"git_repo_path"`
	DefaultBranch string  `gorm:"size:255;not null;default:main;column:default_branch" json:"default_branch"`
	HTTPCloneURL  *string `gorm:"size:1000;column:http_clone_url" json:"http_clone_url,omitempty"`

	// Packaged-artifact pointers, produced by the extension packager bridge.
	InstallSource string `gorm:"size:20;not null;default:gitops;column:install_source" json:"install_source"`
	ContentSha    string `gorm:"size:64;not null;default:'';column:content_sha" json:"content_sha"`
	StorageKey    string `gorm:"size:500;not null;default:'';column:storage_key" json:"storage_key"`
	PackageSize   int64  `gorm:"not null;default:0;column:package_size" json:"package_size"`
	Version       int    `gorm:"not null;default:1" json:"version"`

	CreatedByID int64     `gorm:"column:created_by_id;not null" json:"created_by_id"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (AuthoredSkill) TableName() string { return "authored_skills" }

// Repository owns the authored_skills DB cache rows.
type Repository interface {
	Create(ctx context.Context, s *AuthoredSkill) error
	Update(ctx context.Context, s *AuthoredSkill) error
	Delete(ctx context.Context, orgID, id int64) error
	GetByID(ctx context.Context, orgID, id int64) (*AuthoredSkill, error)
	GetBySlug(ctx context.Context, orgID int64, slug string) (*AuthoredSkill, error)
	SlugExists(ctx context.Context, orgID int64, slug string, excludeID int64) (bool, error)
	List(ctx context.Context, orgID int64, limit, offset int) ([]AuthoredSkill, int64, error)
}
