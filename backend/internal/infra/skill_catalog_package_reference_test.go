package infra

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestSkillCatalogRepositoryDetectsCatalogAndHistoricalInstallReferences(t *testing.T) {
	ctx := context.Background()
	db := testkit.SetupTestDB(t)
	repo := NewSkillCatalogRepository(db)
	const storageKey = "skills/catalog/0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef/shared.tar.gz"

	referenced, err := repo.IsPackageReferenced(ctx, storageKey)
	require.NoError(t, err)
	require.False(t, referenced)

	require.NoError(t, db.Exec(
		`INSERT INTO skills (slug, git_repo_path, storage_key)
		 VALUES ('catalog-skill', 'am-skills/catalog-skill', ?)`,
		storageKey,
	).Error)
	referenced, err = repo.IsPackageReferenced(ctx, storageKey)
	require.NoError(t, err)
	require.True(t, referenced)

	require.NoError(t, db.Exec("DELETE FROM skills").Error)
	require.NoError(t, db.Exec(
		`INSERT INTO installed_skills
		 (organization_id, repository_id, scope, slug, install_source, storage_key)
		 VALUES (1, 2, 'org', 'historical-skill', 'market', ?)`,
		storageKey,
	).Error)
	referenced, err = repo.IsPackageReferenced(ctx, storageKey)
	require.NoError(t, err)
	require.True(t, referenced)
}
