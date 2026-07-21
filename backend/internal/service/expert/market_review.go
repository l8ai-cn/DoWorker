package expert

import (
	"context"
	"strings"
	"time"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/expertmarket"
)

func (s *Service) ApproveMarketRelease(
	ctx context.Context,
	req ReviewMarketReleaseRequest,
) (*expertmarket.Release, error) {
	release, err := s.marketReleaseForReview(ctx, req.ReleaseID)
	if err != nil {
		return nil, err
	}
	if s.marketInstallLock == nil {
		return nil, ErrMarketUnavailable
	}
	var approved *expertmarket.Release
	err = s.marketInstallLock.WithinMarketApplicationLock(
		ctx,
		release.ApplicationID,
		func() error {
			var approveErr error
			approved, approveErr = s.approveMarketRelease(ctx, req, release.ApplicationID)
			return approveErr
		},
	)
	return approved, err
}

func (s *Service) approveMarketRelease(
	ctx context.Context,
	req ReviewMarketReleaseRequest,
	applicationID int64,
) (*expertmarket.Release, error) {
	release, err := s.marketReleaseForReview(ctx, req.ReleaseID)
	if err != nil {
		return nil, err
	}
	if release.ApplicationID != applicationID {
		return nil, ErrMarketInvalidTransition
	}
	now := time.Now().UTC()
	pending := expertmarket.ReleaseStatusPendingReview
	if err := s.market.UpdateReleaseLifecycleAndLatest(
		ctx,
		release.ApplicationID,
		release.ID,
		expertmarket.LifecycleUpdate{
			Status:         expertmarket.ReleaseStatusPublished,
			ExpectedStatus: &pending,
			ReviewerUserID: &req.ReviewerUserID,
			ReviewedAt:     &now,
			PublishedAt:    &now,
		},
	); err != nil {
		return nil, err
	}
	return s.market.GetReleaseByID(ctx, release.ID)
}

func (s *Service) RejectMarketRelease(
	ctx context.Context,
	req ReviewMarketReleaseRequest,
) (*expertmarket.Release, error) {
	reason := strings.TrimSpace(req.RejectionReason)
	if reason == "" {
		return nil, ErrMarketRejectionReasonRequired
	}
	release, err := s.marketReleaseForReview(ctx, req.ReleaseID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	pending := expertmarket.ReleaseStatusPendingReview
	if err := s.market.UpdateReleaseLifecycle(
		ctx,
		release.ID,
		expertmarket.LifecycleUpdate{
			Status:          expertmarket.ReleaseStatusRejected,
			ExpectedStatus:  &pending,
			ReviewerUserID:  &req.ReviewerUserID,
			RejectionReason: &reason,
			ReviewedAt:      &now,
			RejectedAt:      &now,
		},
	); err != nil {
		return nil, err
	}
	return s.market.GetReleaseByID(ctx, release.ID)
}

func (s *Service) WithdrawMarketRelease(
	ctx context.Context,
	req WithdrawMarketReleaseRequest,
) (*expertmarket.Release, error) {
	if s.market == nil {
		return nil, ErrMarketUnavailable
	}
	release, err := s.market.GetReleaseByID(ctx, req.ReleaseID)
	if err != nil {
		return nil, err
	}
	application, err := s.market.GetApplicationByID(ctx, release.ApplicationID)
	if err != nil {
		return nil, err
	}
	if application.PublisherOrganizationID != req.PublisherOrganizationID {
		return nil, ErrMarketApplicationOwnership
	}
	if release.Status != expertmarket.ReleaseStatusPublished {
		return nil, ErrMarketInvalidTransition
	}
	if s.marketInstallLock == nil {
		return nil, ErrMarketUnavailable
	}
	var withdrawn *expertmarket.Release
	err = s.marketInstallLock.WithinMarketApplicationLock(
		ctx,
		application.ID,
		func() error {
			var withdrawErr error
			withdrawn, withdrawErr = s.withdrawMarketRelease(
				ctx, req, application.ID,
			)
			return withdrawErr
		},
	)
	return withdrawn, err
}

func (s *Service) withdrawMarketRelease(
	ctx context.Context,
	req WithdrawMarketReleaseRequest,
	applicationID int64,
) (*expertmarket.Release, error) {
	release, err := s.market.GetReleaseByID(ctx, req.ReleaseID)
	if err != nil {
		return nil, err
	}
	if release.ApplicationID != applicationID ||
		release.Status != expertmarket.ReleaseStatusPublished {
		return nil, ErrMarketInvalidTransition
	}
	now := time.Now().UTC()
	published := expertmarket.ReleaseStatusPublished
	if err := s.market.WithdrawReleaseAndRefreshLatest(
		ctx,
		applicationID,
		release.ID,
		expertmarket.LifecycleUpdate{
			Status:         expertmarket.ReleaseStatusWithdrawn,
			ExpectedStatus: &published,
			WithdrawnAt:    &now,
		},
	); err != nil {
		return nil, err
	}
	return s.market.GetReleaseByID(ctx, release.ID)
}

func (s *Service) marketReleaseForReview(
	ctx context.Context,
	releaseID int64,
) (*expertmarket.Release, error) {
	if s.market == nil {
		return nil, ErrMarketUnavailable
	}
	release, err := s.market.GetReleaseByID(ctx, releaseID)
	if err != nil {
		return nil, err
	}
	if release.Status != expertmarket.ReleaseStatusPendingReview {
		return nil, ErrMarketInvalidTransition
	}
	return release, nil
}
