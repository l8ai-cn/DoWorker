package agentworkbenchconnect

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	sessiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/anthropics/agentsmesh/backend/pkg/embedtoken"
)

func (s *Server) authorizeSession(
	ctx context.Context,
	request interceptors.OrgScopedRequest,
	sessionID string,
) (context.Context, *sessiondomain.Session, error) {
	if sessionID == "" {
		return ctx, nil, invalidArgument("session_id is required")
	}
	if s.sessions == nil {
		return ctx, nil, unavailable("agent workbench authorization is unavailable")
	}
	if claims := embedAccess(ctx); claims != nil {
		return s.authorizeEmbeddedSession(ctx, request, sessionID, claims)
	}
	if s.orgSvc == nil {
		return ctx, nil, unavailable("agent workbench authorization is unavailable")
	}
	ctx, _, err := interceptors.ResolveOrgScope(ctx, request, s.orgSvc)
	if err != nil {
		return ctx, nil, err
	}
	session, err := s.sessions.Get(ctx, sessionID)
	if errors.Is(err, sessionsvc.ErrNotFound) {
		return ctx, nil, notFound("agent session not found")
	}
	if err != nil {
		return ctx, nil, internalError(err)
	}
	if session == nil ||
		session.ID != sessionID ||
		session.DeletedAt != nil ||
		session.Archived {
		return ctx, nil, notFound("active agent session not found")
	}
	tenant := middleware.GetTenant(ctx)
	if tenant == nil || tenant.UserID == 0 {
		return ctx, nil, unavailable("agent workbench tenant context is unavailable")
	}
	if session.OrganizationID != tenant.OrganizationID {
		return ctx, nil, notFound("agent session not found")
	}
	if session.UserID != tenant.UserID {
		return ctx, nil, permissionDenied("agent session owner access is required")
	}
	return ctx, session, nil
}

func (s *Server) authorizeEmbeddedSession(
	ctx context.Context,
	request interceptors.OrgScopedRequest,
	sessionID string,
	claims *embedtoken.Claims,
) (context.Context, *sessiondomain.Session, error) {
	if !hasEmbedCapability(claims, "read") {
		return ctx, nil, permissionDenied("embedded read capability is required")
	}
	if claims.SessionID != sessionID ||
		request.GetOrgSlug() != claims.OrganizationSlug {
		return ctx, nil, notFound("agent session not found")
	}
	session, err := s.sessions.Get(ctx, sessionID)
	if errors.Is(err, sessionsvc.ErrNotFound) {
		return ctx, nil, notFound("agent session not found")
	}
	if err != nil {
		return ctx, nil, internalError(err)
	}
	if session == nil ||
		session.ID != sessionID ||
		session.OrganizationID != claims.OrganizationID ||
		session.DeletedAt != nil ||
		session.Archived {
		return ctx, nil, notFound("active agent session not found")
	}
	return ctx, session, nil
}
