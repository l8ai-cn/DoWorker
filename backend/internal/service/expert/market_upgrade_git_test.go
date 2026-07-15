package expert

import (
	"context"
	"errors"
	"testing"

	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/stretchr/testify/require"
)

func TestMarketUpgradeDatabaseFailureRestoresGitFiles(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	fakeGit := gitops.NewFake("am-experts")
	fixture.service.gitops = fakeGit
	ctx := context.Background()
	v1 := fixture.publishCurrentSource(t)
	installed, _, err := fixture.service.InstallPublishedMarketApplication(
		ctx,
		InstallMarketApplicationRequest{
			OrganizationID:  42,
			UserID:          501,
			ModelResourceID: 301,
			MarketSlug:      string(v1.Application.Slug),
		},
	)
	require.NoError(t, err)
	require.NotNil(t, installed.GitRepoPath)
	repoName := fakeGit.RepoNameFromPath(*installed.GitRepoPath)
	beforeAgent, _, err := fakeGit.ReadFile(ctx, repoName, "main", "agent.md")
	require.NoError(t, err)
	beforeConfig, _, err := fakeGit.ReadFile(ctx, repoName, "main", "expert.json")
	require.NoError(t, err)

	updatedName := "Video Production Expert V2"
	fixture.source, err = fixture.service.Update(ctx, &UpdateExpertRequest{
		OrganizationID: fixture.source.OrganizationID,
		ExpertID:       fixture.source.ID,
		Name:           &updatedName,
		Prompt:         stringPointer("new release prompt"),
	})
	require.NoError(t, err)
	submission, err := fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)
	require.NoError(t, err)
	_, err = fixture.service.ApproveMarketRelease(
		ctx,
		ReviewMarketReleaseRequest{
			ReviewerUserID: 99,
			ReleaseID:      submission.Release.ID,
		},
	)
	require.NoError(t, err)
	fixture.store.updateErr = errors.New("update failed")

	_, _, err = fixture.service.UpgradeMarketApplication(
		ctx,
		UpgradeMarketApplicationRequest{
			OrganizationID: 42,
			UserID:         501,
			ExpertID:       installed.ID,
		},
	)
	require.EqualError(t, err, "update failed")
	afterAgent, _, err := fakeGit.ReadFile(ctx, repoName, "main", "agent.md")
	require.NoError(t, err)
	afterConfig, _, err := fakeGit.ReadFile(ctx, repoName, "main", "expert.json")
	require.NoError(t, err)
	require.Equal(t, beforeAgent, afterAgent)
	require.Equal(t, beforeConfig, afterConfig)
	require.Len(t, fixture.snapshots.created, 2)
}
