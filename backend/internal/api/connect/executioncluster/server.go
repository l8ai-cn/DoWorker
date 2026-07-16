package executionclusterconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	service "github.com/anthropics/agentsmesh/backend/internal/service/executioncluster"
)

const ServiceName = "proto.execution_cluster.v1.ExecutionClusterService"

const (
	ListExecutionClustersProcedure     = "/" + ServiceName + "/ListExecutionClusters"
	CreateRegistrationCommandProcedure = "/" + ServiceName + "/CreateRegistrationCommand"
)

type Service interface {
	List(context.Context, int64) ([]service.View, error)
	CreateRegistrationCommand(context.Context, int64, int64, int64, string) (service.RegistrationCommand, error)
}

type Server struct {
	service Service
	orgs    middleware.OrganizationService
}

func NewServer(service Service, orgs middleware.OrganizationService) *Server {
	return &Server{service: service, orgs: orgs}
}

func resolveOrganization(ctx context.Context, message interface{ GetOrgSlug() string }, orgs middleware.OrganizationService) (context.Context, *middleware.TenantContext, error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, message, orgs)
	if err != nil {
		return nil, nil, err
	}
	tenant := middleware.GetTenant(ctx)
	if tenant == nil {
		return nil, nil, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	return ctx, tenant, nil
}

func mapServiceError(err error) error {
	if errors.Is(err, service.ErrClusterNotFound) {
		return connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewError(connect.CodeInternal, err)
}
