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
	out := map[string]interface{}{}
	if b.envBundleSvc != nil {
		bundles, err := b.envBundleSvc.GetEffectiveForUser(ctx, req.UserID, req.OrganizationID, agentSlug)
		if err == nil {
			for name, doc := range envbundleservice.ParseConfigDocuments(bundles) {
				out[name] = doc
			}
		}
	}
	// Ephemeral session bundles win over persisted ones on name conflict.
	for name, doc := range req.SessionConfigBundles {
		out[name] = doc
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
