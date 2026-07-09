package extension

import (
	"testing"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

// ---------------------------------------------------------------------------
// InstalledSkill.GetEffectiveSha / GetEffectiveStorageKey / GetEffectivePackageSize
//
// Effective values live-follow the linked catalog row when the install is
// unpinned (PinnedVersion == nil) and the row is loaded (Skill != nil);
// otherwise the install's own snapshot wins.
// ---------------------------------------------------------------------------

func TestInstalledSkill_GetEffectiveSha(t *testing.T) {
	pinnedVersion := 5

	tests := []struct {
		name  string
		skill InstalledSkill
		want  string
	}{
		{
			name: "catalog_tracking_latest",
			skill: InstalledSkill{
				InstallSource: InstallSourceCatalog,
				PinnedVersion: nil,
				ContentSha:    "own-sha",
				Skill:         &skilldom.Skill{ContentSha: "catalog-sha"},
			},
			want: "catalog-sha",
		},
		{
			name: "catalog_pinned_version",
			skill: InstalledSkill{
				InstallSource: InstallSourceCatalog,
				PinnedVersion: &pinnedVersion,
				ContentSha:    "own-sha",
				Skill:         &skilldom.Skill{ContentSha: "catalog-sha"},
			},
			want: "own-sha",
		},
		{
			name: "catalog_no_row_loaded",
			skill: InstalledSkill{
				InstallSource: InstallSourceCatalog,
				PinnedVersion: nil,
				ContentSha:    "own-sha",
				Skill:         nil,
			},
			want: "own-sha",
		},
		{
			name: "github_source",
			skill: InstalledSkill{
				InstallSource: InstallSourceGitHub,
				ContentSha:    "github-sha",
			},
			want: "github-sha",
		},
		{
			name: "upload_source",
			skill: InstalledSkill{
				InstallSource: InstallSourceUpload,
				ContentSha:    "upload-sha",
			},
			want: "upload-sha",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.skill.GetEffectiveSha()
			if got != tt.want {
				t.Errorf("GetEffectiveSha() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInstalledSkill_GetEffectiveStorageKey(t *testing.T) {
	pinnedVersion := 5

	tests := []struct {
		name  string
		skill InstalledSkill
		want  string
	}{
		{
			name: "catalog_tracking_latest",
			skill: InstalledSkill{
				InstallSource: InstallSourceCatalog,
				PinnedVersion: nil,
				StorageKey:    "own-key",
				Skill:         &skilldom.Skill{StorageKey: "catalog-key"},
			},
			want: "catalog-key",
		},
		{
			name: "catalog_pinned_version",
			skill: InstalledSkill{
				InstallSource: InstallSourceCatalog,
				PinnedVersion: &pinnedVersion,
				StorageKey:    "own-key",
				Skill:         &skilldom.Skill{StorageKey: "catalog-key"},
			},
			want: "own-key",
		},
		{
			name: "catalog_no_row_loaded",
			skill: InstalledSkill{
				InstallSource: InstallSourceCatalog,
				PinnedVersion: nil,
				StorageKey:    "own-key",
				Skill:         nil,
			},
			want: "own-key",
		},
		{
			name: "github_source",
			skill: InstalledSkill{
				InstallSource: InstallSourceGitHub,
				StorageKey:    "github-key",
			},
			want: "github-key",
		},
		{
			name: "upload_source",
			skill: InstalledSkill{
				InstallSource: InstallSourceUpload,
				StorageKey:    "upload-key",
			},
			want: "upload-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.skill.GetEffectiveStorageKey()
			if got != tt.want {
				t.Errorf("GetEffectiveStorageKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInstalledSkill_GetEffectivePackageSize(t *testing.T) {
	pinnedVersion := 5

	tests := []struct {
		name  string
		skill InstalledSkill
		want  int64
	}{
		{
			name: "catalog_tracking_latest",
			skill: InstalledSkill{
				InstallSource: InstallSourceCatalog,
				PinnedVersion: nil,
				PackageSize:   100,
				Skill:         &skilldom.Skill{PackageSize: 999},
			},
			want: 999,
		},
		{
			name: "catalog_pinned_version",
			skill: InstalledSkill{
				InstallSource: InstallSourceCatalog,
				PinnedVersion: &pinnedVersion,
				PackageSize:   100,
				Skill:         &skilldom.Skill{PackageSize: 999},
			},
			want: 100,
		},
		{
			name: "catalog_no_row_loaded",
			skill: InstalledSkill{
				InstallSource: InstallSourceCatalog,
				PinnedVersion: nil,
				PackageSize:   100,
				Skill:         nil,
			},
			want: 100,
		},
		{
			name: "github_source",
			skill: InstalledSkill{
				InstallSource: InstallSourceGitHub,
				PackageSize:   200,
			},
			want: 200,
		},
		{
			name: "upload_source",
			skill: InstalledSkill{
				InstallSource: InstallSourceUpload,
				PackageSize:   300,
			},
			want: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.skill.GetEffectivePackageSize()
			if got != tt.want {
				t.Errorf("GetEffectivePackageSize() = %d, want %d", got, tt.want)
			}
		})
	}
}
