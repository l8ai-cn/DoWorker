package agent

import (
	"context"

	envbundleservice "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
)

func (b *ConfigBuilder) buildConfigBundleContext(
	ctx context.Context,
	req *ConfigBuildRequest,
	agentSlug string,
) map[string]interface{} {
	if b.envBundleSvc == nil {
		return nil
	}
	bundles, err := b.envBundleSvc.GetEffectiveForUser(ctx, req.UserID, req.OrganizationID, agentSlug)
	if err != nil {
		return nil
	}
	return envbundleservice.ParseConfigDocuments(bundles)
}
