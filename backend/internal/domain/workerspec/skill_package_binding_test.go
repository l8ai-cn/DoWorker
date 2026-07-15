package workerspec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerSpecNormalizesAndValidatesSkillPackageBindings(t *testing.T) {
	spec := validWorkerSpec()
	spec.Workspace.SkillIDs = []int64{9, 3}
	spec.Workspace.SkillPackages = []SkillPackageBinding{
		{
			SkillID:     9,
			Slug:        " beta ",
			Version:     3,
			ContentSHA:  " sha-beta ",
			StorageKey:  " skills/beta.tar.gz ",
			PackageSize: 20,
		},
		{
			SkillID:     3,
			Slug:        "alpha",
			Version:     2,
			ContentSHA:  "sha-alpha",
			StorageKey:  "skills/alpha.tar.gz",
			PackageSize: 10,
		},
	}

	normalized, err := NormalizeAndValidate(spec)

	require.NoError(t, err)
	assert.Equal(t, []SkillPackageBinding{
		{
			SkillID:     3,
			Slug:        "alpha",
			Version:     2,
			ContentSHA:  "sha-alpha",
			StorageKey:  "skills/alpha.tar.gz",
			PackageSize: 10,
		},
		{
			SkillID:     9,
			Slug:        "beta",
			Version:     3,
			ContentSHA:  "sha-beta",
			StorageKey:  "skills/beta.tar.gz",
			PackageSize: 20,
		},
	}, normalized.Workspace.SkillPackages)
}

func TestWorkerSpecRejectsSkillPackageBindingMismatch(t *testing.T) {
	spec := validWorkerSpec()
	spec.Workspace.SkillIDs = []int64{3}
	spec.Workspace.SkillPackages = []SkillPackageBinding{{
		SkillID:    9,
		Slug:       "reviewer",
		Version:    1,
		ContentSHA: "sha-reviewer",
		StorageKey: "skills/reviewer.tar.gz",
	}}

	_, err := NormalizeAndValidate(spec)

	require.Error(t, err)
	assert.ErrorContains(t, err, "skill package")
}
