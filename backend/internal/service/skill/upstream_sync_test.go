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
