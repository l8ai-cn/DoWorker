package agentpod

import (
	"testing"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
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

	envBundleIDs, skillIDs, skillPackages := workerSpecResourceRequirements(spec)

	assert.Equal(t, []int64{4, 6}, envBundleIDs)
	assert.Equal(t, []int64{3, 9}, skillIDs)
	assert.Equal(t, spec.Workspace.SkillPackages, skillPackages)
}
