package adminconnect

import (
	"context"
	"strings"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	adminv1 "github.com/anthropics/agentsmesh/proto/gen/go/admin/v1"
)

func (s *Server) ListExpertMarketReleases(
	ctx context.Context,
	req *connect.Request[adminv1.ListExpertMarketReleasesRequest],
) (*connect.Response[adminv1.ListExpertMarketReleasesResponse], error) {
	ctx, _, err := interceptors.ResolveSystemAdmin(ctx, s.db)
	if err != nil {
		return nil, err
	}
	status, err := expertMarketStatus(req.Msg.GetStatus())
	if err != nil {
		return nil, err
	}
	limit, offset, err := expertMarketPagination(
		req.Msg.GetLimit(),
		req.Msg.GetOffset(),
	)
	if err != nil {
		return nil, err
	}
	if s.expert == nil {
		return nil, connect.NewError(
			connect.CodeUnavailable,
			expertsvc.ErrMarketUnavailable,
		)
	}
	releases, total, err := s.expert.ListMarketReleasesForReview(
		ctx, status, limit, offset,
	)
	if err != nil {
		return nil, mapExpertMarketError(err)
	}
	items := make([]*adminv1.ExpertMarketRelease, 0, len(releases))
	for index := range releases {
		items = append(items, toProtoExpertMarketRelease(&releases[index]))
	}
	return connect.NewResponse(&adminv1.ListExpertMarketReleasesResponse{
		Items:  items,
		Total:  total,
		Limit:  int32(limit),
		Offset: int32(offset),
	}), nil
}

func (s *Server) GetExpertMarketRelease(
	ctx context.Context,
	req *connect.Request[adminv1.GetExpertMarketReleaseRequest],
) (*connect.Response[adminv1.ExpertMarketRelease], error) {
	ctx, _, err := interceptors.ResolveSystemAdmin(ctx, s.db)
	if err != nil {
		return nil, err
	}
	release, err := s.marketRelease(ctx, req.Msg.GetReleaseId())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(toProtoExpertMarketRelease(release)), nil
}

func (s *Server) ApproveExpertMarketRelease(
	ctx context.Context,
	req *connect.Request[adminv1.ApproveExpertMarketReleaseRequest],
) (*connect.Response[adminv1.ExpertMarketRelease], error) {
	ctx, adminUser, err := interceptors.ResolveSystemAdmin(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if req.Msg.GetReleaseId() <= 0 {
		return nil, invalidExpertMarketArgument("release_id must be positive")
	}
	if s.expert == nil {
		return nil, mapExpertMarketError(expertsvc.ErrMarketUnavailable)
	}
	release, err := s.expert.ApproveMarketRelease(
		ctx,
		expertsvc.ReviewMarketReleaseRequest{
			ReviewerUserID: adminUser.ID,
			ReleaseID:      req.Msg.GetReleaseId(),
		},
	)
	if err != nil {
		return nil, mapExpertMarketError(err)
	}
	return connect.NewResponse(toProtoExpertMarketRelease(release)), nil
}

func (s *Server) RejectExpertMarketRelease(
	ctx context.Context,
	req *connect.Request[adminv1.RejectExpertMarketReleaseRequest],
) (*connect.Response[adminv1.ExpertMarketRelease], error) {
	ctx, adminUser, err := interceptors.ResolveSystemAdmin(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if req.Msg.GetReleaseId() <= 0 {
		return nil, invalidExpertMarketArgument("release_id must be positive")
	}
	reason := strings.TrimSpace(req.Msg.GetReason())
	if reason == "" {
		return nil, invalidExpertMarketArgument("reason is required")
	}
	if s.expert == nil {
		return nil, mapExpertMarketError(expertsvc.ErrMarketUnavailable)
	}
	release, err := s.expert.RejectMarketRelease(
		ctx,
		expertsvc.ReviewMarketReleaseRequest{
			ReviewerUserID:  adminUser.ID,
			ReleaseID:       req.Msg.GetReleaseId(),
			RejectionReason: reason,
		},
	)
	if err != nil {
		return nil, mapExpertMarketError(err)
	}
	return connect.NewResponse(toProtoExpertMarketRelease(release)), nil
}

func (s *Server) marketRelease(
	ctx context.Context,
	releaseID int64,
) (*expertmarket.Release, error) {
	if releaseID <= 0 {
		return nil, invalidExpertMarketArgument("release_id must be positive")
	}
	if s.expert == nil {
		return nil, mapExpertMarketError(expertsvc.ErrMarketUnavailable)
	}
	release, err := s.expert.GetMarketReleaseForReview(ctx, releaseID)
	if err != nil {
		return nil, mapExpertMarketError(err)
	}
	return release, nil
}
