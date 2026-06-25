package envbundle

import (
	"encoding/json"

	envbundledomain "github.com/anthropics/agentsmesh/backend/internal/domain/envbundle"
)

// ParseConfigDocuments converts config-kind bundles into parsed JSON values
// keyed by bundle name for AgentFile USE_CONFIG_BUNDLE eval.
func ParseConfigDocuments(bundles []*EffectiveBundle) map[string]interface{} {
	out := make(map[string]interface{})
	for _, b := range bundles {
		if b.Kind != envbundledomain.KindConfig {
			continue
		}
		raw, ok := b.Data[envbundledomain.ConfigJSONDataKey]
		if !ok || raw == "" {
			continue
		}
		var doc interface{}
		if err := json.Unmarshal([]byte(raw), &doc); err != nil {
			continue
		}
		out[b.Name] = doc
	}
	return out
}
