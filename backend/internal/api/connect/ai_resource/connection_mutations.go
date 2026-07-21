package airesourceconnect

import (
	"context"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	service "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	"github.com/l8ai-cn/agentcloud/backend/pkg/slugkit"
	aiv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/ai_resource/v1"
)

func (s *Server) CreatePersonalConnection(ctx context.Context, req *connect.Request[aiv1.CreatePersonalConnectionRequest]) (*connect.Response[aiv1.ProviderConnection], error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.createConnection(ctx, actor, domain.OwnerScopeUser, actor.UserID, req.Msg.GetIdentifier(), req.Msg.GetProviderKey(), req.Msg.GetName(), req.Msg.GetBaseUrl(), req.Msg.Credentials)
}

func (s *Server) CreateOrganizationConnection(ctx context.Context, req *connect.Request[aiv1.CreateOrganizationConnectionRequest]) (*connect.Response[aiv1.ProviderConnection], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgs)
	if err != nil {
		return nil, err
	}
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.createConnection(ctx, actor, domain.OwnerScopeOrg, org.GetID(), req.Msg.GetIdentifier(), req.Msg.GetProviderKey(), req.Msg.GetName(), req.Msg.GetBaseUrl(), req.Msg.Credentials)
}

func (s *Server) createConnection(ctx context.Context, actor service.Actor, scope domain.OwnerScope, ownerID int64, identifier, providerKey, name, baseURL string, credentials map[string]string) (*connect.Response[aiv1.ProviderConnection], error) {
	view, err := s.service.CreateConnection(ctx, actor, service.CreateConnectionInput{
		OwnerScope: scope, OwnerID: ownerID, Identifier: slugkit.Slug(identifier), ProviderKey: slugkit.Slug(providerKey),
		Name: name, BaseURL: baseURL, Credentials: credentials,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(connectionToProto(view)), nil
}

func (s *Server) UpdateConnection(ctx context.Context, req *connect.Request[aiv1.UpdateConnectionRequest]) (*connect.Response[aiv1.ProviderConnection], error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	input := service.UpdateConnectionInput{}
	if update := req.Msg.Connection; update != nil {
		input.Name, input.BaseURL = update.GetName(), update.GetBaseUrl()
		if update.GetHasCredentials() {
			input.Credentials = update.Credentials
			if input.Credentials == nil {
				input.Credentials = map[string]string{}
			}
		}
	}
	view, err := s.service.UpdateConnection(ctx, actor, req.Msg.GetConnectionId(), input)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(connectionToProto(view)), nil
}

func (s *Server) RotateConnectionCredentials(ctx context.Context, req *connect.Request[aiv1.RotateConnectionCredentialsRequest]) (*connect.Response[aiv1.MutationResponse], error) {
	return mutation(ctx, func(actor service.Actor) error {
		return s.service.RotateConnectionCredentials(ctx, actor, req.Msg.GetConnectionId(), req.Msg.Credentials)
	})
}

func (s *Server) SetConnectionEnabled(ctx context.Context, req *connect.Request[aiv1.SetConnectionEnabledRequest]) (*connect.Response[aiv1.MutationResponse], error) {
	return mutation(ctx, func(actor service.Actor) error {
		return s.service.SetConnectionEnabled(ctx, actor, req.Msg.GetConnectionId(), req.Msg.GetEnabled())
	})
}

func (s *Server) ValidateConnection(ctx context.Context, req *connect.Request[aiv1.ValidateConnectionRequest]) (*connect.Response[aiv1.MutationResponse], error) {
	return mutation(ctx, func(actor service.Actor) error {
		return s.service.ValidateConnection(ctx, actor, req.Msg.GetConnectionId())
	})
}

func (s *Server) DeleteConnection(ctx context.Context, req *connect.Request[aiv1.DeleteConnectionRequest]) (*connect.Response[aiv1.MutationResponse], error) {
	return mutation(ctx, func(actor service.Actor) error {
		return s.service.DeleteConnection(ctx, actor, req.Msg.GetConnectionId())
	})
}

func mutation(ctx context.Context, action func(service.Actor) error) (*connect.Response[aiv1.MutationResponse], error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := action(actor); err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&aiv1.MutationResponse{}), nil
}
