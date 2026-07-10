package airesourceconnect

import (
	"context"

	"connectrpc.com/connect"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	service "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	aiv1 "github.com/anthropics/agentsmesh/proto/gen/go/ai_resource/v1"
)

func (s *Server) CreateResource(ctx context.Context, req *connect.Request[aiv1.CreateResourceRequest]) (*connect.Response[aiv1.ModelResource], error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	spec := req.Msg.Resource
	input := service.CreateResourceInput{ConnectionID: req.Msg.GetConnectionId()}
	if spec != nil {
		input.Identifier = slugkit.Slug(spec.GetIdentifier())
		input.ModelID, input.DisplayName = spec.GetModelId(), spec.GetDisplayName()
		input.Modalities = modalitiesFromProto(spec.Modalities)
		input.Capabilities = capabilitiesFromProto(spec.Capabilities)
	}
	view, err := s.service.CreateResource(ctx, actor, input)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(resourceToProto(view)), nil
}

func (s *Server) UpdateResource(ctx context.Context, req *connect.Request[aiv1.UpdateResourceRequest]) (*connect.Response[aiv1.ModelResource], error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	spec := req.Msg.Resource
	input := service.UpdateResourceInput{}
	if spec != nil {
		input.ModelID, input.DisplayName = spec.GetModelId(), spec.GetDisplayName()
		input.Modalities = modalitiesFromProto(spec.Modalities)
		input.Capabilities = capabilitiesFromProto(spec.Capabilities)
	}
	view, err := s.service.UpdateResource(ctx, actor, req.Msg.GetResourceId(), input)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(resourceToProto(view)), nil
}

func (s *Server) SetResourceEnabled(ctx context.Context, req *connect.Request[aiv1.SetResourceEnabledRequest]) (*connect.Response[aiv1.MutationResponse], error) {
	return mutation(ctx, func(actor service.Actor) error {
		return s.service.SetResourceEnabled(ctx, actor, req.Msg.GetResourceId(), req.Msg.GetEnabled())
	})
}

func (s *Server) DeleteResource(ctx context.Context, req *connect.Request[aiv1.DeleteResourceRequest]) (*connect.Response[aiv1.MutationResponse], error) {
	return mutation(ctx, func(actor service.Actor) error {
		return s.service.DeleteResource(ctx, actor, req.Msg.GetResourceId())
	})
}

func (s *Server) SetDefault(ctx context.Context, req *connect.Request[aiv1.SetDefaultRequest]) (*connect.Response[aiv1.MutationResponse], error) {
	return mutation(ctx, func(actor service.Actor) error {
		return s.service.SetDefault(ctx, actor, req.Msg.GetResourceId(), domain.Modality(req.Msg.GetModality()))
	})
}
