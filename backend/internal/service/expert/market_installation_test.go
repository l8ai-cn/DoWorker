package expert

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	"github.com/stretchr/testify/require"
)

func TestMarketInstallIsIdempotentAndCreatesConsumerSnapshot(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	published := fixture.publishCurrentSource(t)

	first, existing, err := fixture.service.InstallPublishedMarketApplication(
		ctx,
		InstallMarketApplicationRequest{
			OrganizationID:  42,
			UserID:          501,
			ModelResourceID: 301,
			MarketSlug:      string(published.Application.Slug),
		},
	)
	require.NoError(t, err)
	require.False(t, existing)
	require.NotNil(t, first.WorkerSpecSnapshotID)
	require.NotEqual(t, *fixture.source.WorkerSpecSnapshotID, *first.WorkerSpecSnapshotID)
	require.Equal(t, int64(42), fixture.snapshots.created[0].OrganizationID)
	require.Equal(
		t,
		int64(301),
		fixture.snapshots.created[0].Spec.Runtime.ModelBinding.ResourceID,
	)
	require.Equal(t, published.Application.ID, *first.SourceMarketApplicationID)
	require.Equal(t, published.Release.ID, *first.SourceMarketReleaseID)

	second, existing, err := fixture.service.InstallPublishedMarketApplication(
		ctx,
		InstallMarketApplicationRequest{
			OrganizationID:  42,
			UserID:          501,
			ModelResourceID: 301,
			MarketSlug:      string(published.Application.Slug),
		},
	)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, first.ID, second.ID)
	require.Len(t, fixture.snapshots.created, 1)
}

func TestMarketInstallRaceRemovesUnusedConsumerSnapshot(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	published := fixture.publishCurrentSource(t)

	first, _, err := fixture.service.InstallPublishedMarketApplication(
		ctx,
		InstallMarketApplicationRequest{
			OrganizationID:  42,
			UserID:          501,
			ModelResourceID: 301,
			MarketSlug:      string(published.Application.Slug),
		},
	)
	require.NoError(t, err)
	fixture.store.marketLookupMisses = 1

	second, existing, err := fixture.service.InstallPublishedMarketApplication(
		ctx,
		InstallMarketApplicationRequest{
			OrganizationID:  42,
			UserID:          501,
			ModelResourceID: 301,
			MarketSlug:      string(published.Application.Slug),
		},
	)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, first.ID, second.ID)
	require.Len(t, fixture.snapshots.created, 1)
}

func TestMarketInstallSerializesGitBackedConcurrentRequests(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	fixture.service.gitops = gitops.NewFake("am-experts")
	published := fixture.publishCurrentSource(t)
	ctx := context.Background()
	start := make(chan struct{})
	type result struct {
		expertID int64
		existing bool
		err      error
	}
	results := make(chan result, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			row, existing, err := fixture.service.InstallPublishedMarketApplication(
				ctx,
				InstallMarketApplicationRequest{
					OrganizationID:  42,
					UserID:          501,
					ModelResourceID: 301,
					MarketSlug:      string(published.Application.Slug),
				},
			)
			var expertID int64
			if row != nil {
				expertID = row.ID
			}
			results <- result{expertID: expertID, existing: existing, err: err}
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	var rows []result
	for row := range results {
		require.NoError(t, row.err)
		rows = append(rows, row)
	}
	require.Len(t, rows, 2)
	require.Equal(t, rows[0].expertID, rows[1].expertID)
	require.NotEqual(t, rows[0].existing, rows[1].existing)
	require.Len(t, fixture.snapshots.created, 1)
	require.Equal(t, 2, fixture.locker.calls)
}

func TestMarketInstallRejectsInvalidPublishedExpertSnapshot(t *testing.T) {
	tests := map[string]func(map[string]any){
		"name is required": func(snapshot map[string]any) {
			snapshot["name"] = ""
		},
		"slug is an identifier": func(snapshot map[string]any) {
			snapshot["slug"] = "Invalid_Slug"
		},
		"interaction mode is exact": func(snapshot map[string]any) {
			snapshot["interaction_mode"] = "ssh"
		},
		"metadata is an object": func(snapshot map[string]any) {
			snapshot["metadata"] = []any{}
		},
		"agentfile layer is not portable": func(snapshot map[string]any) {
			snapshot["agentfile_layer"] = `REPO "private-source"`
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newMarketServiceFixture(t)
			published := fixture.publishCurrentSource(t)
			var snapshot map[string]any
			require.NoError(t, json.Unmarshal(
				published.Release.ExpertSnapshot,
				&snapshot,
			))
			mutate(snapshot)
			raw, err := json.Marshal(snapshot)
			require.NoError(t, err)
			release := fixture.market.releases[published.Release.ID]
			release.ExpertSnapshot = raw
			fixture.market.releases[release.ID] = release

			_, _, err = fixture.service.InstallPublishedMarketApplication(
				context.Background(),
				InstallMarketApplicationRequest{
					OrganizationID:  42,
					UserID:          501,
					ModelResourceID: 301,
					MarketSlug:      string(published.Application.Slug),
				},
			)
			require.ErrorIs(t, err, ErrMarketSnapshotInvalid)
			require.Empty(t, fixture.snapshots.created)
		})
	}
}

