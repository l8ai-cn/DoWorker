package agentworkbenchconnect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	workbenchdomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentworkbench"
	agentworkbenchv2 "github.com/anthropics/agentsmesh/proto/gen/go/agent_workbench/v2"
	"github.com/google/uuid"
)

func (s *Server) GetSessionSnapshot(
	ctx context.Context,
	request *connect.Request[agentworkbenchv2.GetSessionSnapshotRequest],
) (*connect.Response[agentworkbenchv2.SessionSnapshot], error) {
	_, snapshot, err := s.loadAuthorizedSnapshot(
		ctx,
		request.Msg,
		request.Msg.GetSessionId(),
	)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(snapshot), nil
}

func (s *Server) loadAuthorizedSnapshot(
	ctx context.Context,
	request interceptors.OrgScopedRequest,
	sessionID string,
) (context.Context, *agentworkbenchv2.SessionSnapshot, error) {
	ctx, session, err := s.authorizeSession(ctx, request, sessionID)
	if err != nil {
		return ctx, nil, err
	}
	snapshot, err := s.loadSnapshot(ctx, sessionID)
	if err != nil {
		return ctx, nil, err
	}
	authorization, err := viewerAuthorizationFor(ctx, session)
	if err != nil {
		return ctx, nil, err
	}
	if err := authorization.decorateSnapshot(snapshot); err != nil {
		return ctx, nil, internalError(err)
	}
	return withViewerAuthorization(ctx, authorization), snapshot, nil
}

func (s *Server) loadSnapshot(
	ctx context.Context,
	sessionID string,
) (*agentworkbenchv2.SessionSnapshot, error) {
	if s.repository == nil {
		return nil, unavailable("agent workbench repository is unavailable")
	}
	state, err := s.repository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return nil, internalError(err)
	}
	if state == nil {
		state, err = s.ensureSnapshot(ctx, sessionID)
		if err != nil {
			return nil, err
		}
	}
	return decodeSessionState(sessionID, state)
}

func (s *Server) ensureSnapshot(
	ctx context.Context,
	sessionID string,
) (*workbenchdomain.SessionState, error) {
	initial, err := newSessionState(sessionID, uuid.NewString())
	if err != nil {
		return nil, internalError(err)
	}
	state, err := s.repository.EnsureSnapshot(ctx, initial)
	if err != nil {
		return nil, internalError(err)
	}
	if state == nil {
		return nil, dataLoss("agent workbench snapshot was not persisted")
	}
	return state, nil
}
