package executionclusterconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	service "github.com/anthropics/agentsmesh/backend/internal/service/executioncluster"
	"github.com/anthropics/agentsmesh/backend/pkg/protoconv"
	executionclusterv1 "github.com/anthropics/agentsmesh/proto/gen/go/execution_cluster/v1"
)

func (s *Server) ListExecutionClusters(
	ctx context.Context,
	req *connect.Request[executionclusterv1.ListExecutionClustersRequest],
) (*connect.Response[executionclusterv1.ListExecutionClustersResponse], error) {
	ctx, tenant, err := resolveOrganization(ctx, req.Msg, s.orgs)
	if err != nil {
		return nil, err
	}
	views, err := s.service.List(ctx, tenant.OrganizationID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	items := make([]*executionclusterv1.ExecutionCluster, 0, len(views))
	for _, view := range views {
		items = append(items, toProtoCluster(view))
	}
	return connect.NewResponse(&executionclusterv1.ListExecutionClustersResponse{Items: items}), nil
}

func (s *Server) CreateRegistrationCommand(
	ctx context.Context,
	req *connect.Request[executionclusterv1.CreateRegistrationCommandRequest],
) (*connect.Response[executionclusterv1.CreateRegistrationCommandResponse], error) {
	ctx, tenant, err := resolveOrganization(ctx, req.Msg, s.orgs)
	if err != nil {
		return nil, err
	}
	if tenant.UserRole != "owner" && tenant.UserRole != "admin" {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("organization admin role required"))
	}
	if req.Msg.GetClusterId() <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("cluster_id is required"))
	}
	command, err := s.service.CreateRegistrationCommand(
		ctx,
		tenant.OrganizationID,
		tenant.UserID,
		req.Msg.GetClusterId(),
		req.Msg.GetNodeName(),
	)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&executionclusterv1.CreateRegistrationCommandResponse{
		Command:   command.Command,
		ExpiresAt: protoconv.RFC3339(command.ExpiresAt),
	}), nil
}

func toProtoCluster(view service.View) *executionclusterv1.ExecutionCluster {
	result := &executionclusterv1.ExecutionCluster{
		Id:                   view.Cluster.ID,
		Slug:                 view.Cluster.Slug.String(),
		Name:                 view.Cluster.Name,
		Kind:                 view.Cluster.Kind,
		Status:               view.Cluster.Status,
		RunnerCount:          int32(view.RunnerCount),
		OnlineRunnerCount:    int32(view.OnlineRunnerCount),
		AvailableRunnerCount: int32(view.AvailableRunnerCount),
		TunnelStatus:         view.TunnelStatus,
	}
	if view.TunnelLastSeenAt != nil {
		value := protoconv.RFC3339(*view.TunnelLastSeenAt)
		result.TunnelLastSeenAt = &value
	}
	if view.TunnelLastError != nil {
		result.TunnelLastError = view.TunnelLastError
	}
	return result
}
