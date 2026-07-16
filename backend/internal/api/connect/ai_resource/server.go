package airesourceconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/trace"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	service "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
)

type Service interface {
	Catalog() []domain.ProviderDefinition
	ListOwnerConnections(context.Context, service.Actor, domain.OwnerScope, int64) ([]service.ConnectionView, error)
	ListEffective(context.Context, service.Actor, int64, []domain.Modality) ([]service.EffectiveResourceView, error)
	CreateConnection(context.Context, service.Actor, service.CreateConnectionInput) (service.ConnectionView, error)
	UpdateConnection(context.Context, service.Actor, int64, service.UpdateConnectionInput) (service.ConnectionView, error)
	RotateConnectionCredentials(context.Context, service.Actor, int64, map[string]string) error
	SetConnectionEnabled(context.Context, service.Actor, int64, bool) error
	ValidateConnection(context.Context, service.Actor, int64) error
	DeleteConnection(context.Context, service.Actor, int64) error
	CreateResource(context.Context, service.Actor, service.CreateResourceInput) (service.ResourceView, error)
	UpdateResource(context.Context, service.Actor, int64, service.UpdateResourceInput) (service.ResourceView, error)
	SetResourceEnabled(context.Context, service.Actor, int64, bool) error
	DeleteResource(context.Context, service.Actor, int64) error
	SetDefault(context.Context, service.Actor, int64, domain.Modality) error
}

type OrganizationService interface {
	middleware.OrganizationService
}

type Server struct {
	service Service
	orgs    OrganizationService
}

func NewServer(service Service, orgs OrganizationService) *Server {
	if service == nil || orgs == nil {
		panic("AI resource Connect dependencies are required")
	}
	return &Server{service: service, orgs: orgs}
}

func actorFromContext(ctx context.Context) (service.Actor, error) {
	tenant := middleware.GetTenant(ctx)
	if tenant == nil || tenant.UserID <= 0 {
		return service.Actor{}, connect.NewError(connect.CodeUnauthenticated, errors.New("authentication required"))
	}
	actor := service.Actor{UserID: tenant.UserID}
	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		actor.CorrelationID = span.TraceID().String()
	}
	return actor, nil
}

func mapServiceError(err error) error {
	if err == nil {
		return nil
	}
	code, message := connect.CodeInternal, "AI resource operation failed"
	switch {
	case errors.Is(err, service.ErrProbeUnsupported):
		code, message = connect.CodeUnimplemented, "AI resource provider validation unsupported"
	case errors.Is(err, service.ErrProviderEndpointUnavailable):
		code, message = connect.CodeFailedPrecondition, "AI resource provider endpoint unavailable"
	case errors.Is(err, service.ErrNotFound):
		code, message = connect.CodeNotFound, "AI resource not found"
	case errors.Is(err, service.ErrForbidden):
		code, message = connect.CodePermissionDenied, "AI resource access forbidden"
	case errors.Is(err, service.ErrConflict):
		code, message = connect.CodeAlreadyExists, "AI resource conflict"
	case errors.Is(err, service.ErrInvalidCredentials):
		code, message = connect.CodeInvalidArgument, "invalid AI resource credentials"
	case errors.Is(err, service.ErrInvalidOwner), errors.Is(err, service.ErrInvalidProvider),
		errors.Is(err, service.ErrInvalidEndpoint), errors.Is(err, service.ErrInvalidRequirements),
		errors.Is(err, service.ErrIncompatibleModality), errors.Is(err, service.ErrIncompatibleCapability),
		errors.Is(err, service.ErrIncompatibleProtocolAdapter):
		code, message = connect.CodeInvalidArgument, "invalid AI resource request"
	case errors.Is(err, service.ErrDisabled), errors.Is(err, service.ErrUnhealthy), errors.Is(err, service.ErrUnchecked), errors.Is(err, service.ErrValidation):
		code, message = connect.CodeFailedPrecondition, "AI resource connection validation failed"
	}
	return connect.NewError(code, errors.New(message))
}
