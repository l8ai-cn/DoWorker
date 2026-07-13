package skill

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

type alwaysConflictingSkillStore struct {
	*fakeStore
}

func (s *alwaysConflictingSkillStore) WithMutationLock(
	_ context.Context,
	_ int64,
	mutate func(skilldom.Repository) error,
) error {
	return mutate(s)
}

func (s *alwaysConflictingSkillStore) UpdateIfVersion(
	context.Context,
	*skilldom.Skill,
	int,
) (bool, error) {
	return false, nil
}

func TestUpdateRestoresGitWhenDatabaseUpdateFails(t *testing.T) {
	store := newFakeStore()
	internalGit := gitops.NewFake("am-skills")
	svc := newTestService(store, internalGit, &fakePackager{})
	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		Name:           "Video Editing",
		Instructions:   "Original body.",
		Tags:           []string{"initial"},
	})
	require.NoError(t, err)
	beforeRepo := cloneRepoFiles(internalGit.Repos["org7-video-editing"].Files)
	beforeRow := *row
	store.updateErr = errors.New("database unavailable")

	tags := []string{"curated"}
	_, err = svc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7,
		SkillID:        row.ID,
		Tags:           &tags,
	})
	require.ErrorContains(t, err, "database unavailable")

	assert.Equal(t, beforeRepo, internalGit.Repos["org7-video-editing"].Files)
	saved, getErr := store.GetByID(context.Background(), 7, row.ID)
	require.NoError(t, getErr)
	assert.Equal(t, beforeRow.ContentSha, saved.ContentSha)
	assert.Equal(t, beforeRow.StorageKey, saved.StorageKey)
	assert.Equal(t, beforeRow.PackageSize, saved.PackageSize)
	assert.Equal(t, beforeRow.Version, saved.Version)
}

func TestUpstreamSyncRestoresGitAfterContinuousConflicts(t *testing.T) {
	upstream := createTagUpstream(t, []string{"initial"})
	baseStore := newFakeStore()
	store := &alwaysConflictingSkillStore{fakeStore: baseStore}
	internalGit := gitops.NewFake("am-skills")
	svc := newTestService(store, internalGit, &fakePackager{})
	request := &ImportFromGitRequest{
		OrganizationID: 7,
		UserID:         3,
		URL:            "https://example.test/video-editing.git",
	}
	row, err := importTagSkill(t, svc, upstream, request)
	require.NoError(t, err)
	beforeRepo := cloneRepoFiles(internalGit.Repos["org7-video-editing"].Files)
	beforeRow := *row

	require.NoError(t, os.WriteFile(
		filepath.Join(upstream, "SKILL.md"),
		[]byte("---\nname: video-editing\n---\nChanged body.\n"),
		0o644,
	))
	require.NoError(t, os.WriteFile(filepath.Join(upstream, "new.txt"), []byte("new"), 0o644))
	infos, err := extensionsvc.ScanSkillSource(upstream, "")
	require.NoError(t, err)
	files, err := readSkillDirFiles(infos[0].DirPath)
	require.NoError(t, err)

	_, err = svc.refreshImportedSkill(
		context.Background(),
		row,
		&extensionsvc.ClonedSkillSource{CommitSha: "fedcba654321"},
		infos[0],
		files,
	)
	require.ErrorIs(t, err, ErrMutationConflict)

	assert.Equal(t, beforeRepo, internalGit.Repos["org7-video-editing"].Files)
	saved, getErr := baseStore.GetByID(context.Background(), 7, row.ID)
	require.NoError(t, getErr)
	assert.Equal(t, beforeRow.ContentSha, saved.ContentSha)
	assert.Equal(t, beforeRow.StorageKey, saved.StorageKey)
	assert.Equal(t, beforeRow.PackageSize, saved.PackageSize)
	assert.Equal(t, beforeRow.Version, saved.Version)
}

func cloneRepoFiles(files map[string][]byte) map[string][]byte {
	cloned := make(map[string][]byte, len(files))
	for path, content := range files {
		cloned[path] = append([]byte(nil), content...)
	}
	return cloned
}
