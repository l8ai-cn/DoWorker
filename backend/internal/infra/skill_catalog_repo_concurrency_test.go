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
