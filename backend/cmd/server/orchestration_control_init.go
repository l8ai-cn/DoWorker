package main

import (
	"fmt"
	"time"

	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/infra"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	"github.com/anthropics/agentsmesh/backend/internal/service/organization"
	"github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const orchestrationPlanTTL = 15 * time.Minute

type orchestrationControlServices struct {
	control             *controlservice.Service
	bindingApply        *workerplanner.BindingApplyService
	workerTemplateApply *workerplanner.WorkerTemplateApplyService
	promptApply         *workerplanner.PromptApplyService
	expertApply         *workerplanner.ExpertApplyService
	workflowApply       *workerplanner.WorkflowApplyService
	goalLoopApply       *workerplanner.GoalLoopApplyService
	workerApplyRuntime  orchestrationWorkerApplyRuntime
}

func initializeOrchestrationControl(
	db *gorm.DB,
	organizations *organization.Service,
	workerCreation *workercreation.Service,
) (orchestrationControlServices, error) {
	if db == nil || organizations == nil || workerCreation == nil {
		return orchestrationControlServices{}, fmt.Errorf(
			"orchestration control dependencies are incomplete",
		)
	}
	registry := resource.NewRegistry()
	if err := resource.RegisterWorkerSchemas(registry); err != nil {
		return orchestrationControlServices{}, fmt.Errorf(
			"register orchestration schemas: %w",
			err,
		)
	}
	if err := resource.RegisterDefinitionSchemas(registry); err != nil {
		return orchestrationControlServices{}, fmt.Errorf(
			"register orchestration definition schemas: %w",
			err,
		)
	}
	repository := infra.NewOrchestrationResourceRepository(db)
	authorizer := controlservice.NewMemberAuthorizer(organizations)
	references, err := controlservice.NewRepositoryReferenceResolver(
		repository,
		authorizer,
	)
	if err != nil {
		return orchestrationControlServices{}, err
	}
	bindings, err := workerplanner.NewResourceBindingResolver(
		registry,
		repository,
		authorizer,
	)
	if err != nil {
		return orchestrationControlServices{}, err
	}
	compiler, err := workerplanner.NewWorkerCreationCompiler(workerCreation)
	if err != nil {
		return orchestrationControlServices{}, err
	}
	workerTemplate, err := workerplanner.NewWorkerTemplatePlanner(
		bindings,
		compiler,
	)
	if err != nil {
		return orchestrationControlServices{}, err
	}
	planners, required, err := orchestrationTargetPlanners(
		workerTemplate,
		bindings,
	)
	if err != nil {
		return orchestrationControlServices{}, err
	}
	control, err := controlservice.NewService(controlservice.ServiceDeps{
		Registry: registry, Repository: repository,
		Authorizer: authorizer, References: references,
		Planners: planners, RequiredTypes: required,
		Clock: time.Now, IDGenerator: uuid.NewString,
		PlanTTL: orchestrationPlanTTL,
	})
	if err != nil {
		return orchestrationControlServices{}, err
	}
	apply, err := initializeOrchestrationApplyServices(
		registry,
		repository,
		bindings,
	)
	if err != nil {
		return orchestrationControlServices{}, err
	}
	return orchestrationControlServices{
		control:             control,
		bindingApply:        apply.binding,
		workerTemplateApply: apply.workerTemplate,
		promptApply:         apply.prompt,
		expertApply:         apply.expert,
		workflowApply:       apply.workflow,
		goalLoopApply:       apply.goalLoop,
		workerApplyRuntime:  apply.workerRuntime,
	}, nil
}

func attachOrchestrationControl(
	services *serviceContainer,
	db *gorm.DB,
) error {
	if services == nil {
		return fmt.Errorf("service container is required")
	}
	orchestration, err := initializeOrchestrationControl(
		db,
		services.org,
		services.workerCreation,
	)
	if err != nil {
		return err
	}
	services.orchestration = orchestration.control
	services.bindingApply = orchestration.bindingApply
	services.workerTemplateApply = orchestration.workerTemplateApply
	services.promptApply = orchestration.promptApply
	services.expertApply = orchestration.expertApply
	services.workflowApply = orchestration.workflowApply
	services.goalLoopApply = orchestration.goalLoopApply
	services.workerApplyRuntime = orchestration.workerApplyRuntime
	return nil
}

func orchestrationTargetPlanners(
	workerTemplate controlservice.TargetPlanner,
	definitions workerplanner.DefinitionResolver,
) ([]controlservice.TargetPlanner, []resource.TypeMeta, error) {
	kinds := []string{
		resource.KindModelBinding,
		resource.KindRepository,
		resource.KindSkill,
		resource.KindKnowledgeBase,
		resource.KindEnvironmentBundle,
		resource.KindComputeTarget,
		resource.KindResourceProfile,
		resource.KindToolBinding,
	}
	definitionKinds := []string{
		resource.KindPrompt,
		resource.KindWorker,
		resource.KindExpert,
		resource.KindWorkflow,
		resource.KindGoalLoop,
	}
	planners := make(
		[]controlservice.TargetPlanner,
		0,
		len(kinds)+len(definitionKinds)+1,
	)
	required := make(
		[]resource.TypeMeta,
		0,
		len(kinds)+len(definitionKinds)+1,
	)
	for _, kind := range kinds {
		planner, err := workerplanner.NewResourceBindingPlanner(kind)
		if err != nil {
			return nil, nil, err
		}
		planners = append(planners, planner)
		required = append(required, planner.TypeMeta())
	}
	planners = append(planners, workerTemplate)
	required = append(required, workerTemplate.TypeMeta())
	for _, kind := range definitionKinds {
		planner, err := workerplanner.NewDefinitionPlanner(kind, definitions)
		if err != nil {
			return nil, nil, err
		}
		planners = append(planners, planner)
		required = append(required, planner.TypeMeta())
	}
	return planners, required, nil
}
