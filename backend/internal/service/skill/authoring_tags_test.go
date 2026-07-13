package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

func TestValidateTagsEnforcesNormalizedBoundaries(t *testing.T) {
	valid := make([]string, 20)
	for i := range valid {
		valid[i] = fmt.Sprintf("%02d-%s", i, strings.Repeat("界", 17))
	}
	normalized, err := ValidateTags(valid)
	require.NoError(t, err)
	assert.Len(t, normalized, 20)
	_, err = ValidateTags([]string{strings.Repeat("界", 40)})
	require.NoError(t, err)

	tests := map[string][]string{
		"more than twenty": append(valid, "extra"),
		"tag over forty code points": {
			strings.Repeat("界", 41),
		},
		"total over four hundred code points": func() []string {
			tags := make([]string, 20)
			for i := range tags {
				tags[i] = fmt.Sprintf("%02d%s", i, strings.Repeat("a", 19))
			}
			return tags
		}(),
	}
	for name, tags := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateTags(tags)
			assert.ErrorIs(t, err, ErrInvalidTags)
		})
	}
}

func TestCreateAndUpdateRejectInvalidTags(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	svc := newTestService(store, fake, &fakePackager{})
	tooMany := make([]string, 21)
	for i := range tooMany {
		tooMany[i] = fmt.Sprintf("tag-%02d", i)
	}

	_, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		Name:           "Video Editing",
		Instructions:   "Edit the video.",
		Tags:           tooMany,
	})
	assert.ErrorIs(t, err, ErrInvalidTags)
	assert.Empty(t, fake.Repos)

	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7,
		Name:           "Audio Editing",
		Instructions:   "Edit the audio.",
	})
	require.NoError(t, err)
	_, err = svc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7,
		SkillID:        row.ID,
		Tags:           &tooMany,
	})
	assert.ErrorIs(t, err, ErrInvalidTags)
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
