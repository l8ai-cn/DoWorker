package expert

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
)

func (s *Service) createMarketSubmission(
	ctx context.Context,
	req SubmitMarketApplicationRequest,
	application *expertmarket.Application,
	release *expertmarket.Release,
) (*MarketSubmission, error) {
	if application.ID == 0 {
		if err := s.market.CreateSubmission(ctx, application, release); err != nil {
			if errors.Is(err, expertmarket.ErrConflict) {
				return s.retryMarketSubmission(ctx, req, release)
			}
			return nil, err
		}
		return &MarketSubmission{Application: *application, Release: *release}, nil
	}
	return s.createLockedMarketSubmission(ctx, req, application, release)
}

func (s *Service) retryMarketSubmission(
	ctx context.Context,
	req SubmitMarketApplicationRequest,
	release *expertmarket.Release,
) (*MarketSubmission, error) {
	application, err := s.market.GetApplicationBySourceExpert(
		ctx,
		req.OrganizationID,
		req.SourceExpertID,
	)
	if errors.Is(err, expertmarket.ErrNotFound) {
		application, err = s.market.GetApplicationBySlug(ctx, req.Slug)
	}
	if err != nil {
		return nil, err
	}
	return s.createLockedMarketSubmission(ctx, req, application, release)
}

func (s *Service) createLockedMarketSubmission(
	ctx context.Context,
	req SubmitMarketApplicationRequest,
	application *expertmarket.Application,
	release *expertmarket.Release,
) (*MarketSubmission, error) {
	if s.marketInstallLock == nil {
		return nil, ErrMarketUnavailable
	}
	err := s.marketInstallLock.WithinMarketApplicationLock(
		ctx,
		application.ID,
		func() error {
			current, lockErr := s.market.GetApplicationByID(ctx, application.ID)
			if lockErr != nil {
				return lockErr
			}
			if lockErr = validateMarketSubmissionApplication(current, req); lockErr != nil {
				return lockErr
			}
			if lockErr = s.validateMarketReleaseRuntimeContract(
				ctx,
				current,
				release,
			); lockErr != nil {
				return lockErr
			}
			application = current
			return s.market.CreateSubmission(ctx, application, release)
		},
	)
	if err != nil {
		return nil, err
	}
	return &MarketSubmission{Application: *application, Release: *release}, nil
}

func validateMarketSubmissionApplication(
	application *expertmarket.Application,
	req SubmitMarketApplicationRequest,
) error {
	if application.PublisherOrganizationID != req.OrganizationID {
		return ErrMarketApplicationOwnership
	}
	if application.Slug.String() != req.Slug ||
		application.SourceExpertID != req.SourceExpertID {
		return ErrMarketApplicationSlugMismatch
	}
	return nil
}
