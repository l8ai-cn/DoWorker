package infra

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
)

func TestExpertRepositoryMarketUpdatePreservesConsumerFields(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	createRuntimeBindingExpertsTable(t, db)
	repo := NewExpertRepository(db)
	applicationID := int64(90)
	releaseID := int64(91)
	row := &expertdom.Expert{
		OrganizationID:            77,
		Slug:                      "installed-video",
		Name:                      "Version One",
		AgentSlug:                 "video-studio",
		InteractionMode:           expertdom.InteractionModePTY,
		AutomationLevel:           expertdom.AutomationLevelAutonomous,
		SkillSlugs:                pq.StringArray{"old-skill"},
		KnowledgeMounts:           json.RawMessage("[]"),
		ConfigOverrides:           json.RawMessage("{}"),
		Metadata:                  json.RawMessage("{}"),
		SourceMarketApplicationID: &applicationID,
		SourceMarketReleaseID:     &releaseID,
		CreatedByID:               12,
		RunCount:                  4,
	}
	require.NoError(t, repo.Create(ctx, row))
	nextReleaseID := int64(92)
	require.NoError(t, repo.UpdateMarketRelease(
		ctx,
		77,
		row.ID,
		applicationID,
		expertdom.MarketReleaseUpdate{
			Name:                  "Version Two",
			AgentSlug:             "video-studio",
			InteractionMode:       expertdom.InteractionModePTY,
			AutomationLevel:       expertdom.AutomationLevelAutonomous,
			UsedEnvBundles:        pq.StringArray{},
			SkillSlugs:            pq.StringArray{"new-skill"},
			KnowledgeMounts:       json.RawMessage("[]"),
			ConfigOverrides:       json.RawMessage("{}"),
			Metadata:              json.RawMessage(`{"version":2}`),
			WorkerSpecSnapshotID:  200,
			SourceMarketReleaseID: nextReleaseID,
			ExpectedRevision:      row.Revision,
		},
	))

	stored, err := repo.GetByMarketApplication(ctx, 77, applicationID)
	require.NoError(t, err)
	require.Equal(t, "Version Two", stored.Name)
	require.Equal(t, row.Slug, stored.Slug)
	require.Equal(t, int64(12), stored.CreatedByID)
	require.Equal(t, 4, stored.RunCount)
	require.Equal(t, nextReleaseID, *stored.SourceMarketReleaseID)
	require.Equal(t, int64(200), *stored.WorkerSpecSnapshotID)
	stale := expertdom.MarketReleaseUpdate{
		ExpectedRevision: row.Revision,
	}
	require.ErrorIs(t, repo.UpdateMarketRelease(
		ctx,
		77,
		row.ID,
		applicationID,
		stale,
	), expertdom.ErrConflict)
	require.ErrorIs(t, repo.UpdateMarketRelease(
		ctx,
		77,
		row.ID,
		999,
		expertdom.MarketReleaseUpdate{},
	), expertdom.ErrNotFound)
}
