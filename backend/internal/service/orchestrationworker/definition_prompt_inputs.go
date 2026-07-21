package orchestrationworker

import (
	"context"
	"sort"

	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	resource "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationresource"
	controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
)

func (planner *DefinitionPlanner) promptInputIssues(
	ctx context.Context,
	input controlservice.TargetPlanInput,
	pins pinnedReferenceIndex,
) ([]control.PlanIssue, error) {
	promptRef, inputs, exists := promptInputs(input.TypedSpec)
	if !exists {
		return []control.PlanIssue{}, nil
	}
	resolved, err := pins.resolve(promptRef)
	if err != nil {
		return nil, control.ErrCorrupt
	}
	prompt, err := planner.resolver.ResolvePromptSpec(ctx, input.Scope, resolved)
	if err != nil {
		return nil, err
	}
	return validatePromptInputs(prompt, inputs), nil
}

func promptInputs(value any) (
	resource.Reference,
	map[string]string,
	bool,
) {
	switch spec := value.(type) {
	case *resource.WorkerInvocationSpec:
		if spec.PromptRef == nil {
			return resource.Reference{}, nil, false
		}
		return *spec.PromptRef, spec.Inputs, true
	case *resource.ExpertResourceSpec:
		if spec.PromptRef == nil {
			return resource.Reference{}, nil, false
		}
		return *spec.PromptRef, map[string]string{}, true
	case *resource.WorkflowResourceSpec:
		return spec.PromptRef, spec.Inputs, true
	default:
		return resource.Reference{}, nil, false
	}
}

func validatePromptInputs(
	prompt resource.PromptSpec,
	inputs map[string]string,
) []control.PlanIssue {
	issues := make([]control.PlanIssue, 0)
	keys := make([]string, 0, len(prompt.Variables))
	for key := range prompt.Variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		variable := prompt.Variables[key]
		if _, exists := inputs[key]; variable.Required &&
			variable.Default == nil && !exists {
			issues = append(issues, control.PlanIssue{
				Severity: control.PlanIssueBlocking,
				Path:     "/spec/inputs/" + key,
				Code:     "missing-prompt-input",
				Message:  "A required prompt input is missing.",
			})
		}
	}
	inputKeys := make([]string, 0, len(inputs))
	for key := range inputs {
		inputKeys = append(inputKeys, key)
	}
	sort.Strings(inputKeys)
	for _, key := range inputKeys {
		if _, exists := prompt.Variables[key]; !exists {
			issues = append(issues, control.PlanIssue{
				Severity: control.PlanIssueBlocking,
				Path:     "/spec/inputs/" + key,
				Code:     "unknown-prompt-input",
				Message:  "The prompt does not declare this input.",
			})
		}
	}
	return issues
}
