package agent

import (
	"context"
	"fmt"
	"log/slog"

	envbundleservice "github.com/l8ai-cn/agentcloud/backend/internal/service/envbundle"
)

func (b *ConfigBuilder) buildEnvBundleContext(
	ctx context.Context,
	req *ConfigBuildRequest,
	agentSlug string,
) (map[string]map[string]string, error) {
	if req.PinnedEnvBundles != nil {
		bundles, err := b.loadRequiredEnvBundles(ctx, req, agentSlug)
		if err != nil {
			return nil, err
		}
		return envbundleservice.AsContextMap(append(req.PinnedEnvBundles, bundles...)), nil
	}
	if len(req.RequiredEnvBundleIDs) > 0 {
		bundles, err := b.loadRequiredEnvBundles(ctx, req, agentSlug)
		if err != nil {
			return nil, err
		}
		return envbundleservice.AsContextMap(bundles), nil
	}
	bundles, err := b.envBundleSvc.GetEffectiveForUser(
		ctx,
		req.UserID,
		req.OrganizationID,
		agentSlug,
	)
	if err != nil {
		slog.WarnContext(
			ctx,
			"Failed to load env bundles for agentfile",
			"agent_slug",
			agentSlug,
			"error",
			err,
		)
		return nil, nil
	}
	return envbundleservice.AsContextMap(bundles), nil
}

func (b *ConfigBuilder) loadRequiredEnvBundles(
	ctx context.Context,
	req *ConfigBuildRequest,
	agentSlug string,
) ([]*envbundleservice.EffectiveBundle, error) {
	if len(req.RequiredEnvBundleIDs) == 0 {
		return nil, nil
	}
	loader, ok := b.envBundleSvc.(ExactEnvBundleLoader)
	if !ok {
		return nil, fmt.Errorf("exact env bundle loader is unavailable")
	}
	bundles, err := loader.GetEffectiveByIDs(
		ctx,
		req.UserID,
		req.OrganizationID,
		agentSlug,
		req.RequiredEnvBundleIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("load required env bundles: %w", err)
	}
	return bundles, validateExactEnvBundles(req.RequiredEnvBundleIDs, bundles)
}

func validateExactEnvBundles(
	requiredIDs []int64,
	bundles []*envbundleservice.EffectiveBundle,
) error {
	expected := make(map[int64]struct{}, len(requiredIDs))
	for _, id := range requiredIDs {
		expected[id] = struct{}{}
	}
	names := make(map[string]int64, len(bundles))
	for _, bundle := range bundles {
		if bundle == nil {
			return fmt.Errorf("required env bundle resolution returned nil")
		}
		if _, exists := expected[bundle.ID]; !exists {
			return fmt.Errorf("required env bundle resolution substituted id %d", bundle.ID)
		}
		delete(expected, bundle.ID)
		if existingID, exists := names[bundle.Name]; exists {
			return fmt.Errorf(
				"required env bundles %d and %d share name %q",
				existingID,
				bundle.ID,
				bundle.Name,
			)
		}
		names[bundle.Name] = bundle.ID
	}
	if len(expected) != 0 {
		return fmt.Errorf("required env bundle resolution is incomplete")
	}
	return nil
}
