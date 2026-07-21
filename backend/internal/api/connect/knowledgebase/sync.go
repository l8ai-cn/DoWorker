package knowledgebaseconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	kbv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/knowledgebase/v1"
)

var errSyncWorkerUnavailable = errors.New("knowledge base sync worker unavailable")

func (s *Server) SyncKnowledgeBase(
	ctx context.Context, req *connect.Request[kbv1.SyncKnowledgeBaseRequest],
) (*connect.Response[kbv1.KnowledgeBase], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.syncWorker == nil {
		return nil, connect.NewError(connect.CodeInternal, errSyncWorkerUnavailable)
	}
	if err := s.syncWorker.SyncSingle(ctx, org.GetID(), req.Msg.GetSlug()); err != nil {
		return nil, mapKBError(err)
	}
	kb, err := s.svc.GetBySlug(ctx, org.GetID(), req.Msg.GetSlug())
	if err != nil {
		return nil, mapKBError(err)
	}
	return connect.NewResponse(toProtoKnowledgeBase(kb)), nil
}
