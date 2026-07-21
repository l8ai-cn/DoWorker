package operatorcatalog

import (
	"context"
	"testing"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	"github.com/stretchr/testify/require"
)

func TestBootstrapVideoExpertsRejectsLegacyProtocolAdapterSnapshot(t *testing.T) {
	snapshots := newBootstrapSnapshotStore()
	bootstrapper := NewBootstrapper(
		&bootstrapSkillStore{},
		newBootstrapExpertStore(),
		&bootstrapWorkerPreparer{},
		snapshots,
		&bootstrapDependencyArtifactStore{},
	)
	request := BootstrapRequest{
		OrganizationID: 7, OrganizationSlug: slugkit.MustNewForTest("dev-org"),
		PublisherUserID: 11, ReviewerUserID: 13,
		ModelResourceID: 17, RuntimeImageID: 19,
	}
	_, err := bootstrapper.Run(context.Background(), request)
	require.NoError(t, err)
	for id, snapshot := range snapshots.rows {
		snapshot.Spec.Runtime.ModelBinding.ProtocolAdapter = ""
		snapshots.rows[id] = snapshot
	}

	_, err = bootstrapper.Run(context.Background(), request)

	require.ErrorIs(t, err, ErrCatalogConflict)
}
