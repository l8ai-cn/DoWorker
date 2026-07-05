package knowledgebaseconnect

import (
	"context"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	kbv1 "github.com/anthropics/agentsmesh/proto/gen/go/knowledgebase/v1"
)

func (s *Server) GetKnowledgeBaseFile(
	ctx context.Context, req *connect.Request[kbv1.GetKnowledgeBaseFileRequest],
) (*connect.Response[kbv1.KnowledgeBaseFile], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	file, err := s.svc.ReadFile(ctx, org.GetID(), req.Msg.GetSlug(), req.Msg.GetPath())
	if err != nil {
		return nil, mapKBError(err)
	}
	return connect.NewResponse(&kbv1.KnowledgeBaseFile{
		Path:    file.Path,
		Content: file.Content,
		Size:    file.Size,
	}), nil
}

func (s *Server) ListKnowledgeBaseDir(
	ctx context.Context, req *connect.Request[kbv1.ListKnowledgeBaseDirRequest],
) (*connect.Response[kbv1.ListKnowledgeBaseDirResponse], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	entries, err := s.svc.ListDir(ctx, org.GetID(), req.Msg.GetSlug(), req.Msg.GetPath())
	if err != nil {
		return nil, mapKBError(err)
	}
	out := make([]*kbv1.KnowledgeBaseDirEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, &kbv1.KnowledgeBaseDirEntry{
			Name: e.Name, Path: e.Path, Type: e.Type, Size: e.Size,
		})
	}
	return connect.NewResponse(&kbv1.ListKnowledgeBaseDirResponse{Entries: out}), nil
}
