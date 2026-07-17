package skill

import (
	"context"
	"testing"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/stretchr/testify/require"
)

func TestEnsurePlatformSkillCreatesPackagedPlatformSkill(t *testing.T) {
	store := newFakeStore()
	service := newTestService(
		store,
		gitops.NewFake("am-skills"),
		&fakePackager{},
	)

	row, created, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)

	require.NoError(t, err)
	require.True(t, created)
	require.Nil(t, row.OrganizationID)
	require.Equal(t, "video-delivery-qa", row.Slug)
	require.Equal(t, []string{"qa", "video"}, []string(row.Tags))
	require.NotEmpty(t, row.ContentSha)
	require.NotEmpty(t, row.StorageKey)
	require.Positive(t, row.PackageSize)
}

func TestEnsurePlatformSkillIsIdempotentForExactManifest(t *testing.T) {
	store := newFakeStore()
	service := newTestService(
		store,
		gitops.NewFake("am-skills"),
		&fakePackager{},
	)
	first, created, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)
	require.NoError(t, err)
	require.True(t, created)

	second, created, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)

	require.NoError(t, err)
	require.False(t, created)
	require.Equal(t, first.ID, second.ID)
	require.Len(t, store.rows, 1)
}

func TestEnsurePlatformSkillRejectsManifestDrift(t *testing.T) {
	service := newTestService(
		newFakeStore(),
		gitops.NewFake("am-skills"),
		&fakePackager{},
	)
	_, _, err := service.EnsurePlatformSkill(
		context.Background(),
		platformSkillRequest(),
	)
	require.NoError(t, err)
	changed := platformSkillRequest()
	changed.Instructions = "Use a different release gate."

	_, _, err = service.EnsurePlatformSkill(context.Background(), changed)

	require.ErrorIs(t, err, ErrPlatformSkillConflict)
}

func TestEnsurePlatformSkillValidatesIdentifierBeforeProvisioning(t *testing.T) {
	service := newTestService(
		newFakeStore(),
		gitops.NewFake("am-skills"),
		&fakePackager{},
	)
	request := platformSkillRequest()
	request.Slug = "Video_QA"

	_, _, err := service.EnsurePlatformSkill(context.Background(), request)

	require.Error(t, err)
}

func platformSkillRequest() *EnsurePlatformSkillRequest {
	return &EnsurePlatformSkillRequest{
		RepositoryOwnerOrganizationID: 7,
		UserID:                        9,
		Slug:                          "video-delivery-qa",
		Name:                          "Video Delivery QA",
		Description:                   "Checks short-form video delivery.",
		License:                       "Apache-2.0",
		Instructions:                  "Validate the rendered video before delivery.",
		Tags:                          []string{"video", "qa"},
	}
}

var _ skilldom.Repository = (*fakeStore)(nil)
