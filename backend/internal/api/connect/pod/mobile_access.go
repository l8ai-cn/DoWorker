package podconnect

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	"github.com/l8ai-cn/agentcloud/backend/pkg/policy"
	podv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/pod/v1"
)

func (s *Server) GetMobileAccessDescriptor(
	ctx context.Context,
	req *connect.Request[podv1.GetMobileAccessDescriptorRequest],
) (*connect.Response[podv1.MobileAccessDescriptor], error) {
	if strings.TrimSpace(s.mobileBaseURL) == "" {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("mobile public base URL not configured"))
	}
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
	if !policy.PodPolicy.AllowRead(subject, resource) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("forbidden"))
	}

	relayAvailable := s.relayManager != nil && s.relayManager.GetHealthyRelayCount() > 0
	consoleAvailable := pod.IsActive() && pod.RunnerID > 0 && relayAvailable
	previewAvailable := consoleAvailable && pod.PreviewPort > 0
	descriptor := &podv1.MobileAccessDescriptor{
		CanonicalUrl:     mobileWorkerURL(s.mobileBaseURL, pod.PodKey),
		PodKey:           pod.PodKey,
		Status:           pod.Status,
		InteractionMode:  pod.InteractionMode,
		ConsoleAvailable: consoleAvailable,
		PreviewAvailable: previewAvailable,
		RelayAvailable:   relayAvailable,
	}
	if pod.PreviewPath != "" {
		descriptor.PreviewPath = &pod.PreviewPath
	}
	return connect.NewResponse(descriptor), nil
}

func mobileWorkerURL(baseURL, podKey string) string {
	return strings.TrimRight(baseURL, "/") + "/workers/" + url.PathEscape(podKey)
}
