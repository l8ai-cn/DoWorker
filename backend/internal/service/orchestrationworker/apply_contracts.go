package orchestrationworker

import (
	"context"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
)

type BindingApplyBuilder = controlservice.BindingApplyBuilder

type BindingApplyRepository interface {
	RunBindingApplyTransaction(
		context.Context,
		control.Scope,
		string,
		BindingApplyBuilder,
	) (control.ResourceHead, error)
}

type WorkerTemplateApplyBuilder = controlservice.WorkerTemplateApplyBuilder
type AppliedWorkerTemplate = controlservice.AppliedWorkerTemplate

type WorkerTemplateApplyRepository interface {
	RunWorkerTemplateApplyTransaction(
		context.Context,
		control.Scope,
		string,
		WorkerTemplateApplyBuilder,
	) (AppliedWorkerTemplate, error)
}

type PromptApplyRepository interface {
	RunPromptApplyTransaction(
		context.Context,
		control.Scope,
		string,
		controlservice.ApplyBuilder,
	) (control.ResourceHead, error)
}

type ExpertApplyProjection = controlservice.ExpertApplyProjection
type ExpertApplyMutation = controlservice.ExpertApplyMutation
type ExpertApplyBuilder = controlservice.ExpertApplyBuilder
type AppliedExpert = controlservice.AppliedExpert

type ExpertApplyRepository interface {
	RunExpertApplyTransaction(
		context.Context,
		control.Scope,
		string,
		ExpertApplyBuilder,
	) (AppliedExpert, error)
}

type WorkflowApplyProjection = controlservice.WorkflowApplyProjection
type WorkflowApplyMutation = controlservice.WorkflowApplyMutation
type WorkflowApplyBuilder = controlservice.WorkflowApplyBuilder
type AppliedWorkflow = controlservice.AppliedWorkflow

type WorkflowApplyRepository interface {
	RunWorkflowApplyTransaction(
		context.Context,
		control.Scope,
		string,
		WorkflowApplyBuilder,
	) (AppliedWorkflow, error)
}

type GoalLoopApplyProjection = controlservice.GoalLoopApplyProjection
type GoalLoopApplyMutation = controlservice.GoalLoopApplyMutation
type GoalLoopApplyBuilder = controlservice.GoalLoopApplyBuilder
type AppliedGoalLoop = controlservice.AppliedGoalLoop

type GoalLoopApplyRepository interface {
	RunGoalLoopApplyTransaction(
		context.Context,
		control.Scope,
		string,
		GoalLoopApplyBuilder,
	) (AppliedGoalLoop, error)
}
