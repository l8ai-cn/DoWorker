package expert

import (
	"context"
	"testing"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/stretchr/testify/require"
)

func TestUpdateRefreshesWorkerSpecSnapshotForEditableRuntimeFields(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	originalSnapshotID := *fixture.source.WorkerSpecSnapshotID
	prompt := "cut a stronger opening"
	mode := expertdom.InteractionModePTY
	automation := expertdom.AutomationLevelAutoEdit

	updated, err := fixture.service.Update(
		context.Background(),
		&UpdateExpertRequest{
			OrganizationID:  fixture.source.OrganizationID,
			ExpertID:        fixture.source.ID,
			Prompt:          &prompt,
			InteractionMode: &mode,
			AutomationLevel: &automation,
			ConfigOverrides: map[string]interface{}{"approval_mode": "never"},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, updated.WorkerSpecSnapshotID)
	require.NotEqual(
		t,
		originalSnapshotID,
		*updated.WorkerSpecSnapshotID,
	)
	require.Len(t, fixture.snapshots.created, 1)
	spec := fixture.snapshots.created[0].Spec
	require.Equal(t, prompt, spec.Workspace.Instructions)
	require.Equal(t, mode, string(spec.TypeConfig.InteractionMode))
	require.Equal(t, automation, string(spec.TypeConfig.AutomationLevel))
	require.Equal(t, "never", spec.TypeConfig.Values["approval_mode"])
}

func TestUpdateRejectsUnresolvableSnapshotRuntimeChanges(t *testing.T) {
	fixture := newMarketServiceFixture(t)

	_, err := fixture.service.Update(
		context.Background(),
		&UpdateExpertRequest{
			OrganizationID: fixture.source.OrganizationID,
			ExpertID:       fixture.source.ID,
			SkillSlugs:     []string{"different-skill"},
		},
	)

	require.ErrorIs(t, err, ErrExpertSnapshotUpdateUnsupported)
	require.Empty(t, fixture.snapshots.created)
}

func TestUpdateDeletesNewSnapshotWhenExpertPersistenceFails(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	fixture.store.updateErr = context.Canceled
	prompt := "new prompt"

	_, err := fixture.service.Update(
		context.Background(),
		&UpdateExpertRequest{
			OrganizationID: fixture.source.OrganizationID,
			ExpertID:       fixture.source.ID,
			Prompt:         &prompt,
		},
	)

	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, fixture.snapshots.created)
	require.Len(t, fixture.snapshots.deleteContexts, 1)
	require.NoError(t, fixture.snapshots.deleteErrors[0])
}
