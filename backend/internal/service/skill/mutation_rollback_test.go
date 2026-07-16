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

type installedReferenceSkillStore struct {
	*fakeStore
	checkedKeys []string
}

func (s *installedReferenceSkillStore) WithMutationLock(
	_ context.Context,
	_ int64,
	mutate func(skilldom.Repository) error,
) error {
	return mutate(s)
}

func (s *installedReferenceSkillStore) WithPackageLock(
	_ context.Context,
	_ string,
	mutate func(skilldom.Repository) error,
) error {
	return mutate(s)
}

func (s *installedReferenceSkillStore) IsPackageReferenced(
	_ context.Context,
	storageKey string,
) (bool, error) {
	s.checkedKeys = append(s.checkedKeys, storageKey)
	return true, nil
}

func (s *alwaysConflictingSkillStore) WithMutationLock(
	_ context.Context,
	_ int64,
	mutate func(skilldom.Repository) error,
) error {
	return mutate(s)
}

func (s *alwaysConflictingSkillStore) WithPackageLock(
	_ context.Context,
	_ string,
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

func TestUpdateRestoresGitAndCleansNewPackageWhenDatabaseUpdateFails(t *testing.T) {
	store := newFakeStore()
	internalGit := gitops.NewFake("am-skills")
	packager := &fakePackager{}
	svc := newTestService(store, internalGit, packager)
	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		Name:           "Video Editing",
		Instructions:   "Original body.",
		Tags:           []string{"initial"},
	})
	require.NoError(t, err)
	beforeRepo := cloneRepoFiles(internalGit.Repos["org7-video-editing"].Files)
	beforeRow := *row
	packager.deletedKeys = nil
	packager.deleteHook = func() {
		assert.Equal(t, beforeRepo, internalGit.Repos["org7-video-editing"].Files)
	}
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
	require.Len(t, packager.deletedKeys, 1)
	assert.NotEqual(t, beforeRow.StorageKey, packager.deletedKeys[0])
}

func TestUpstreamSyncRestoresGitAfterContinuousConflicts(t *testing.T) {
	upstream := createTagUpstream(t, []string{"initial"})
	baseStore := newFakeStore()
	store := &alwaysConflictingSkillStore{fakeStore: baseStore}
	internalGit := gitops.NewFake("am-skills")
	packager := &fakePackager{}
	svc := newTestService(store, internalGit, packager)
	request := &ImportFromGitRequest{
		OrganizationID: 7,
		UserID:         3,
		URL:            "https://example.test/video-editing.git",
	}
	row, err := importTagSkill(t, svc, upstream, request)
	require.NoError(t, err)
	beforeRepo := cloneRepoFiles(internalGit.Repos["org7-video-editing"].Files)
	beforeRow := *row
	packager.deletedKeys = nil

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
	require.Len(t, packager.deletedKeys, maxSkillMutationAttempts)
	for _, storageKey := range packager.deletedKeys {
		assert.NotEqual(t, beforeRow.StorageKey, storageKey)
	}
}

func TestUpdateReturnsDatabaseAndPackageCleanupErrors(t *testing.T) {
	store := newFakeStore()
	internalGit := gitops.NewFake("am-skills")
	packager := &fakePackager{}
	svc := newTestService(store, internalGit, packager)
	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		Name:           "Video Editing",
		Instructions:   "Original body.",
		Tags:           []string{"initial"},
	})
	require.NoError(t, err)
	beforeRepo := cloneRepoFiles(internalGit.Repos["org7-video-editing"].Files)
	packager.deletedKeys = nil
	store.updateErr = errors.New("database unavailable")
	packager.deleteErr = errors.New("object delete failed")

	tags := []string{"curated"}
	_, err = svc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7,
		SkillID:        row.ID,
		Tags:           &tags,
	})

	require.ErrorContains(t, err, "database unavailable")
	require.ErrorContains(t, err, "object delete failed")
	assert.Equal(t, beforeRepo, internalGit.Repos["org7-video-editing"].Files)
	require.Len(t, packager.deletedKeys, 1)
}

func TestUpdateKeepsNewPackageReferencedByHistoricalInstall(t *testing.T) {
	baseStore := newFakeStore()
	store := &installedReferenceSkillStore{fakeStore: baseStore}
	internalGit := gitops.NewFake("am-skills")
	packager := &fakePackager{}
	svc := newTestService(store, internalGit, packager)
	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		Name:           "Video Editing",
		Instructions:   "Original body.",
		Tags:           []string{"initial"},
	})
	require.NoError(t, err)
	packager.deletedKeys = nil
	store.updateErr = errors.New("database unavailable")

	tags := []string{"curated"}
	_, err = svc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7,
		SkillID:        row.ID,
		Tags:           &tags,
	})

	require.ErrorContains(t, err, "database unavailable")
	require.Len(t, store.checkedKeys, 1)
	assert.NotEqual(t, row.StorageKey, store.checkedKeys[0])
	assert.Empty(t, packager.deletedKeys)
}

func TestUpstreamDatabaseErrorKeepsReusedPackage(t *testing.T) {
	upstream := createTagUpstream(t, []string{"initial"})
	store := newFakeStore()
	internalGit := gitops.NewFake("am-skills")
	packager := &fakePackager{}
	svc := newTestService(store, internalGit, packager)
	row, err := importTagSkill(t, svc, upstream, &ImportFromGitRequest{
		OrganizationID: 7,
		UserID:         3,
		URL:            "https://example.test/video-editing.git",
	})
	require.NoError(t, err)
	packager.deletedKeys = nil
	packager.reused = true
	store.updateErr = errors.New("database unavailable")
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

	require.ErrorContains(t, err, "database unavailable")
	assert.Empty(t, packager.deletedKeys)
}

func cloneRepoFiles(files map[string][]byte) map[string][]byte {
	cloned := make(map[string][]byte, len(files))
	for path, content := range files {
		cloned[path] = append([]byte(nil), content...)
	}
	return cloned
}
