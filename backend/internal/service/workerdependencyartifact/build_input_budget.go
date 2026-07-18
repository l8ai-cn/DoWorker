package workerdependencyartifact

import "github.com/anthropics/agentsmesh/backend/internal/domain/workerdependency"

func validateBuildInputBudget(input Input) error {
	budget := workerdependency.MaxDocumentBytes
	consume := func(size int) bool {
		if size < 0 || size > budget {
			return false
		}
		budget -= size
		return true
	}
	text := func(values ...string) bool {
		for _, value := range values {
			if !consume(len(value)) {
				return false
			}
		}
		return true
	}
	count := func(size int) bool {
		const minimumEncodedEntryBytes = 16
		if size > budget/minimumEncodedEntryBytes {
			return false
		}
		return consume(size * minimumEncodedEntryBytes)
	}
	reference := func(resolution ResourceResolution) bool {
		value := resolution.reference
		return text(
			value.APIVersion,
			value.Kind,
			value.Namespace.String(),
			value.Name.String(),
			value.UID,
			value.Digest,
		)
	}
	if !text(
		string(input.Definition.DefinitionSource),
		input.Definition.AgentFile,
		input.AgentfileLayer,
	) {
		return workerdependency.ErrDocumentTooLarge
	}
	if !consumePlanReferencesBudget(input.PlanReferences, text, count) ||
		!consumeWorkerSpecBudget(input.WorkerSpec, consume, text, count) {
		return workerdependency.ErrDocumentTooLarge
	}
	resolved := input.Dependencies
	if !count(len(resolved.ToolModels)) ||
		!count(len(resolved.Skills)) ||
		!count(len(resolved.KnowledgeBases)) ||
		!count(len(resolved.RuntimeBundles)) ||
		!count(len(resolved.SecretReferences)) {
		return workerdependency.ErrDocumentTooLarge
	}
	if resolved.PrimaryModel != nil &&
		!consumeModelBudget(*resolved.PrimaryModel, text, count, reference) {
		return workerdependency.ErrDocumentTooLarge
	}
	for _, tool := range resolved.ToolModels {
		if !text(
			tool.Binding.APIVersion,
			tool.Binding.Kind,
			tool.Binding.Namespace.String(),
			tool.Binding.Name.String(),
			tool.Binding.UID,
			tool.Binding.Digest,
			tool.Role.String(),
			string(tool.Modality),
			string(tool.Capability),
			tool.Environment.APIKeyTarget,
			tool.Environment.BaseURLTarget,
			tool.Environment.ModelIDTarget,
		) || !consumeModelBudget(tool.Model, text, count, reference) {
			return workerdependency.ErrDocumentTooLarge
		}
	}
	if !consumeWorkspaceBudget(resolved, text, count, reference) {
		return workerdependency.ErrDocumentTooLarge
	}
	return nil
}

func consumeModelBudget(
	model ModelResolution,
	text func(...string) bool,
	count func(int) bool,
	reference func(ResourceResolution) bool,
) bool {
	if !reference(model.ResourceResolution) ||
		!text(
			model.ProviderKey.String(),
			model.ProtocolAdapter.String(),
			model.ModelID,
			model.BaseURL,
		) ||
		!count(len(model.Modalities)) ||
		!count(len(model.Capabilities)) {
		return false
	}
	for _, modality := range model.Modalities {
		if !text(string(modality)) {
			return false
		}
	}
	for _, capability := range model.Capabilities {
		if !text(string(capability)) {
			return false
		}
	}
	return true
}
