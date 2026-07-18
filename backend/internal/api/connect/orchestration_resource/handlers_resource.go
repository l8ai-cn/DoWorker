package orchestrationresourceconnect

import (
	"context"

	"connectrpc.com/connect"

	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	resourcev1 "github.com/anthropics/agentsmesh/proto/gen/go/orchestration_resource/v1"
)

func (server *Server) GetResource(
	ctx context.Context,
	request *connect.Request[resourcev1.GetResourceRequest],
) (*connect.Response[resourcev1.Resource], error) {
	ctx, scope, err := server.resolveScope(ctx, request.Msg)
	if err != nil {
		return nil, err
	}
	target, err := targetFromProto(scope, request.Msg.GetTarget())
	if err != nil {
		return nil, err
	}
	head, err := server.service.GetResource(ctx, scope, target)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(resourceToProto(head)), nil
}

func (server *Server) ListResources(
	ctx context.Context,
	request *connect.Request[resourcev1.ListResourcesRequest],
) (*connect.Response[resourcev1.ListResourcesResponse], error) {
	ctx, scope, err := server.resolveScope(ctx, request.Msg)
	if err != nil {
		return nil, err
	}
	filter, err := listFilterFromProto(request.Msg)
	if err != nil {
		return nil, err
	}
	page, err := server.service.ListResources(ctx, scope, filter)
	if err != nil {
		return nil, mapServiceError(err)
	}
	items := make([]*resourcev1.Resource, len(page.Items))
	for index := range page.Items {
		items[index] = resourceToProto(page.Items[index])
	}
	appliedEnvironmentFilter, err := environmentBundleFilterToProto(
		page.AppliedFilter.EnvironmentBundle,
	)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&resourcev1.ListResourcesResponse{
		Items:                          items,
		Total:                          page.Total,
		Limit:                          int32(page.AppliedFilter.Limit),
		Offset:                         int32(page.AppliedFilter.Offset),
		AppliedEnvironmentBundleFilter: appliedEnvironmentFilter,
		AppliedModelBindingFilter: modelBindingFilterToProto(
			page.AppliedFilter.ModelBinding,
		),
	}), nil
}

func (server *Server) ExportResource(
	ctx context.Context,
	request *connect.Request[resourcev1.ExportResourceRequest],
) (*connect.Response[resourcev1.ExportResourceResponse], error) {
	ctx, scope, err := server.resolveScope(ctx, request.Msg)
	if err != nil {
		return nil, err
	}
	target, err := targetFromProto(scope, request.Msg.GetTarget())
	if err != nil {
		return nil, err
	}
	revision, err := revisionFromProto(request.Msg.Revision)
	if err != nil {
		return nil, err
	}
	format, err := sourceFormatFromProto(request.Msg.GetFormat())
	if err != nil {
		return nil, err
	}
	exported, err := server.service.ExportResource(ctx, service.ExportResourceRequest{
		Scope:    scope,
		Target:   target,
		Revision: revision,
		Format:   format,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&resourcev1.ExportResourceResponse{
		Content: append([]byte(nil), exported.Content...),
	}), nil
}

func (server *Server) GetResourcePlan(
	ctx context.Context,
	request *connect.Request[resourcev1.GetResourcePlanRequest],
) (*connect.Response[resourcev1.ResourcePlan], error) {
	ctx, scope, err := server.resolveScope(ctx, request.Msg)
	if err != nil {
		return nil, err
	}
	planID, err := planIDFromProto(request.Msg.GetPlanId())
	if err != nil {
		return nil, err
	}
	plan, err := server.service.GetResourcePlan(ctx, scope, planID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	response, err := planToProto(plan)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(response), nil
}
