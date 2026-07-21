package agentpod

import (
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	specdomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
)

func TestWorkerSpecResourceRequirementsUseExactIDs(t *testing.T) {
	spec := &specdomain.Spec{
		TypeConfig: specdomain.TypeConfig{
			SecretRefs: map[string]specdomain.SecretReference{
				"TOKEN": {
					Kind: slugkit.MustNewForTest("env-bundle"),
					ID:   6,
				},
			},
		},
		Workspace: specdomain.Workspace{
			SkillIDs: []int64{3, 9},
			SkillPackages: []specdomain.SkillPackageBinding{
				{SkillID: 3, Slug: "alpha", Version: 1, ContentSHA: "sha-a", StorageKey: "skills/a"},
				{SkillID: 9, Slug: "beta", Version: 2, ContentSHA: "sha-b", StorageKey: "skills/b"},
			},
			EnvBundleIDs: []specdomain.RuntimeEnvBundleID{4, 6},
		},
	}

	envBundleIDs, skillIDs, configBindings := workerSpecResourceRequirements(spec)

	assert.Equal(t, []int64{4, 6}, envBundleIDs)
	assert.Equal(t, []int64{3, 9}, skillIDs)
	assert.Empty(t, configBindings)
}

func TestWorkerSpecSecretEnvBundleIDsDeduplicateSecretRefs(t *testing.T) {
	spec := &specdomain.Spec{
		TypeConfig: specdomain.TypeConfig{
			SecretRefs: map[string]specdomain.SecretReference{
				"ACCESS_KEY": {
					Kind: slugkit.MustNewForTest("env-bundle"),
					ID:   6,
				},
				"SECRET_KEY": {
					Kind: slugkit.MustNewForTest("env-bundle"),
					ID:   6,
				},
			},
		},
		Workspace: specdomain.Workspace{
			EnvBundleIDs: []specdomain.RuntimeEnvBundleID{4},
		},
	}

	assert.Equal(t, []int64{6}, workerSpecSecretEnvBundleIDs(spec))
}

func TestArtifactSkillPackagesUsesBareContentSHAForRunnerCache(t *testing.T) {
	packages := artifactSkillPackages(&workerdependency.Document{
		Skills: []workerdependency.Skill{{
			Pin:           workerdependency.ResourcePin{DomainID: 3},
			Slug:          slugkit.MustNewForTest("canvas-compose"),
			Version:       1,
			ContentDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			StorageKey:    "skills/catalog/canvas-compose.tar.gz",
			PackageSize:   42,
		}},
	})

	assert.Equal(t,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		packages[0].ContentSHA,
	)
}
