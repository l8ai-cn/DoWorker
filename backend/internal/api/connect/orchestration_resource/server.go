package orchestrationresourceconnect

import (
	"context"
	"errors"
	"reflect"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	service "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationworker"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
)

type Service interface {
	Validate(context.Context, service.ValidateRequest) (service.ValidationResult, error)
	Plan(context.Context, service.PlanRequest) (service.PlanResult, error)
	GetResource(context.Context, control.Scope, control.ResourceTarget) (control.ResourceHead, error)
	GetResourceCapabilities(context.Context, control.Scope, control.ResourceTarget) (service.ResourceCapabilities, error)
	ListResources(context.Context, control.Scope, service.ResourceListFilter) (service.ResourceListPage, error)
	ExportResource(context.Context, service.ExportResourceRequest) (service.ResourceExport, error)
	GetResourcePlan(context.Context, control.Scope, string) (control.Plan, error)
	AuthorizeApply(context.Context, control.Scope, string) error
}

type BindingPlanApplier interface {
	Apply(context.Context, control.Scope, string) (control.ResourceHead, error)
}

type WorkerTemplatePlanApplier interface {
	Apply(
		context.Context,
		control.Scope,
		string,
	) (workerplanner.AppliedWorkerTemplate, error)
}

type WorkerPlanCreator interface {
	Apply(
		context.Context,
		control.Scope,
		string,
	) (workerplanner.AppliedWorker, error)
}

type PromptPlanApplier interface {
	Apply(context.Context, control.Scope, string) (control.ResourceHead, error)
}

type ExpertPlanApplier interface {
	Apply(
		context.Context,
		control.Scope,
		string,
	) (workerplanner.AppliedExpert, error)
}

type WorkflowPlanApplier interface {
	Apply(
		context.Context,
		control.Scope,
		string,
	) (workerplanner.AppliedWorkflow, error)
}

type GoalLoopPlanCreator interface {
	Apply(
		context.Context,
		control.Scope,
		string,
	) (workerplanner.AppliedGoalLoop, error)
}

type Server struct {
	service             Service
	bindingApply        BindingPlanApplier
	workerTemplateApply WorkerTemplatePlanApplier
	workerApply         WorkerPlanCreator
	promptApply         PromptPlanApplier
	expertApply         ExpertPlanApplier
	workflowApply       WorkflowPlanApplier
	goalLoopApply       GoalLoopPlanCreator
	orgs                middleware.OrganizationService
}

func NewServer(
	service Service,
	bindingApply BindingPlanApplier,
	workerTemplateApply WorkerTemplatePlanApplier,
	workerApply WorkerPlanCreator,
	promptApply PromptPlanApplier,
	expertApply ExpertPlanApplier,
	workflowApply WorkflowPlanApplier,
	goalLoopApply GoalLoopPlanCreator,
	orgs middleware.OrganizationService,
) *Server {
	if isNilDependency(service) ||
		isNilDependency(bindingApply) ||
		isNilDependency(workerTemplateApply) ||
		isNilDependency(workerApply) ||
		isNilDependency(promptApply) ||
		isNilDependency(expertApply) ||
		isNilDependency(workflowApply) ||
		isNilDependency(goalLoopApply) ||
		isNilDependency(orgs) {
		panic("orchestration resource Connect dependencies are required")
	}
	return &Server{
		service:             service,
		bindingApply:        bindingApply,
		workerTemplateApply: workerTemplateApply,
		workerApply:         workerApply,
		promptApply:         promptApply,
		expertApply:         expertApply,
		workflowApply:       workflowApply,
		goalLoopApply:       goalLoopApply,
		orgs:                orgs,
	}
}

func isNilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func (server *Server) resolveScope(
	ctx context.Context,
	request interface{ GetOrgSlug() string },
) (context.Context, control.Scope, error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, request, server.orgs)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeInternal {
			return nil, control.Scope{}, connect.NewError(
				connect.CodeInternal,
				errors.New("orchestration resource scope unavailable"),
			)
		}
		return nil, control.Scope{}, err
	}
	tenant := middleware.GetTenant(ctx)
	if tenant == nil {
		return nil, control.Scope{}, connect.NewError(
			connect.CodeInternal,
			errors.New("orchestration resource scope unavailable"),
		)
	}
	scope := control.Scope{
		OrganizationID:   tenant.OrganizationID,
		OrganizationSlug: slugkit.Slug(tenant.OrganizationSlug),
		ActorID:          tenant.UserID,
	}
	if err := scope.Validate(); err != nil {
		return nil, control.Scope{}, connect.NewError(
			connect.CodeInternal,
			errors.New("orchestration resource scope unavailable"),
		)
	}
	return ctx, scope, nil
}
