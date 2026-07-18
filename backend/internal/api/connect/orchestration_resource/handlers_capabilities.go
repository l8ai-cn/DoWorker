package orchestrationresourceconnect

import (
	"context"

	"connectrpc.com/connect"

	resourcev1 "github.com/anthropics/agentsmesh/proto/gen/go/orchestration_resource/v1"
)

func (server *Server) GetResourceCapabilities(
	ctx context.Context,
	request *connect.Request[resourcev1.GetResourceCapabilitiesRequest],
) (*connect.Response[resourcev1.GetResourceCapabilitiesResponse], error) {
	ctx, scope, err := server.resolveScope(ctx, request.Msg)
	if err != nil {
		return nil, err
	}
	target, err := targetFromProto(scope, request.Msg.GetTarget())
	if err != nil {
		return nil, err
	}
	capabilities, err := server.service.GetResourceCapabilities(ctx, scope, target)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&resourcev1.GetResourceCapabilitiesResponse{
		Target: targetToProto(target),
		Capabilities: &resourcev1.ResourceCapabilities{
			Exists:        capabilities.Exists,
			CanViewSource: capabilities.CanViewSource,
			CanReference:  capabilities.CanReference,
			CanPlan:       capabilities.CanPlan,
		},
	}), nil
}
