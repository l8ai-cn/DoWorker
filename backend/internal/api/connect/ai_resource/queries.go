package airesourceconnect

import (
	"context"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	domain "github.com/l8ai-cn/agentcloud/backend/internal/domain/airesource"
	service "github.com/l8ai-cn/agentcloud/backend/internal/service/airesource"
	aiv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/ai_resource/v1"
)

func (s *Server) GetCatalog(ctx context.Context, _ *connect.Request[aiv1.GetCatalogRequest]) (*connect.Response[aiv1.GetCatalogResponse], error) {
	if _, err := actorFromContext(ctx); err != nil {
		return nil, err
	}
	providers := s.service.Catalog()
	items := make([]*aiv1.ProviderDefinition, len(providers))
	for index := range providers {
		items[index] = providerToProto(providers[index])
	}
	return connect.NewResponse(&aiv1.GetCatalogResponse{Providers: items}), nil
}

func (s *Server) ListPersonalConnections(ctx context.Context, _ *connect.Request[aiv1.ListPersonalConnectionsRequest]) (*connect.Response[aiv1.ListConnectionsResponse], error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.service.ListOwnerConnections(ctx, actor, domain.OwnerScopeUser, actor.UserID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&aiv1.ListConnectionsResponse{Connections: connectionsToProto(items)}), nil
}

func (s *Server) ListOrganizationConnections(ctx context.Context, req *connect.Request[aiv1.ListOrganizationConnectionsRequest]) (*connect.Response[aiv1.ListConnectionsResponse], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgs)
	if err != nil {
		return nil, err
	}
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.service.ListOwnerConnections(ctx, actor, domain.OwnerScopeOrg, org.GetID())
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&aiv1.ListConnectionsResponse{Connections: connectionsToProto(items)}), nil
}

func (s *Server) ListPersonalEffectiveResources(ctx context.Context, req *connect.Request[aiv1.ListPersonalEffectiveResourcesRequest]) (*connect.Response[aiv1.ListEffectiveResourcesResponse], error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.listEffective(ctx, actor, 0, req.Msg.Modalities)
}

func (s *Server) ListOrganizationEffectiveResources(ctx context.Context, req *connect.Request[aiv1.ListOrganizationEffectiveResourcesRequest]) (*connect.Response[aiv1.ListEffectiveResourcesResponse], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgs)
	if err != nil {
		return nil, err
	}
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.listEffective(ctx, actor, org.GetID(), req.Msg.Modalities)
}

func (s *Server) listEffective(ctx context.Context, actor service.Actor, orgID int64, values []string) (*connect.Response[aiv1.ListEffectiveResourcesResponse], error) {
	items, err := s.service.ListEffective(ctx, actor, orgID, modalitiesFromProto(values))
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&aiv1.ListEffectiveResourcesResponse{Resources: effectiveToProto(items)}), nil
}
