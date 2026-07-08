package sessionapi

import (
	"context"

	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
)

func (d *Deps) resolvePrimaryEnvBundle(ctx context.Context, userID, orgID int64, agentSlug string, layer **string) {
	agentpod.AppendPrimaryCredentialBundle(ctx, d.EnvBundles, userID, orgID, agentSlug, layer)
}
