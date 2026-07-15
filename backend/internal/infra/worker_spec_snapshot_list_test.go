package infra

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkerSpecSnapshotRepositoryListsOrganizationNewestFirst(t *testing.T) {
	ctx := context.Background()
	repo := NewWorkerSpecSnapshotRepository(workerSpecSnapshotDBForContract(t))

	first, err := repo.Create(ctx, workerSpecSnapshotForContract(t, 77))
	require.NoError(t, err)
	_, err = repo.Create(ctx, workerSpecSnapshotForContract(t, 78))
	require.NoError(t, err)
	latest, err := repo.Create(ctx, workerSpecSnapshotForContract(t, 77))
	require.NoError(t, err)

	snapshots, err := repo.ListByOrganization(ctx, 77)

	require.NoError(t, err)
	require.Len(t, snapshots, 2)
	require.Equal(t, []int64{latest.ID, first.ID}, []int64{
		snapshots[0].ID,
		snapshots[1].ID,
	})
}
