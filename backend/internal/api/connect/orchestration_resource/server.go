package orchestrationresourceconnect

import (
	"context"
	"errors"
	"reflect"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type Service interface {
	Validate(context.Context, service.ValidateRequest) (service.ValidationResult, error)
	Plan(context.Context, service.PlanRequest) (service.PlanResult, error)
	GetResource(context.Context, control.Scope, control.ResourceTarget) (control.ResourceHead, error)
	ListResources(context.Context, control.Scope, service.ResourceListFilter) (service.ResourceListPage, error)
	ExportResource(context.Context, service.ExportResourceRequest) (service.ResourceExport, error)
	GetResourcePlan(context.Context, control.Scope, string) (control.Plan, error)
}

type Server struct {
	service Service
	orgs    middleware.OrganizationService
}

func NewServer(
	service Service,
	orgs middleware.OrganizationService,
) *Server {
	if isNilDependency(service) || isNilDependency(orgs) {
		panic("orchestration resource Connect dependencies are required")
	}
	return &Server{service: service, orgs: orgs}
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
