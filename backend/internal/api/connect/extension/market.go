// Market sub-service handlers — read-only marketplace catalog
// (ListMarketSkills + ListMarketMcpServers). Mirrors
// backend/internal/api/rest/v1/extension.go:231-272. Org-scoped; no admin
// guard (any member of the org can browse marketplace).
package extensionconnect

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	extensionv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/extension/v1"
)

const MarketServiceName = "proto.extension.v1.MarketService"

const (
	ListMarketSkillsProcedure     = "/" + MarketServiceName + "/ListMarketSkills"
	ListMarketMcpServersProcedure = "/" + MarketServiceName + "/ListMarketMcpServers"
)

// MarketServer is a thin handler over extensionservice marketplace queries.
type MarketServer struct {
	*Server
}

func NewMarketServer(srv *Server) *MarketServer { return &MarketServer{Server: srv} }

func (s *MarketServer) ListMarketSkills(
	ctx context.Context, req *connect.Request[extensionv1.ListMarketSkillsRequest],
) (*connect.Response[extensionv1.ListMarketSkillsResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}

	tenant := middleware.GetTenant(ctx)
	skills, err := s.extensionSvc.ListMarketSkills(
		ctx, tenant.OrganizationID, req.Msg.GetQuery(), req.Msg.GetCategory(),
	)
	if err != nil {
		return nil, mapServiceError(err)
	}
	items := make([]*extensionv1.SkillMarketItem, 0, len(skills))
	for i := range skills {
		items = append(items, toProtoSkillMarketItem(&skills[i]))
	}
	return connect.NewResponse(&extensionv1.ListMarketSkillsResponse{
		Items: items,
		Total: int64(len(items)),
	}), nil
}

func (s *MarketServer) ListMarketMcpServers(
	ctx context.Context, req *connect.Request[extensionv1.ListMarketMcpServersRequest],
) (*connect.Response[extensionv1.ListMarketMcpServersResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}

	limit := int(req.Msg.GetLimit())
	if limit <= 0 {
		limit = 50 // mirrors REST default at extension.go:252
	}
	offset := int(req.Msg.GetOffset())
	if offset < 0 {
		offset = 0
	}

	servers, total, err := s.extensionSvc.ListMarketMcpServers(
		ctx, req.Msg.GetQuery(), req.Msg.GetCategory(), limit, offset,
	)
	if err != nil {
		return nil, mapServiceError(err)
	}
	items := make([]*extensionv1.McpMarketItem, 0, len(servers))
	for _, it := range servers {
		items = append(items, toProtoMcpMarketItem(it))
	}
	return connect.NewResponse(&extensionv1.ListMarketMcpServersResponse{
		Items:  items,
		Total:  total,
		Limit:  int32(limit),
		Offset: int32(offset),
	}), nil
}

// MountMarket registers MarketService procedures behind the auth interceptor
// supplied via opts, wired as its own sub-service — mirrors the multi-Service
// pattern used by invitation/billing in the same connect package.
func MountMarket(mux *http.ServeMux, srv *MarketServer, opts ...connect.HandlerOption) {
	mux.Handle(ListMarketSkillsProcedure, connect.NewUnaryHandler(
		ListMarketSkillsProcedure, srv.ListMarketSkills, opts...,
	))
	mux.Handle(ListMarketMcpServersProcedure, connect.NewUnaryHandler(
		ListMarketMcpServersProcedure, srv.ListMarketMcpServers, opts...,
	))
}
