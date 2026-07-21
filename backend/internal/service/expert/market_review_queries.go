package expert

import (
	"context"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
)

func (s *Service) ListMarketReleasesForReview(
	ctx context.Context,
	status expertmarket.ReleaseStatus,
	limit, offset int,
) ([]expertmarket.Release, int64, error) {
	if s.market == nil {
		return nil, 0, ErrMarketUnavailable
	}
	if !status.Valid() {
		return nil, 0, expertmarket.ErrInvalidStatus
	}
	return s.market.ListReleases(ctx, expertmarket.ReleaseListFilter{
		Status: &status,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *Service) GetMarketReleaseForReview(
	ctx context.Context,
	releaseID int64,
) (*expertmarket.Release, error) {
	if s.market == nil {
		return nil, ErrMarketUnavailable
	}
	return s.market.GetReleaseByID(ctx, releaseID)
}
