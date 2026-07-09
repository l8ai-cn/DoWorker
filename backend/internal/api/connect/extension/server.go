// Package extensionconnect hosts Connect-RPC handlers for the extension
// domain (marketplace catalog, per-repo skill/MCP installs) over the binary
// protobuf wire (conventions.md §2.5).
package extensionconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
)

// Server carries the shared dependencies for the extension sub-services
// (MarketService, RepoSkillService, RepoMcpService).
type Server struct {
	extensionSvc *extensionservice.Service
	orgSvc       middleware.OrganizationService
}

func NewServer(extSvc *extensionservice.Service, orgSvc middleware.OrganizationService) *Server {
	return &Server{extensionSvc: extSvc, orgSvc: orgSvc}
}

// requireOrgAdmin gates org-scoped mutations on the admin/owner role.
// ResolveOrgScope already populated TenantContext with the user role.
func requireOrgAdmin(ctx context.Context) error {
	tenant := middleware.GetTenant(ctx)
	if tenant == nil {
		return connect.NewError(connect.CodeUnauthenticated, errors.New("missing tenant context"))
	}
	if tenant.UserRole != "admin" && tenant.UserRole != "owner" {
		return connect.NewError(
			connect.CodePermissionDenied,
			errors.New("organization admin role required"),
		)
	}
	return nil
}

// mapServiceError translates extension-domain sentinels to Connect codes
// per conventions §10.
func mapServiceError(err error) error {
	switch {
	case errors.Is(err, extensionservice.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, extensionservice.ErrForbidden):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, extensionservice.ErrInvalidScope),
		errors.Is(err, extensionservice.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, extensionservice.ErrAlreadyInstalled):
		return connect.NewError(connect.CodeAlreadyExists, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
