package podconnect

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	eventsv1 "github.com/anthropics/agentsmesh/proto/gen/go/events/v1"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

func (s *Server) CreatePod(
	ctx context.Context,
	req *connect.Request[podv1.CreatePodRequest],
) (*connect.Response[podv1.CreatePodResponse], error) {
	if s.orchestrator == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("pod orchestrator not configured"))
	}
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	tenant := middleware.GetTenant(ctx)
	alias := normalizeAlias(req.Msg.Alias)
	if err := validateAlias(alias); err != nil {
		return nil, err
	}
	orchReq, err := buildCreatePodRequest(req.Msg, tenant)
	if err != nil {
		return nil, err
	}
	result, err := s.orchestrator.CreatePod(ctx, orchReq)
	if err != nil {
		return nil, mapServiceError(err)
	}
	s.publishPodCreated(ctx, result.Pod, orchReq.TicketSlug)
	response := &podv1.CreatePodResponse{Pod: ToProtoPod(result.Pod)}
	if result.Warning != "" {
		response.Warning = &result.Warning
	}
	return connect.NewResponse(response), nil
}

func (s *Server) publishPodCreated(
	ctx context.Context,
	pod *podDomain.Pod,
	ticketSlug *string,
) {
	if s.eventBus == nil || pod == nil {
		return
	}
	data := &eventsv1.PodCreatedEventData{
		PodKey:      pod.PodKey,
		Status:      pod.Status,
		AgentStatus: pod.AgentStatus,
		RunnerId:    pod.RunnerID,
		CreatedById: pod.CreatedByID,
	}
	if pod.TicketID != nil {
		data.TicketId = pod.TicketID
	}
	if ticketSlug != nil {
		data.TicketSlug = *ticketSlug
	}
	event, err := eventbus.NewEntityEvent(eventbus.EventPodCreated, pod.OrganizationID, "pod", pod.PodKey, data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build pod:created event", "pod_key", pod.PodKey, "error", err)
		return
	}
	if err := s.eventBus.Publish(ctx, event); err != nil {
		slog.ErrorContext(ctx, "failed to publish pod:created event", "pod_key", pod.PodKey, "error", err)
	}
}
