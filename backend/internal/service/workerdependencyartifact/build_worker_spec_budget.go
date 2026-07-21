package workerdependencyartifact

import (
	"reflect"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/workerspec"
)

func consumePlanReferencesBudget(
	references []control.ResolvedReference,
	text func(...string) bool,
	count func(int) bool,
) bool {
	if !count(len(references)) {
		return false
	}
	for _, reference := range references {
		if !text(
			reference.APIVersion,
			reference.Kind,
			reference.Namespace.String(),
			reference.Name.String(),
			reference.UID,
			reference.Digest,
		) {
			return false
		}
	}
	return true
}

func consumeWorkerSpecBudget(
	spec workerspec.Spec,
	consume func(int) bool,
	text func(...string) bool,
	count func(int) bool,
) bool {
	runtime := spec.Runtime
	if !consumeModelBindingBudget(runtime.ModelBinding, text) ||
		!text(
			runtime.WorkerType.Slug.String(),
			runtime.WorkerType.DefinitionHash,
			runtime.Image.Digest,
		) ||
		!count(len(runtime.ToolModelBindings)) {
		return false
	}
	for _, binding := range runtime.ToolModelBindings {
		if !consumeModelBindingBudget(binding.ModelBinding, text) || !text(
			binding.Role.String(),
			string(binding.Modality),
			string(binding.Capability),
			binding.Environment.APIKey,
			binding.Environment.BaseURL,
			binding.Environment.ModelID,
		) {
			return false
		}
	}
	config := spec.TypeConfig
	if !text(
		string(config.InteractionMode),
		string(config.AutomationLevel),
	) ||
		!consumeJSONBudget(reflect.ValueOf(config.Values), consume, text, count, 0) ||
		!count(len(config.SecretRefs)) {
		return false
	}
	for field, reference := range config.SecretRefs {
		if !text(field, reference.Kind.String()) {
			return false
		}
	}
	workspace := spec.Workspace
	if !text(
		workspace.Branch,
		workspace.Instructions,
		workspace.InitialTask,
	) || !count(len(workspace.SkillIDs)) ||
		!count(len(workspace.KnowledgeMounts)) ||
		!count(len(workspace.EnvBundleIDs)) ||
		!count(len(workspace.ConfigDocumentBindings)) {
		return false
	}
	for _, mount := range workspace.KnowledgeMounts {
		if !text(string(mount.Mode)) {
			return false
		}
	}
	for _, binding := range workspace.ConfigDocumentBindings {
		if !text(binding.DocumentID) {
			return false
		}
	}
	return text(
		string(spec.Placement.Policy),
		string(spec.Placement.ComputeTarget.Kind),
		string(spec.Placement.DeploymentMode),
		string(spec.Lifecycle.TerminationPolicy),
		spec.Metadata.Alias,
	)
}

func consumeModelBindingBudget(
	binding workerspec.ModelBinding,
	text func(...string) bool,
) bool {
	return text(
		binding.ProviderKey.String(),
		binding.ProtocolAdapter.String(),
		binding.ModelID,
	)
}
