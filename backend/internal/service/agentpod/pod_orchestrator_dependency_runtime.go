package agentpod

import (
	"encoding/json"
	"fmt"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	envbundleservice "github.com/l8ai-cn/agentcloud/backend/internal/service/envbundle"
)

func workerDependencyRuntimeInputs(
	document *workerdependency.Document,
) ([]*envbundleservice.EffectiveBundle, map[string]interface{}, error) {
	if document == nil {
		return nil, nil, nil
	}
	bundles := make([]*envbundleservice.EffectiveBundle, 0, len(document.RuntimeBundles))
	configs := map[string]interface{}{}
	for _, bundle := range document.RuntimeBundles {
		if bundle.ConfigDocument != nil {
			config, err := pinnedConfigDocument(bundle)
			if err != nil {
				return nil, nil, err
			}
			configs[bundle.ConfigDocument.ID] = config
			continue
		}
		if bundle.Kind != envbundle.KindRuntime && bundle.Kind != envbundle.KindShared {
			return nil, nil, fmt.Errorf(
				"%w: runtime bundle %d is not runtime-safe",
				ErrWorkerSpecDependencyUnavailable,
				bundle.Pin.DomainID,
			)
		}
		bundles = append(bundles, &envbundleservice.EffectiveBundle{
			ID:         bundle.Pin.DomainID,
			Name:       bundle.Pin.Reference.Name.String(),
			Kind:       bundle.Kind,
			OwnerScope: envbundle.OwnerScopeOrg,
			Data:       runtimeBundleData(bundle.Values),
		})
	}
	if len(configs) == 0 {
		configs = nil
	}
	return bundles, configs, nil
}

func pinnedConfigDocument(bundle workerdependency.RuntimeBundle) (map[string]interface{}, error) {
	raw := ""
	for _, value := range bundle.Values {
		if value.Name == envbundle.ConfigJSONDataKey {
			raw = value.Value
			break
		}
	}
	if raw == "" {
		return nil, fmt.Errorf(
			"%w: config document %q lacks immutable artifact content",
			ErrWorkerSpecDependencyUnavailable,
			bundle.ConfigDocument.ID,
		)
	}
	var document map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &document); err != nil || document == nil {
		return nil, fmt.Errorf(
			"%w: config document %q artifact content is invalid",
			ErrWorkerSpecDependencyUnavailable,
			bundle.ConfigDocument.ID,
		)
	}
	return document, nil
}

func runtimeBundleData(values []workerdependency.RuntimeValue) map[string]string {
	data := make(map[string]string, len(values))
	for _, value := range values {
		data[value.Name] = value.Value
	}
	return data
}
