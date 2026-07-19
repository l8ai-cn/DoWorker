package operatorcatalog

import (
	"context"
	"errors"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
)

func (bootstrapper *Bootstrapper) rebuildExpertSnapshot(
	ctx context.Context,
	request BootstrapRequest,
	expert *expertdom.Expert,
	prepared workercreation.Prepared,
) error {
	snapshot, err := bootstrapper.snapshots.Create(ctx, prepared.Snapshot)
	if err != nil {
		return err
	}
	if err := bootstrapper.createSnapshotArtifact(ctx, request, snapshot.ID, prepared); err != nil {
		cleanupErr := bootstrapper.snapshots.Delete(
			context.WithoutCancel(ctx),
			request.OrganizationID,
			snapshot.ID,
		)
		return errors.Join(err, cleanupErr)
	}
	updated, err := bootstrapper.experts.RebindWorkerSpecSnapshot(
		ctx,
		request.OrganizationID,
		expert.ID,
		snapshot.ID,
	)
	if err != nil {
		artifactErr := bootstrapper.artifacts.Delete(
			context.WithoutCancel(ctx),
			request.OrganizationID,
			snapshot.ID,
		)
		snapshotErr := bootstrapper.snapshots.Delete(
			context.WithoutCancel(ctx),
			request.OrganizationID,
			snapshot.ID,
		)
		return errors.Join(err, artifactErr, snapshotErr)
	}
	expert.WorkerSpecSnapshotID = updated.WorkerSpecSnapshotID
	return nil
}
