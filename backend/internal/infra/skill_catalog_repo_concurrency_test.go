package infra

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

func TestSkillCatalogRepositoryUpdateIfVersionRejectsStaleRow(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	require.NoError(t, db.Exec(
		"ALTER TABLE skills ADD COLUMN tags TEXT NOT NULL DEFAULT '{}'",
	).Error)
	repo := NewSkillCatalogRepository(db)
	orgID := int64(77)
	row := &skilldom.Skill{
		OrganizationID: &orgID,
		Slug:           "video-editing",
		DisplayName:    "Video Editing",
		Tags:           skilldom.NormalizeTags([]string{"video"}),
		GitRepoPath:    "am-skills/org77-video-editing",
		IsActive:       true,
		Version:        1,
	}
	require.NoError(t, repo.Create(ctx, row))

	first := *row
	first.DisplayName = "First Writer"
	first.Version = 2
	updated, err := repo.UpdateIfVersion(ctx, &first, 1)
	require.NoError(t, err)
	require.True(t, updated)

	stale := *row
	stale.DisplayName = "Stale Writer"
	stale.Version = 2
	updated, err = repo.UpdateIfVersion(ctx, &stale, 1)
	require.NoError(t, err)
	assert.False(t, updated)

	saved, err := repo.GetByID(ctx, orgID, row.ID)
	require.NoError(t, err)
	assert.Equal(t, "First Writer", saved.DisplayName)
	assert.Equal(t, 2, saved.Version)
}

func TestSkillCatalogRepositoryListAllIsOrgScopedAndIDOrdered(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	require.NoError(t, db.Exec(
		"ALTER TABLE skills ADD COLUMN tags TEXT NOT NULL DEFAULT '{}'",
	).Error)
	repo := NewSkillCatalogRepository(db)
	orgID := int64(77)
	otherOrgID := int64(88)
	for _, row := range []*skilldom.Skill{
		{ID: 9, OrganizationID: &orgID, Slug: "ninth", DisplayName: "Ninth", GitRepoPath: "skills/ninth"},
		{ID: 3, OrganizationID: &otherOrgID, Slug: "other", DisplayName: "Other", GitRepoPath: "skills/other"},
		{ID: 5, OrganizationID: &orgID, Slug: "fifth", DisplayName: "Fifth", GitRepoPath: "skills/fifth"},
	} {
		require.NoError(t, repo.Create(ctx, row))
	}

	rows, err := repo.ListAll(ctx, orgID)

	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, []int64{5, 9}, []int64{rows[0].ID, rows[1].ID})
}

func TestSkillCatalogRepositoryListsOnlyActivePlatformDependencies(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	require.NoError(t, db.Exec(
		"ALTER TABLE skills ADD COLUMN tags TEXT NOT NULL DEFAULT '{}'",
	).Error)
	repo := NewSkillCatalogRepository(db)
	orgID := int64(77)
	seeded := []*skilldom.Skill{
		{OrganizationID: nil, Slug: "active", GitRepoPath: "skills/active", IsActive: true},
		{OrganizationID: nil, Slug: "inactive", GitRepoPath: "skills/inactive", IsActive: true},
		{OrganizationID: &orgID, Slug: "org-only", GitRepoPath: "skills/org-only", IsActive: true},
		{OrganizationID: nil, Slug: "unrequested", GitRepoPath: "skills/unrequested", IsActive: true},
	}
	for _, row := range seeded {
		require.NoError(t, repo.Create(ctx, row))
	}
	require.NoError(t, db.Model(&skilldom.Skill{}).
		Where("slug = ?", "inactive").
		Update("is_active", false).Error)

	rows, err := repo.ListActivePlatformBySlugs(
		ctx,
		[]string{"org-only", "active", "inactive", "missing"},
	)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "active", rows[0].Slug)

	rows, err = repo.ListActivePlatformBySlugs(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, rows)

	rows, err = repo.ListByIDs(
		ctx,
		[]int64{seeded[2].ID, seeded[1].ID, seeded[0].ID, 9999},
	)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(
		t,
		[]int64{seeded[0].ID, seeded[1].ID, seeded[2].ID},
		[]int64{rows[0].ID, rows[1].ID, rows[2].ID},
	)
}
