package podconnect

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessionDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/infra/eventbus"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
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
	orchReq := &agentpod.OrchestrateCreatePodRequest{
		OrganizationID:     tenant.OrganizationID,
		UserID:             tenant.UserID,
		RunnerID:           req.Msg.GetRunnerId(),
		AgentSlug:          req.Msg.GetAgentSlug(),
		RepositoryID:       optionalInt64(req.Msg.RepositoryId),
		TicketSlug:         optionalString(req.Msg.TicketSlug),
		Alias:              alias,
		AgentfileLayer:     optionalString(req.Msg.AgentfileLayer),
		AutomationLevel:    req.Msg.GetAutomationLevel(),
		Cols:               req.Msg.GetCols(),
		Rows:               req.Msg.GetRows(),
		SourcePodKey:       req.Msg.GetSourcePodKey(),
		ResumeAgentSession: optionalBool(req.Msg.ResumeAgentSession),
		Perpetual:          req.Msg.GetPerpetual(),
		KnowledgeMounts:    knowledgeMountsFromProto(req.Msg.GetKnowledgeMounts()),
		ModelResourceID:    optionalInt64(req.Msg.ModelResourceId),
		TokenBudget:        optionalInt64(req.Msg.TokenBudget),
		SessionProvision:   &sessionDomain.ProvisionSpec{},
	}
	if req.Msg.WorkerSpec != nil {
		draft, err := workerDraftFromProto(req.Msg.WorkerSpec)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		orchReq.WorkerSpecDraft = &draft
		orchReq.PrepareSession = s.prepareWorkerInitialMessage(
			draft.WorkerSpec.Workspace.InitialTask,
		)
	}
	result, err := s.orchestrator.CreatePod(ctx, orchReq)
	if err != nil {
		return nil, mapServiceError(err)
	}
	s.publishPodCreated(ctx, result.Pod)
	response := &podv1.CreatePodResponse{Pod: ToProtoPod(result.Pod)}
	if result.Warning != "" {
		response.Warning = &result.Warning
	}
	return connect.NewResponse(response), nil
}

func (s *Server) publishPodCreated(ctx context.Context, pod *podDomain.Pod) {
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
	event, err := eventbus.NewEntityEvent(eventbus.EventPodCreated, pod.OrganizationID, "pod", pod.PodKey, data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build pod:created event", "pod_key", pod.PodKey, "error", err)
		return
	}
	if err := s.eventBus.Publish(ctx, event); err != nil {
		slog.ErrorContext(ctx, "failed to publish pod:created event", "pod_key", pod.PodKey, "error", err)
	}
}
