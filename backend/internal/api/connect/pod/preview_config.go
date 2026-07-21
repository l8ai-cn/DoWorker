package podconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/l8ai-cn/agentcloud/backend/pkg/policy"
	podv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/pod/v1"
)

func (s *Server) UpdatePodPreviewConfig(
	ctx context.Context,
	req *connect.Request[podv1.UpdatePodPreviewConfigRequest],
) (*connect.Response[podv1.UpdatePodPreviewConfigResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	tenant := middleware.GetTenant(ctx)
	podKey := req.Msg.GetPodKey()
	pod, err := s.podSvc.GetPod(ctx, podKey)
	if err != nil {
		return nil, mapServiceError(err)
	}
	subject := policy.NewSubject(tenant.OrganizationID, tenant.UserID, tenant.UserRole)
	resource := s.podResourceWithGrants(ctx, podKey, pod.OrganizationID, pod.CreatedByID)
	if !policy.PodPolicy.AllowWrite(subject, resource) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("forbidden"))
	}
	updated, err := s.podSvc.UpdatePreviewConfig(
		ctx,
		podKey,
		tenant.UserID,
		int(req.Msg.GetPreviewPort()),
		req.Msg.GetPreviewPath(),
	)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&podv1.UpdatePodPreviewConfigResponse{
		Pod: ToProtoPod(updated),
	}), nil
}
