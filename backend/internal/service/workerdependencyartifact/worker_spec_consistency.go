package workerdependencyartifact

import (
	"fmt"
	"reflect"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerdependency"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func ValidateWorkerSpecConsistency(
	spec workerspec.Spec,
	document workerdependency.Document,
) error {
	if document.Worker.SpecVersion != spec.Version ||
		document.Worker.WorkerType != spec.Runtime.WorkerType.Slug ||
		document.Worker.DefinitionHash != spec.Runtime.WorkerType.DefinitionHash {
		return fmt.Errorf("worker dependency worker identity does not match WorkerSpec")
	}
	if document.Placement.RuntimeImage.ID != spec.Runtime.Image.ID ||
		document.Placement.RuntimeImage.Digest != spec.Runtime.Image.Digest ||
		!reflect.DeepEqual(document.Placement.Spec, spec.Placement) {
		return fmt.Errorf("worker dependency placement does not match WorkerSpec")
	}
	if err := validatePrimaryModel(spec.Runtime.ModelBinding, document.Models.Primary); err != nil {
		return err
	}
	if err := validateToolModels(spec.Runtime.ToolModelBindings, document.Models.Tools); err != nil {
		return err
	}
	return validateWorkspace(spec, document)
}

func validatePrimaryModel(
	binding workerspec.ModelBinding,
	model *workerdependency.Model,
) error {
	if binding.IsEmpty() {
		if model != nil {
			return fmt.Errorf("worker dependency primary model is absent from WorkerSpec")
		}
		return nil
	}
	if model == nil || !modelBindingMatches(binding, *model) {
		return fmt.Errorf("worker dependency primary model does not match WorkerSpec")
	}
	return nil
}

func validateToolModels(
	bindings []workerspec.ToolModelBinding,
	models []workerdependency.ToolModel,
) error {
	if len(bindings) != len(models) {
		return fmt.Errorf("worker dependency tool models do not match WorkerSpec")
	}
	byRole := make(map[string]workerdependency.ToolModel, len(models))
	for _, model := range models {
		byRole[model.Role.String()] = model
	}
	for _, binding := range bindings {
		model, exists := byRole[binding.Role.String()]
		if !exists ||
			!modelBindingMatches(binding.ModelBinding, model.Model) ||
			binding.Modality != model.Modality ||
			binding.Capability != model.Capability ||
			!environmentMatches(binding.Environment, model.Environment) {
			return fmt.Errorf(
				"worker dependency tool model %q does not match WorkerSpec",
				binding.Role,
			)
		}
	}
	return nil
}

func environmentMatches(
	binding workerspec.ToolModelEnvironment,
	environment workerdependency.ToolModelEnvironment,
) bool {
	return binding.APIKey == environment.APIKeyTarget &&
		binding.BaseURL == environment.BaseURLTarget &&
		binding.ModelID == environment.ModelIDTarget
}

func modelBindingMatches(
	binding workerspec.ModelBinding,
	model workerdependency.Model,
) bool {
	return binding.ResourceID == model.Pin.DomainID &&
		binding.ResourceRevision == model.ResourceRevision &&
		binding.ConnectionID == model.ConnectionID &&
		binding.ConnectionRevision == model.ConnectionRevision &&
		binding.ProviderKey == model.ProviderKey &&
		binding.ProtocolAdapter == model.ProtocolAdapter &&
		binding.ModelID == model.ModelID
}
