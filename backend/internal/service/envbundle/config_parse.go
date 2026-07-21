package envbundle

import (
	"fmt"

	envbundledomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/envbundle"
)

// ParseConfigDocuments converts config-kind bundles into parsed JSON values
// keyed by bundle name for AgentFile USE_CONFIG_BUNDLE eval.
func ParseConfigDocuments(
	bundles []*EffectiveBundle,
) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	for _, b := range bundles {
		if b == nil {
			return nil, fmt.Errorf("config bundle resolution returned nil")
		}
		if b.Kind != envbundledomain.KindConfig {
			continue
		}
		doc, err := ParseConfigDocument(b)
		if err != nil {
			return nil, fmt.Errorf("config bundle %q: %w", b.Name, err)
		}
		out[b.Name] = doc
	}
	return out, nil
}

func ParseConfigDocument(bundle *EffectiveBundle) (interface{}, error) {
	if bundle == nil {
		return nil, fmt.Errorf("config bundle resolution returned nil")
	}
	if bundle.Kind != envbundledomain.KindConfig {
		return nil, fmt.Errorf("bundle kind %q is not config", bundle.Kind)
	}
	return decodeConfigDocument(bundle.Data)
}
