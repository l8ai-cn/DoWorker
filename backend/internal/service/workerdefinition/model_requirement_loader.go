package workerdefinition

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type modelRequirementDocument struct {
	Required         *bool    `json:"required"`
	ProtocolAdapters []string `json:"protocol_adapters"`
}

func decodeModelRequirement(raw json.RawMessage) (ModelRequirement, error) {
	var document modelRequirementDocument
	if err := decodeStrict(raw, &document); err != nil {
		return ModelRequirement{}, err
	}
	if document.Required == nil {
		return ModelRequirement{}, fmt.Errorf("required is missing")
	}
	seen := map[string]struct{}{}
	adapters := make([]string, 0, len(document.ProtocolAdapters))
	for _, adapter := range document.ProtocolAdapters {
		if err := slugkit.Validate(adapter); err != nil {
			return ModelRequirement{}, fmt.Errorf("invalid protocol adapter %q: %w", adapter, err)
		}
		if _, exists := seen[adapter]; exists {
			return ModelRequirement{}, fmt.Errorf("duplicate protocol adapter %q", adapter)
		}
		seen[adapter] = struct{}{}
		adapters = append(adapters, adapter)
	}
	if *document.Required && len(adapters) == 0 {
		return ModelRequirement{}, fmt.Errorf("required model resource needs a protocol adapter")
	}
	if !*document.Required && len(adapters) != 0 {
		return ModelRequirement{}, fmt.Errorf("non-model worker cannot declare protocol adapters")
	}
	return ModelRequirement{
		Required:         *document.Required,
		ProtocolAdapters: adapters,
	}, nil
}
