package expert

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func (s *Service) SubmitMarketApplication(
	ctx context.Context,
	req SubmitMarketApplicationRequest,
) (*MarketSubmission, error) {
	if s.market == nil || s.marketSkills == nil || s.workerSpecs == nil {
		return nil, ErrMarketUnavailable
	}
	if err := slugkit.ValidateIdentifier(
		"expert_market_applications.slug",
		req.Slug,
	); err != nil {
		return nil, err
	}
	if !validMarketIcon(req.Icon) {
		return nil, fmt.Errorf("market icon %q is unsupported", req.Icon)
	}
	source, err := s.store.GetByID(ctx, req.OrganizationID, req.SourceExpertID)
	if err != nil {
		return nil, err
	}
	if source.WorkerSpecSnapshotID == nil {
		return nil, ErrMarketSourceSnapshotRequired
	}
	specSnapshot, err := s.workerSpecs.GetByID(
		ctx,
		req.OrganizationID,
		*source.WorkerSpecSnapshotID,
	)
	if err != nil {
		return nil, err
	}
	if specSnapshot.ID != *source.WorkerSpecSnapshotID || specSnapshot.OrganizationID != req.OrganizationID {
		return nil, ErrMarketSnapshotInvalid
	}
	if err := validatePortableMarketSpec(specSnapshot.Spec); err != nil {
		return nil, errors.Join(ErrMarketSnapshotInvalid, err)
	}
	skills, err := s.loadMarketSkills(
		ctx,
		specSnapshot.Spec.Workspace.SkillIDs,
		source.SkillSlugs,
	)
	if err != nil {
		return nil, err
	}
	expertSnapshot, workerSnapshot, dependencies, err := encodeMarketSnapshots(
		source,
		specSnapshot,
		skills,
	)
	if err != nil {
		return nil, err
	}
	application, err := s.marketApplicationForSubmission(
		ctx,
		req,
		source.ID,
	)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	release := expertmarket.Release{
		SourceExpertID:          source.ID,
		PublisherOrganizationID: req.OrganizationID,
		PublisherUserID:         req.UserID,
		Status:                  expertmarket.ReleaseStatusPendingReview,
		Name:                    source.Name,
		Summary:                 strings.TrimSpace(req.Summary),
		Description:             strings.TrimSpace(req.Description),
		Category:                strings.TrimSpace(req.Category),
		Icon:                    strings.TrimSpace(req.Icon),
		Tags:                    normalizeMarketStrings(req.Tags),
		Outcomes:                normalizeMarketStrings(req.Outcomes),
		Featured:                req.Featured,
		ExpertSnapshot:          expertSnapshot,
		WorkerSpecSnapshot:      workerSnapshot,
		SkillDependencies:       dependencies,
		SubmittedAt:             &now,
	}
	newApplication := application.ID == 0
	if err := s.market.CreateSubmission(ctx, application, &release); err != nil {
		if errors.Is(err, expertmarket.ErrConflict) && newApplication {
			return s.retryMarketSubmission(ctx, req, &release)
		}
		return nil, err
	}
	return &MarketSubmission{Application: *application, Release: release}, nil
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
	if application.PublisherOrganizationID != req.OrganizationID {
		return nil, ErrMarketApplicationOwnership
	}
	if application.Slug.String() != req.Slug {
		return nil, ErrMarketApplicationSlugMismatch
	}
	if application.SourceExpertID != req.SourceExpertID {
		return nil, ErrMarketApplicationSlugMismatch
	}
	if err := s.market.CreateSubmission(ctx, application, release); err != nil {
		return nil, err
	}
	return &MarketSubmission{Application: *application, Release: *release}, nil
}
func (s *Service) marketApplicationForSubmission(
	ctx context.Context,
	req SubmitMarketApplicationRequest,
	sourceExpertID int64,
) (*expertmarket.Application, error) {
	existing, err := s.market.GetApplicationBySourceExpert(
		ctx,
		req.OrganizationID,
		sourceExpertID,
	)
	if err == nil {
		if existing.Slug.String() != req.Slug {
			return nil, ErrMarketApplicationSlugMismatch
		}
		return existing, nil
	}
	if !errors.Is(err, expertmarket.ErrNotFound) {
		return nil, err
	}
	application, err := s.market.GetApplicationBySlug(ctx, req.Slug)
	if err == nil {
		if application.PublisherOrganizationID != req.OrganizationID {
			return nil, ErrMarketApplicationOwnership
		}
		if application.SourceExpertID != sourceExpertID {
			return nil, ErrMarketApplicationSlugMismatch
		}
		return application, nil
	}
	if !errors.Is(err, expertmarket.ErrNotFound) {
		return nil, err
	}
	return &expertmarket.Application{
		Slug:                    slugkit.Slug(req.Slug),
		PublisherOrganizationID: req.OrganizationID,
		SourceExpertID:          sourceExpertID,
		PublisherUserID:         req.UserID,
		IsOperatorOwned:         req.IsOperatorOwned,
	}, nil
}
