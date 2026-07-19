package agent

import (
	"context"
	"fmt"

	envbundledomain "github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
	envbundleservice "github.com/anthropics/agentsmesh/backend/internal/service/envbundle"
)

func (b *ConfigBuilder) buildConfigBundleContext(
	ctx context.Context,
	req *ConfigBuildRequest,
	agentSlug string,
) (map[string]interface{}, error) {
	if req.RequiredConfigDocumentBindings != nil {
		return b.buildExactConfigDocumentContext(ctx, req, agentSlug)
	}
	out := map[string]interface{}{}
	if b.envBundleSvc != nil {
		bundles, err := b.envBundleSvc.GetEffectiveForUser(ctx, req.UserID, req.OrganizationID, agentSlug)
		if err != nil {
			return nil, fmt.Errorf("load config bundles: %w", err)
		}
		documents, err := envbundleservice.ParseConfigDocuments(bundles)
		if err != nil {
			return nil, err
		}
		for name, doc := range documents {
			out[name] = doc
		}
	}
	// Ephemeral session bundles win over persisted ones on name conflict.
	for name, doc := range req.SessionConfigBundles {
		out[name] = doc
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}

func (b *ConfigBuilder) buildExactConfigDocumentContext(
	ctx context.Context,
	req *ConfigBuildRequest,
	agentSlug string,
) (map[string]interface{}, error) {
	if req.PinnedConfigDocuments != nil {
		out := make(map[string]interface{}, len(req.PinnedConfigDocuments)+len(req.SessionConfigBundles))
		for name, document := range req.PinnedConfigDocuments {
			out[name] = document
		}
		for name, document := range req.SessionConfigBundles {
			if _, exists := out[name]; exists {
				return nil, fmt.Errorf("session config document %q conflicts with worker binding", name)
			}
			out[name] = document
		}
		return out, nil
	}
	if len(req.RequiredConfigDocumentBindings) == 0 {
		if len(req.SessionConfigBundles) == 0 {
			return nil, nil
		}
		out := make(map[string]interface{}, len(req.SessionConfigBundles))
		for name, document := range req.SessionConfigBundles {
			out[name] = document
		}
		return out, nil
	}
	loader, ok := b.envBundleSvc.(ExactEnvBundleLoader)
	if !ok {
		return nil, fmt.Errorf("exact config bundle loader is unavailable")
	}
	ids := make([]int64, len(req.RequiredConfigDocumentBindings))
	for index, binding := range req.RequiredConfigDocumentBindings {
		ids[index] = binding.ConfigBundleID
	}
	bundles, err := loader.GetEffectiveByIDs(
		ctx,
		req.UserID,
		req.OrganizationID,
		agentSlug,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("load required config bundles: %w", err)
	}
	byID := make(map[int64]*envbundleservice.EffectiveBundle, len(bundles))
	for _, bundle := range bundles {
		if bundle == nil {
			return nil, fmt.Errorf("required config bundle resolution returned nil")
		}
		if bundle.Kind != envbundledomain.KindConfig {
			return nil, fmt.Errorf(
				"required config bundle %d has kind %q",
				bundle.ID,
				bundle.Kind,
			)
		}
		byID[bundle.ID] = bundle
	}
	out := make(map[string]interface{}, len(req.RequiredConfigDocumentBindings)+len(req.SessionConfigBundles))
	for _, binding := range req.RequiredConfigDocumentBindings {
		bundle, exists := byID[binding.ConfigBundleID]
		if !exists {
			return nil, fmt.Errorf(
				"required config bundle %d was not resolved",
				binding.ConfigBundleID,
			)
		}
		documentID := binding.DocumentID
		if _, exists := out[documentID]; exists {
			return nil, fmt.Errorf("required config document %q is duplicated", documentID)
		}
		document, err := envbundleservice.ParseConfigDocument(bundle)
		if err != nil {
			return nil, fmt.Errorf("config document %q: %w", documentID, err)
		}
		out[documentID] = document
	}
	for name, document := range req.SessionConfigBundles {
		if _, exists := out[name]; exists {
			return nil, fmt.Errorf("session config document %q conflicts with worker binding", name)
		}
		out[name] = document
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}
