package mcp

import (
	"encoding/json"
	"fmt"
)

func resourceManifestArgument(
	args map[string]interface{},
) (json.RawMessage, error) {
	value, ok := args["resource"]
	if !ok {
		return nil, fmt.Errorf("resource is required")
	}
	if _, ok := value.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("resource must be a JSON object")
	}
	content, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode resource: %w", err)
	}
	return json.RawMessage(content), nil
}
