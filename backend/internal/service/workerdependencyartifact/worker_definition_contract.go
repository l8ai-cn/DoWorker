package workerdependencyartifact

import (
	"fmt"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/workerdefinition"
)

func validateDefinitionDependencies(
	scope control.Scope,
	definition workerdefinition.Definition,
	spec workerspec.Spec,
	document workerdependency.Document,
) error {
	if !supportsInteractionMode(definition.Modes, spec.TypeConfig.InteractionMode) {
		return fmt.Errorf("worker definition does not support the planned interaction mode")
	}
	if err := validateDefinitionModels(definition, spec); err != nil {
		return err
	}
	if err := validateDefinitionConfigDocuments(
		definition.ConfigDocuments,
		document.RuntimeBundles,
	); err != nil {
		return err
	}
	return validateDefinitionSecrets(scope, definition, document.SecretReferences)
}

func supportsInteractionMode(
	modes []string,
	mode workerspec.InteractionMode,
) bool {
	for _, candidate := range modes {
		if candidate == string(mode) {
			return true
		}
	}
	return false
}

func validateDefinitionModels(
	definition workerdefinition.Definition,
	spec workerspec.Spec,
) error {
	primary := spec.Runtime.ModelBinding
	if definition.ModelRequirement.Required != !primary.IsEmpty() {
		return fmt.Errorf("worker definition primary model requirement does not match WorkerSpec")
	}
	if !primary.IsEmpty() &&
		!containsString(
			definition.ModelRequirement.ProtocolAdapters,
			primary.ProtocolAdapter.String(),
		) {
		return fmt.Errorf("worker definition rejects the planned primary model adapter")
	}
	requirements := make(
		map[string]workerdefinition.ToolModelRequirement,
		len(definition.ToolModelRequirements),
	)
	for _, requirement := range definition.ToolModelRequirements {
		requirements[requirement.ID] = requirement
	}
	if len(requirements) != len(spec.Runtime.ToolModelBindings) {
		return fmt.Errorf("worker definition tool model requirements do not match WorkerSpec")
	}
	for _, binding := range spec.Runtime.ToolModelBindings {
		requirement, exists := requirements[binding.Role.String()]
		if !exists || !toolRequirementMatches(requirement, binding) {
			return fmt.Errorf(
				"worker definition rejects tool model role %q",
				binding.Role,
			)
		}
	}
	return nil
}

func toolRequirementMatches(
	requirement workerdefinition.ToolModelRequirement,
	binding workerspec.ToolModelBinding,
) bool {
	return requirement.Modality == string(binding.Modality) &&
		requirement.Capability == string(binding.Capability) &&
		requirement.Environment.APIKey == binding.Environment.APIKey &&
		requirement.Environment.BaseURL == binding.Environment.BaseURL &&
		requirement.Environment.ModelID == binding.Environment.ModelID &&
		containsString(
			requirement.ProviderKeys,
			binding.ModelBinding.ProviderKey.String(),
		) &&
		containsString(
			requirement.ProtocolAdapters,
			binding.ModelBinding.ProtocolAdapter.String(),
		)
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
