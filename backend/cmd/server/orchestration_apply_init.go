package main

import (
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
)

type orchestrationApplyRepository interface {
	workerplanner.BindingApplyRepository
	workerplanner.WorkerTemplateApplyRepository
	workerplanner.WorkerApplyRepository
	workerplanner.PromptApplyRepository
	workerplanner.ExpertApplyRepository
	workerplanner.WorkflowApplyRepository
	workerplanner.GoalLoopApplyRepository
}

type orchestrationWorkerApplyRuntime struct {
	registry   *resource.Registry
	repository orchestrationApplyRepository
	resolver   workerplanner.DefinitionResolver
}

type orchestrationApplyServices struct {
	binding        *workerplanner.BindingApplyService
	workerTemplate *workerplanner.WorkerTemplateApplyService
	prompt         *workerplanner.PromptApplyService
	expert         *workerplanner.ExpertApplyService
	workflow       *workerplanner.WorkflowApplyService
	goalLoop       *workerplanner.GoalLoopApplyService
	workerRuntime  orchestrationWorkerApplyRuntime
}

func initializeOrchestrationApplyServices(
	registry *resource.Registry,
	repository orchestrationApplyRepository,
	resolver workerplanner.DefinitionResolver,
) (orchestrationApplyServices, error) {
	binding, err := workerplanner.NewBindingApplyService(registry, repository)
	if err != nil {
		return orchestrationApplyServices{}, err
	}
	workerTemplate, err := workerplanner.NewWorkerTemplateApplyService(
		registry,
		repository,
	)
	if err != nil {
		return orchestrationApplyServices{}, err
	}
	prompt, err := workerplanner.NewPromptApplyService(registry, repository)
	if err != nil {
		return orchestrationApplyServices{}, err
	}
	expert, err := workerplanner.NewExpertApplyService(
		registry,
		repository,
		resolver,
	)
	if err != nil {
		return orchestrationApplyServices{}, err
	}
	workflow, err := workerplanner.NewWorkflowApplyService(
		registry,
		repository,
		resolver,
	)
	if err != nil {
		return orchestrationApplyServices{}, err
	}
	goalLoop, err := workerplanner.NewGoalLoopApplyService(
		registry,
		repository,
	)
	if err != nil {
		return orchestrationApplyServices{}, err
	}
	return orchestrationApplyServices{
		binding: binding, workerTemplate: workerTemplate,
		prompt: prompt, expert: expert, workflow: workflow,
		goalLoop: goalLoop,
		workerRuntime: orchestrationWorkerApplyRuntime{
			registry: registry, repository: repository,
			resolver: resolver,
		},
	}, nil
}
