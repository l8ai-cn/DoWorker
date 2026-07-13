package skill

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

func TestNormalizeTags(t *testing.T) {
	assert.Equal(
		t,
		[]string{"editing", "motion", "video"},
		[]string(skilldom.NormalizeTags([]string{" Video ", "editing", "", "VIDEO", " Editing ", "motion"})),
	)
}

func TestSkillJSONExposesTags(t *testing.T) {
	data, err := json.Marshal(skilldom.Skill{
		Tags: skilldom.NormalizeTags([]string{"Video", "Editing"}),
	})
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(data, &payload))
	assert.Equal(t, []any{"editing", "video"}, payload["tags"])
}

func TestCreate_NormalizesTagsInCatalogAndSkillConfig(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	svc := newTestService(store, fake, &fakePackager{})

	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		UserID:         3,
		Name:           "Video Editing",
		Instructions:   "Edit the video.",
		Tags:           []string{" Video ", "editing", "VIDEO"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"editing", "video"}, []string(row.Tags))

	var cfg skillConfig
	require.NoError(t, json.Unmarshal(fake.Repos["org7-video-editing"].Files["skill.json"], &cfg))
	assert.Equal(t, 2, cfg.Schema)
	assert.Equal(t, []string{"editing", "video"}, cfg.Tags)
}

func TestUpdate_NormalizesAndClearsTags(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	svc := newTestService(store, fake, &fakePackager{})

	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		Name:           "Video Editing",
		Instructions:   "Edit the video.",
		Tags:           []string{"video"},
	})
	require.NoError(t, err)

	tags := []string{" Motion ", "editing", "MOTION"}
	updated, err := svc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7,
		SkillID:        row.ID,
		Tags:           &tags,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"editing", "motion"}, []string(updated.Tags))

	tags = []string{}
	updated, err = svc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7,
		SkillID:        row.ID,
		Tags:           &tags,
	})
	require.NoError(t, err)
	assert.Empty(t, updated.Tags)
}

func TestUpdateTags_PreservesUnknownLargeInteger(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	svc := newTestService(store, fake, &fakePackager{})

	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		Name:           "Video Editing",
		Instructions:   "Edit the video.",
		Tags:           []string{"video"},
	})
	require.NoError(t, err)

	repo := fake.Repos["org7-video-editing"]
	repo.Files["skill.json"] = []byte(
		`{"schema":2,"slug":"video-editing","tags":["video"],"future_number":9007199254740993}`,
	)
	tags := []string{"curated"}
	_, err = svc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7,
		SkillID:        row.ID,
		Tags:           &tags,
	})
	require.NoError(t, err)

	var config map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(repo.Files["skill.json"], &config))
	assert.Equal(t, "9007199254740993", string(config["future_number"]))
}
