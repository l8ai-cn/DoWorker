package skill

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

func TestImportFromGit_InitialImportSynchronizesTags(t *testing.T) {
	upstream := createTagUpstream(t, []string{" Video ", "editing", "VIDEO"})
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

	assert.Equal(t, []string{"editing", "video"}, []string(row.Tags))
	assert.Contains(t, row.StorageKey, "skills/catalog/")
	assert.Equal(t, []string{"am-skills/org7-video-editing"}, packager.catalogIdentities)
	assertSkillConfigTags(t, internalGit.Repos["org7-video-editing"].Files["skill.json"], []string{"editing", "video"})
	assertSkillConfigTags(t, []byte(packager.lastSkillCfg), []string{"editing", "video"})
}

func TestImportFromGit_ReimportPreservesCuratorTagsEverywhere(t *testing.T) {
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

	row, err = importTagSkill(t, svc, upstream, request)
	require.NoError(t, err)

	assert.Equal(t, []string{"curated", "video"}, []string(row.Tags))
	assertSkillConfigTags(t, internalGit.Repos["org7-video-editing"].Files["skill.json"], []string{"curated", "video"})
	assertSkillConfigTags(t, []byte(packager.lastSkillCfg), []string{"curated", "video"})
}

func createTagUpstream(t *testing.T, tags []string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "SKILL.md"),
		[]byte("---\nname: video-editing\n---\n"),
		0644,
	))
	writeTagSkillConfig(t, dir, tags)
	return dir
}

func replaceUpstreamTags(t *testing.T, dir string, tags []string) {
	t.Helper()
	writeTagSkillConfig(t, dir, tags)
}

func writeTagSkillConfig(t *testing.T, dir string, tags []string) {
	t.Helper()
	content, err := json.Marshal(skillConfig{
		Schema: skillConfigSchema,
		Slug:   "video-editing",
		Tags:   tags,
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.json"), content, 0644))
}

func importTagSkill(
	t *testing.T,
	svc *Service,
	dir string,
	request *ImportFromGitRequest,
) (*skilldom.Skill, error) {
	t.Helper()
	infos, err := extensionsvc.ScanSkillSource(dir, "")
	require.NoError(t, err)
	require.Len(t, infos, 1)
	return svc.importSkillDir(
		context.Background(),
		request,
		&extensionsvc.ClonedSkillSource{Dir: dir, CommitSha: "abcdef123456"},
		infos[0],
	)
}

func assertSkillConfigTags(t *testing.T, content []byte, expected []string) {
	t.Helper()
	var cfg skillConfig
	require.NoError(t, json.Unmarshal(content, &cfg))
	assert.Equal(t, skillConfigSchema, cfg.Schema)
	assert.Equal(t, expected, cfg.Tags)
}
