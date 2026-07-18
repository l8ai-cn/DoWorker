package workerdependencyartifact

import (
	"fmt"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"
)

func validateToolModelResolutions(
	scope control.Scope,
	resolutions []ToolModelResolution,
	document workerdependency.Document,
) error {
	actual := make(map[string]string, len(resolutions))
	for _, resolution := range resolutions {
		if err := validateToolModelResolution(scope, resolution); err != nil {
			return err
		}
		bindingKey := resolvedReferenceKey(resolution.Binding)
		modelKey := resolvedReferenceKey(resolution.Model.reference)
		if existing, exists := actual[bindingKey]; exists && existing != modelKey {
			return fmt.Errorf("ToolBinding resolves to multiple ModelBindings")
		}
		actual[bindingKey] = modelKey
	}
	required := make(map[string]string, len(document.Models.Tools))
	for _, tool := range document.Models.Tools {
		bindingKey := referenceKey(tool.Binding)
		modelKey := referenceKey(tool.Model.Pin.Reference)
		if existing, exists := required[bindingKey]; exists && existing != modelKey {
			return fmt.Errorf("worker dependency ToolBinding has conflicting models")
		}
		required[bindingKey] = modelKey
	}
	for binding, model := range required {
		if actual[binding] != model {
			return fmt.Errorf("worker dependency ToolBinding resolution is incomplete")
		}
	}
	for binding, model := range actual {
		if required[binding] != model {
			return fmt.Errorf("worker dependency artifact has unused ToolBinding resolution")
		}
	}
	return nil
}

func validateToolModelResolution(
	scope control.Scope,
	resolution ToolModelResolution,
) error {
	if resolution.Binding.Kind != resource.KindToolBinding {
		return fmt.Errorf("tool model resolution parent must be a ToolBinding")
	}
	if resolution.Model.reference.Kind != resource.KindModelBinding {
		return fmt.Errorf("tool model resolution child must be a ModelBinding")
	}
	if err := resolution.Binding.Validate(scope); err != nil {
		return fmt.Errorf("validate ToolBinding resolution: %w", err)
	}
	if err := resolution.Model.reference.Validate(scope); err != nil {
		return fmt.Errorf("validate ToolBinding model resolution: %w", err)
	}
	return nil
}
