package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	extensionsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/extension"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/gitops"
)

func TestPreserveCuratorTags_ReplacesUpstreamTags(t *testing.T) {
	files := []gitops.FileChange{
		{
			Path:    "skill.json",
			Content: []byte(`{"schema":2,"slug":"video-editing","tags":["upstream"]}`),
		},
	}

	synced, err := preserveCuratorTags(files, "video-editing", []string{" Curated ", "VIDEO", "curated"})
	require.NoError(t, err)
	require.Len(t, synced, 1)

	var cfg skillConfig
	require.NoError(t, json.Unmarshal(synced[0].Content, &cfg))
	assert.Equal(t, 2, cfg.Schema)
	assert.Equal(t, []string{"curated", "video"}, cfg.Tags)
}

func TestPreserveCuratorTags_CreatesMissingSkillConfig(t *testing.T) {
	synced, err := preserveCuratorTags(
		[]gitops.FileChange{{Path: "SKILL.md", Content: []byte("---\nname: video-editing\n---\n")}},
		"video-editing",
		[]string{"video"},
	)
	require.NoError(t, err)
	require.Len(t, synced, 2)
	assert.Equal(t, "skill.json", synced[1].Path)
	assert.JSONEq(t, `{"schema":2,"slug":"video-editing","tags":["video"]}`, string(synced[1].Content))
}

func TestPreserveCuratorTags_PreservesUnknownLargeInteger(t *testing.T) {
	files := []gitops.FileChange{{
		Path: "skill.json",
		Content: []byte(
			`{"schema":2,"slug":"video-editing","tags":["upstream"],"future_number":9007199254740993}`,
		),
	}}

	synced, err := preserveCuratorTags(files, "video-editing", []string{"curated"})
	require.NoError(t, err)

	var config map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(synced[0].Content, &config))
	assert.Equal(t, "9007199254740993", string(config["future_number"]))
}

func TestRefreshImportedSkill_PreservesCatalogTags(t *testing.T) {
	dir := t.TempDir()
	skillMD := []byte("---\nname: video-editing\n---\n")
	skillJSON := []byte(`{"schema":2,"slug":"video-editing","tags":["upstream"]}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), skillMD, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.json"), skillJSON, 0644))

	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	repo, err := fake.Provision(context.Background(), gitops.ProvisionParams{
		OrgID: 7,
		Slug:  "video-editing",
		Seed:  []gitops.FileChange{{Path: "SKILL.md", Content: skillMD}},
	})
	require.NoError(t, err)

	orgID := int64(7)
	row := &skilldom.Skill{
		ID:             1,
		OrganizationID: &orgID,
		Slug:           "video-editing",
		Tags:           skilldom.NormalizeTags([]string{"curated", "video"}),
		GitRepoPath:    repo.Path,
		DefaultBranch:  repo.DefaultBranch,
	}
	store.rows[row.ID] = row
	svc := newTestService(store, fake, &fakePackager{})

	updated, err := svc.refreshImportedSkill(
		context.Background(),
		row,
		&extensionsvc.ClonedSkillSource{CommitSha: "abcdef123456"},
		extensionsvc.SkillInfo{
			Slug:    "video-editing",
			Tags:    []string{"upstream"},
			DirPath: dir,
		},
		[]gitops.FileChange{
			{Path: "SKILL.md", Content: skillMD},
			{Path: "skill.json", Content: skillJSON},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"curated", "video"}, []string(updated.Tags))
}

func TestExplicitUpstreamSync_PreservesCuratorTagsInGitAndPackage(t *testing.T) {
	upstream := createTagUpstream(t, []string{"upstream"})
	store := newFakeStore()
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
	row.Tags = skilldom.NormalizeTags([]string{"curated", "video"})
	store.rows[row.ID].Tags = row.Tags
	replaceUpstreamTags(t, upstream, []string{"upstream-new"})

	infos, err := extensionsvc.ScanSkillSource(upstream, "")
	require.NoError(t, err)
	files, err := readSkillDirFiles(infos[0].DirPath)
	require.NoError(t, err)
	row, err = svc.refreshImportedSkill(
		context.Background(),
		row,
		&extensionsvc.ClonedSkillSource{Dir: upstream, CommitSha: "fedcba654321"},
		infos[0],
		files,
	)
	require.NoError(t, err)
	assert.Equal(t, []string{"curated", "video"}, []string(row.Tags))
	assert.Contains(t, row.StorageKey, "skills/catalog/")
	assert.Equal(t, []string{
		"am-skills/org7-video-editing",
		"am-skills/org7-video-editing",
	}, packager.catalogIdentities)
	assertSkillConfigTags(t, internalGit.Repos["org7-video-editing"].Files["skill.json"], []string{"curated", "video"})
	assertSkillConfigTags(t, []byte(packager.lastSkillCfg), []string{"curated", "video"})
}

func TestSyncFromUpstreamRejectsInvalidCuratorTags(t *testing.T) {
	upstream := createTagUpstream(t, []string{"upstream"})
	store := newFakeStore()
	internalGit := gitops.NewFake("am-skills")
	svc := newTestService(store, internalGit, &fakePackager{})
	row, err := importTagSkill(t, svc, upstream, &ImportFromGitRequest{
		OrganizationID: 7,
		UserID:         3,
		URL:            "https://example.test/video-editing.git",
	})
	require.NoError(t, err)
	row.Tags = make([]string, 21)
	for i := range row.Tags {
		row.Tags[i] = fmt.Sprintf("tag-%02d", i)
	}
	store.rows[row.ID].Tags = row.Tags

	_, err = svc.SyncFromUpstream(context.Background(), 7, row.Slug)

	assert.ErrorIs(t, err, ErrInvalidTags)
}

func TestUpstreamSyncDoesNotIncrementVersionWhenContentIsUnchanged(t *testing.T) {
	upstream := createTagUpstream(t, []string{"video"})
	store := newFakeStore()
	internalGit := gitops.NewFake("am-skills")
	svc := newTestService(store, internalGit, &fakePackager{})
	request := &ImportFromGitRequest{
		OrganizationID: 7,
		UserID:         3,
		URL:            "https://example.test/video-editing.git",
	}
	row, err := importTagSkill(t, svc, upstream, request)
	require.NoError(t, err)
	require.Equal(t, 1, row.Version)
	beforeContentSha := row.ContentSha

	infos, err := extensionsvc.ScanSkillSource(upstream, "")
	require.NoError(t, err)
	files, err := readSkillDirFiles(infos[0].DirPath)
	require.NoError(t, err)
	updated, err := svc.refreshImportedSkill(
		context.Background(),
		row,
		&extensionsvc.ClonedSkillSource{CommitSha: "fedcba654321"},
		infos[0],
		files,
	)
	require.NoError(t, err)

	assert.Equal(t, beforeContentSha, updated.ContentSha)
	assert.Equal(t, 1, updated.Version)
	assert.Equal(t, "fedcba654321", updated.UpstreamCommitSha)
}
