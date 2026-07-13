package skill

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

type conflictingSkillStore struct {
	*fakeStore
	syncAtCAS chan struct{}
	allowCAS  chan struct{}
}

func (s *conflictingSkillStore) UpdateIfVersion(
	ctx context.Context,
	row *skilldom.Skill,
	expectedVersion int,
) (bool, error) {
	if expectedVersion == 1 && assert.ObjectsAreEqual([]string{"initial"}, []string(row.Tags)) {
		close(s.syncAtCAS)
		<-s.allowCAS
		return false, nil
	}
	return s.fakeStore.UpdateIfVersion(ctx, row, expectedVersion)
}

func TestTagUpdateAndUpstreamSyncRetryToConsistentGitDBAndPackage(t *testing.T) {
	baseStore := newFakeStore()
	store := &conflictingSkillStore{
		fakeStore: baseStore,
		syncAtCAS: make(chan struct{}),
		allowCAS:  make(chan struct{}),
	}
	internalGit := gitops.NewFake("am-skills")
	packager := &fakePackager{}
	authorSvc := newTestService(store, internalGit, packager)
	syncSvc := newTestService(store, internalGit, packager)

	row, err := authorSvc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		Name:           "Video Editing",
		Instructions:   "# Video Editing",
		Tags:           []string{"initial"},
	})
	require.NoError(t, err)

	upstreamSkillMD := []byte("---\nname: video-editing\n---\n# Upstream")
	upstreamSkillJSON := []byte(`{"schema":2,"slug":"video-editing","tags":["upstream"]}`)
	syncResult := make(chan error, 1)
	go func() {
		_, syncErr := syncSvc.refreshImportedSkill(
			context.Background(),
			row,
			&extensionsvc.ClonedSkillSource{CommitSha: "abcdef123456"},
			extensionsvc.SkillInfo{Slug: "video-editing"},
			[]gitops.FileChange{
				{Path: "SKILL.md", Content: upstreamSkillMD},
				{Path: "skill.json", Content: upstreamSkillJSON},
			},
		)
		syncResult <- syncErr
	}()

	select {
	case <-store.syncAtCAS:
	case <-time.After(time.Second):
		t.Fatal("upstream sync did not reach optimistic update")
	}

	updatedTags := []string{"curated"}
	_, err = authorSvc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7,
		SkillID:        row.ID,
		Tags:           &updatedTags,
	})
	require.NoError(t, err)
	close(store.allowCAS)
	require.NoError(t, <-syncResult)

	saved, err := baseStore.GetByID(context.Background(), 7, row.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"curated"}, []string(saved.Tags))
	assert.Equal(t, 3, saved.Version)
	repo := internalGit.Repos["org7-video-editing"]
	assert.Equal(t, upstreamSkillMD, repo.Files["SKILL.md"])
	assertSkillConfigTags(t, repo.Files["skill.json"], []string{"curated"})
	assertSkillConfigTags(t, []byte(packager.lastSkillCfg), []string{"curated"})
	sum := sha256.Sum256(append(repo.Files["SKILL.md"], repo.Files["skill.json"]...))
	assert.Equal(t, fmt.Sprintf("%x", sum), saved.ContentSha)
	assert.Equal(t, saved.ContentSha, packageContentSHA(packager))
}

func packageContentSHA(packager *fakePackager) string {
	sum := sha256.Sum256(append([]byte(packager.lastSkillMd), []byte(packager.lastSkillCfg)...))
	return fmt.Sprintf("%x", sum)
}
