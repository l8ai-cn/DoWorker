package knowledgebaseconnect

import (
	"context"
	"encoding/json"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	kbservice "github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase"
	kbv1 "github.com/anthropics/agentsmesh/proto/gen/go/knowledgebase/v1"
)

func (s *Server) ListKnowledgeBases(
	ctx context.Context, req *connect.Request[kbv1.ListKnowledgeBasesRequest],
) (*connect.Response[kbv1.ListKnowledgeBasesResponse], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	kbs, err := s.svc.List(ctx, org.GetID(), req.Msg.GetSourceType())
	if err != nil {
		return nil, mapKBError(err)
	}
	items := make([]*kbv1.KnowledgeBase, 0, len(kbs))
	for _, kb := range kbs {
		items = append(items, toProtoKnowledgeBase(kb))
	}
	total := int64(len(items))
	return connect.NewResponse(&kbv1.ListKnowledgeBasesResponse{
		Items: items, Total: total, Limit: int32(total), Offset: 0,
	}), nil
}

func (s *Server) GetKnowledgeBase(
	ctx context.Context, req *connect.Request[kbv1.GetKnowledgeBaseRequest],
) (*connect.Response[kbv1.KnowledgeBase], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	kb, err := s.svc.GetBySlug(ctx, org.GetID(), req.Msg.GetSlug())
	if err != nil {
		return nil, mapKBError(err)
	}
	return connect.NewResponse(toProtoKnowledgeBase(kb)), nil
}

func (s *Server) CreateKnowledgeBase(
	ctx context.Context, req *connect.Request[kbv1.CreateKnowledgeBaseRequest],
) (*connect.Response[kbv1.KnowledgeBase], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	tenant := middleware.GetTenant(ctx)
	params := &kbservice.CreateParams{
		OrganizationID:  org.GetID(),
		CreatedByUserID: tenant.UserID,
		Name:            req.Msg.GetName(),
		Description:     req.Msg.GetDescription(),
		SourceType:      req.Msg.GetSourceType(),
	}
	if raw := req.Msg.GetSourceConfigJson(); raw != "" {
		if !json.Valid([]byte(raw)) {
			return nil, connect.NewError(connect.CodeInvalidArgument, kbservice.ErrInvalidInput)
		}
		params.SourceConfig = json.RawMessage(raw)
	}
	kb, err := s.svc.Create(ctx, params)
	if err != nil {
		return nil, mapKBError(err)
	}
	return connect.NewResponse(toProtoKnowledgeBase(kb)), nil
}

func (s *Server) UpdateKnowledgeBase(
	ctx context.Context, req *connect.Request[kbv1.UpdateKnowledgeBaseRequest],
) (*connect.Response[kbv1.KnowledgeBase], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	kb, err := s.svc.GetBySlug(ctx, org.GetID(), req.Msg.GetSlug())
	if err != nil {
		return nil, mapKBError(err)
	}
	sourceConfig, err := parseOptionalSourceConfig(req.Msg.GetSourceConfigJson())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	updated, err := s.svc.Update(ctx, org.GetID(), kb.ID, &kbservice.UpdateParams{
		Name:         req.Msg.Name,
		Description:  req.Msg.Description,
		SourceConfig: sourceConfig,
	})
	if err != nil {
		return nil, mapKBError(err)
	}
	return connect.NewResponse(toProtoKnowledgeBase(updated)), nil
}

func parseOptionalSourceConfig(raw string) (json.RawMessage, error) {
	if raw == "" {
		return nil, nil
	}
	if !json.Valid([]byte(raw)) {
		return nil, kbservice.ErrInvalidInput
	}
	return json.RawMessage(raw), nil
}

func (s *Server) DeleteKnowledgeBase(
	ctx context.Context, req *connect.Request[kbv1.DeleteKnowledgeBaseRequest],
) (*connect.Response[kbv1.DeleteKnowledgeBaseResponse], error) {
	ctx, org, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	kb, err := s.svc.GetBySlug(ctx, org.GetID(), req.Msg.GetSlug())
	if err != nil {
		return nil, mapKBError(err)
	}
	if err := s.svc.Delete(ctx, org.GetID(), kb.ID); err != nil {
		return nil, mapKBError(err)
	}
	return connect.NewResponse(&kbv1.DeleteKnowledgeBaseResponse{}), nil
}
