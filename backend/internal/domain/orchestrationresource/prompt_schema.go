package orchestrationresource

import (
	"fmt"
	"sort"

	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type PromptVariableSpec struct {
	Required bool    `json:"required" yaml:"required"`
	Default  *string `json:"default,omitempty" yaml:"default,omitempty"`
}

type PromptSpec struct {
	Content   string                        `json:"content" yaml:"content"`
	Variables map[string]PromptVariableSpec `json:"variables" yaml:"variables"`
}

func promptSchema() Schema {
	return Schema{
		NewSpec: func() any { return &PromptSpec{} },
		Validate: func(_ Metadata, value any) error {
			return validatePromptSpec(value.(*PromptSpec))
		},
	}
}

func validatePromptSpec(spec *PromptSpec) error {
	if err := validateDefinitionText(
		"content",
		spec.Content,
		65_536,
		true,
	); err != nil {
		return err
	}
	if spec.Variables == nil {
		return fmt.Errorf("variables must be an object")
	}
	if len(spec.Variables) > 128 {
		return fmt.Errorf("variables exceeds 128 entries")
	}
	keys := make([]string, 0, len(spec.Variables))
	for key := range spec.Variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := slugkit.Validate(key); err != nil {
			return fmt.Errorf("prompt variable %q: %w", summarizeValue(key), err)
		}
		variable := spec.Variables[key]
		if variable.Default != nil {
			if err := validateDefinitionText(
				"prompt variable default",
				*variable.Default,
				8_192,
				false,
			); err != nil {
				return err
			}
		}
	}
	return nil
}
