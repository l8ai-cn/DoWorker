package operatorcatalog

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/service/workercreation"
)

func (bootstrapper *Bootstrapper) createSnapshotArtifact(
	ctx context.Context,
	request BootstrapRequest,
	snapshotID int64,
	prepared workercreation.Prepared,
) error {
	if prepared.Artifact == nil {
		return ErrCatalogConflict
	}
	return bootstrapper.artifacts.Create(
		ctx,
		request.OrganizationID,
		snapshotID,
		prepared.Artifact.JSON(),
		prepared.Artifact.Digest(),
	)
}
