package agentpod

import (
	"encoding/json"
)

type ConfigReference struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Available bool   `json:"available"`
}

type configSummary struct {
	References map[string]ConfigReference `json:"references,omitempty"`
}

func NewSafeConfigSummary(references map[string]ConfigReference, _ map[string]string) (json.RawMessage, error) {
	allowed := map[string]ConfigReference{}
	for _, key := range []string{"model_resource", "repository", "env_bundle", "config_bundle"} {
		if reference, ok := references[key]; ok {
			allowed[key] = reference
		}
	}
	return json.Marshal(configSummary{References: allowed})
}
