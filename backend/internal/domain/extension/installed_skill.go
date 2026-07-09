package extension

import (
	"time"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

const (
	// InstallSourceCatalog marks installs sourced from the unified skill
	// catalog (skills table). Live-follows the catalog row unless pinned.
	InstallSourceCatalog = "catalog"
	InstallSourceGitHub  = "github"
	InstallSourceUpload  = "upload"
	// InstallSourceMarket is the legacy value written by the retired
	// registry/market pipeline. Rows keep working via their copied
	// content_sha/storage_key (no live-follow).
	InstallSourceMarket = "market"
)

type InstalledSkill struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	OrganizationID int64     `gorm:"not null" json:"organization_id"`
	RepositoryID   int64     `gorm:"not null" json:"repository_id"`
	SkillID        *int64    `json:"skill_id,omitempty"`
	Scope          string    `gorm:"size:20;not null" json:"scope"` // org / user
	InstalledBy    *int64    `json:"installed_by,omitempty"`
	Slug           string    `gorm:"size:100;not null" json:"slug"`
	InstallSource  string    `gorm:"size:20;not null" json:"install_source"` // catalog / github / upload
	SourceURL      string    `gorm:"size:500" json:"source_url,omitempty"`
	ContentSha     string    `gorm:"size:64" json:"content_sha,omitempty"`
	StorageKey     string    `gorm:"size:500" json:"storage_key,omitempty"`
	PackageSize    int64     `json:"package_size"`
	PinnedVersion  *int      `json:"pinned_version,omitempty"`
	IsEnabled      bool      `gorm:"not null;default:true" json:"is_enabled"`
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt      time.Time `gorm:"not null;default:now()" json:"updated_at"`

	Skill *skilldom.Skill `gorm:"foreignKey:SkillID" json:"skill,omitempty"`
}

func (InstalledSkill) TableName() string { return "installed_skills" }

// followsCatalog reports whether the install live-follows its catalog row
// (unpinned catalog install with the row preloaded).
func (s *InstalledSkill) followsCatalog() bool {
	return s.PinnedVersion == nil && s.Skill != nil
}

func (s *InstalledSkill) GetEffectiveSha() string {
	if s.followsCatalog() {
		return s.Skill.ContentSha
	}
	return s.ContentSha
}

func (s *InstalledSkill) GetEffectiveStorageKey() string {
	if s.followsCatalog() {
		return s.Skill.StorageKey
	}
	return s.StorageKey
}

func (s *InstalledSkill) GetEffectivePackageSize() int64 {
	if s.followsCatalog() {
		return s.Skill.PackageSize
	}
	return s.PackageSize
}
