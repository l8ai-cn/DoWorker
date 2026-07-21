package orchestrationworker

import (
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

func definitionReferences(
	kind string,
	value any,
) ([]controlservice.DraftReference, error) {
	appendPrompt := func(
		refs []controlservice.DraftReference,
		prompt *resource.Reference,
	) []controlservice.DraftReference {
		if prompt == nil {
			return refs
		}
		return append(refs, controlservice.DraftReference{
			Path: "/spec/promptRef", Reference: *prompt,
		})
	}
	switch spec := value.(type) {
	case *resource.PromptSpec:
		if kind == resource.KindPrompt {
			return []controlservice.DraftReference{}, nil
		}
	case *resource.WorkerInvocationSpec:
		if kind == resource.KindWorker {
			return appendPrompt(workerDefinitionRef(spec.WorkerTemplateRef), spec.PromptRef), nil
		}
	case *resource.ExpertResourceSpec:
		if kind == resource.KindExpert {
			return appendPrompt(workerDefinitionRef(spec.WorkerTemplateRef), spec.PromptRef), nil
		}
	case *resource.WorkflowResourceSpec:
		if kind == resource.KindWorkflow {
			prompt := spec.PromptRef
			return appendPrompt(workerDefinitionRef(spec.WorkerTemplateRef), &prompt), nil
		}
	case *resource.GoalLoopResourceSpec:
		if kind == resource.KindGoalLoop {
			return workerDefinitionRef(spec.WorkerTemplateRef), nil
		}
	}
	return nil, control.ErrCorrupt
}

func workerDefinitionRef(
	reference resource.Reference,
) []controlservice.DraftReference {
	return []controlservice.DraftReference{{
		Path: "/spec/workerTemplateRef", Reference: reference,
	}}
}

func definitionWorkerTemplateReference(value any) resource.Reference {
	switch spec := value.(type) {
	case *resource.WorkerInvocationSpec:
		return spec.WorkerTemplateRef
	case *resource.ExpertResourceSpec:
		return spec.WorkerTemplateRef
	case *resource.WorkflowResourceSpec:
		return spec.WorkerTemplateRef
	case *resource.GoalLoopResourceSpec:
		return spec.WorkerTemplateRef
	default:
		return resource.Reference{}
	}
}

func matchesDefinitionSpec(kind string, value any) bool {
	_, err := definitionReferences(kind, value)
	return err == nil
}
