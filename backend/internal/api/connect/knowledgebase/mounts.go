package knowledgebaseconnect

import (
	"context"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	kbservice "github.com/l8ai-cn/agentcloud/backend/internal/service/knowledgebase"
	kbv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/knowledgebase/v1"
)

func (s *Server) SetAgentMounts(
	ctx context.Context, req *connect.Request[kbv1.SetAgentMountsRequest],
) (*connect.Response[kbv1.SetAgentMountsResponse], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	kb, err := s.svc.GetBySlug(ctx, org.GetID(), req.Msg.GetSlug())
	if err != nil {
		return nil, mapKBError(err)
	}
	inputs := make([]kbservice.AgentMountInput, 0, len(req.Msg.GetMounts()))
	for _, m := range req.Msg.GetMounts() {
		inputs = append(inputs, kbservice.AgentMountInput{
			AgentSlug: m.GetAgentSlug(),
			Mode:      m.GetMode(),
		})
	}
	if err := s.svc.SetAgentMounts(ctx, org.GetID(), kb.ID, inputs); err != nil {
		return nil, mapKBError(err)
	}
	return connect.NewResponse(&kbv1.SetAgentMountsResponse{}), nil
}

func (s *Server) ListAgentMounts(
	ctx context.Context, req *connect.Request[kbv1.ListAgentMountsRequest],
) (*connect.Response[kbv1.ListAgentMountsResponse], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	kb, err := s.svc.GetBySlug(ctx, org.GetID(), req.Msg.GetSlug())
	if err != nil {
		return nil, mapKBError(err)
	}
	mounts, err := s.svc.ListAgentMounts(ctx, org.GetID(), kb.ID)
	if err != nil {
		return nil, mapKBError(err)
	}
	out := make([]*kbv1.AgentMount, 0, len(mounts))
	for _, m := range mounts {
		out = append(out, &kbv1.AgentMount{AgentSlug: m.AgentSlug, Mode: m.Mode})
	}
	return connect.NewResponse(&kbv1.ListAgentMountsResponse{Mounts: out}), nil
}