func TestMarketInstallUsesApprovedSkillPackagesAfterCatalogDrift(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	published := fixture.publishCurrentSource(t)
	fixture.skills.rows[0].Version++
	fixture.skills.rows[0].ContentSha = "changed-sha"
	fixture.skills.rows[0].StorageKey = "skills/changed-package"

	installed, _, err := fixture.service.InstallPublishedMarketApplication(
		context.Background(),
		InstallMarketApplicationRequest{
			OrganizationID:  42,
			UserID:          501,
			ModelResourceID: 301,
			MarketSlug:      string(published.Application.Slug),
		},
	)
	require.NoError(t, err)
	require.NotNil(t, installed.WorkerSpecSnapshotID)
	require.Len(t, fixture.snapshots.created, 1)
	packages := fixture.snapshots.created[0].Spec.Workspace.SkillPackages
	require.Len(t, packages, 2)
	require.Equal(t, "sha-remotion-best-practices", packages[0].ContentSHA)
	require.Equal(t, "skills/remotion-best-practices", packages[0].StorageKey)
}

func TestMarketUpgradeIsExplicitAndPreservesConsumerFields(t *testing.T) {
	fixture := newMarketServiceFixture(t)
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
	require.NoError(t, fixture.store.RecordRun(
		ctx,
		installed.OrganizationID,
		installed.ID,
		time.Now().UTC(),
	))

	updatedName := "Video Production Expert V2"
	fixture.source, err = fixture.service.Update(ctx, &UpdateExpertRequest{
		OrganizationID: fixture.source.OrganizationID,
		ExpertID:       fixture.source.ID,
		Name:           &updatedName,
		Prompt:         stringPointer("render a stronger hook"),
	})
	require.NoError(t, err)
	v2Submission, err := fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)
	require.NoError(t, err)
	_, err = fixture.service.ApproveMarketRelease(
		ctx,
		ReviewMarketReleaseRequest{
			ReviewerUserID: 99,
			ReleaseID:      v2Submission.Release.ID,
		},
	)
	require.NoError(t, err)

	beforeUpgrade, err := fixture.store.GetByID(ctx, 42, installed.ID)
	require.NoError(t, err)
	require.Equal(t, v1.Release.ID, *beforeUpgrade.SourceMarketReleaseID)
	require.Equal(t, "Video Production Expert", beforeUpgrade.Name)
	available, err := fixture.service.MarketUpgradeAvailable(
		ctx,
		42,
		v1.Application.ID,
	)
	require.NoError(t, err)
	require.True(t, available)

	upgraded, changed, err := fixture.service.UpgradeMarketApplication(
		ctx,
		UpgradeMarketApplicationRequest{
			OrganizationID: 42,
			UserID:         501,
			ExpertID:       installed.ID,
		},
	)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, v2Submission.Release.ID, *upgraded.SourceMarketReleaseID)
	require.Equal(t, "Video Production Expert V2", upgraded.Name)
	require.Equal(t, "render a stronger hook", *upgraded.Prompt)
	require.Equal(t, installed.Slug, upgraded.Slug)
	require.Equal(t, int64(501), upgraded.CreatedByID)
	require.Equal(t, 1, upgraded.RunCount)
	require.Len(t, fixture.snapshots.created, 3)

	unchanged, changed, err := fixture.service.UpgradeMarketApplication(
		ctx,
		UpgradeMarketApplicationRequest{
			OrganizationID: 42,
			UserID:         501,
			ExpertID:       installed.ID,
		},
	)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, upgraded.ID, unchanged.ID)
	require.Len(t, fixture.snapshots.created, 3)
}

func TestMarketUpgradeFailureRemovesUnusedConsumerSnapshot(t *testing.T) {
	fixture := newMarketServiceFixture(t)
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

	fixture.source.Name = "Video Production Expert V2"
	require.NoError(t, fixture.store.Update(ctx, fixture.source))
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
	require.Len(t, fixture.snapshots.created, 1)
}

func (fixture *marketServiceFixture) publishCurrentSource(
	t *testing.T,
) *MarketSubmission {
	t.Helper()
	submission, err := fixture.service.SubmitMarketApplication(
		context.Background(),
		fixture.submissionRequest(),
	)
	require.NoError(t, err)
	_, err = fixture.service.ApproveMarketRelease(
		context.Background(),
		ReviewMarketReleaseRequest{
			ReviewerUserID: 99,
			ReleaseID:      submission.Release.ID,
		},
	)
	require.NoError(t, err)
	application, err := fixture.market.GetApplicationByID(
		context.Background(),
		submission.Application.ID,
	)
	require.NoError(t, err)
	release, err := fixture.market.GetReleaseByID(
		context.Background(),
		submission.Release.ID,
	)
	require.NoError(t, err)
	return &MarketSubmission{
		Application: *application,
		Release:     *release,
	}
}
