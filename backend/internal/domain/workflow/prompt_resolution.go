package workflow

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ResolvePrompt(template string, defaults, overrides json.RawMessage) string {
	variables := make(map[string]any)
	if len(defaults) > 0 {
		_ = json.Unmarshal(defaults, &variables)
	}
	if len(overrides) > 0 {
		var overrideValues map[string]any
		if err := json.Unmarshal(overrides, &overrideValues); err == nil {
			for key, value := range overrideValues {
				variables[key] = value
			}
		}
	}

	resolved := template
	for key, value := range variables {
		resolved = strings.ReplaceAll(
			resolved,
			"{{"+key+"}}",
			fmt.Sprintf("%v", value),
		)
	}
	return resolved
}
