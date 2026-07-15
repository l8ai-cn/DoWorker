package expert

import (
	"context"
	"errors"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
)

func (s *Service) ListMarketApplications(
	ctx context.Context,
) ([]MarketApplication, error) {
	if s.market == nil {
		return nil, ErrMarketUnavailable
	}
	applications := make([]expertmarket.Application, 0)
	for offset := 0; ; {
		page, total, err := s.market.ListApplications(
			ctx,
			expertmarket.ApplicationListFilter{Limit: 100, Offset: offset},
		)
		if err != nil {
			return nil, err
		}
		applications = append(applications, page...)
		offset += len(page)
		if len(page) == 0 || int64(offset) >= total {
			break
		}
	}
	items := make([]MarketApplication, 0, len(applications))
	for index := range applications {
		application := &applications[index]
		if application.LatestPublishedReleaseID == nil {
			continue
		}
		release, err := s.market.GetReleaseByID(
			ctx,
			*application.LatestPublishedReleaseID,
		)
		if err != nil {
			return nil, err
		}
		if release.Status != expertmarket.ReleaseStatusPublished {
			continue
		}
		item, err := marketApplicationView(application, release)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) GetPublishedMarketApplication(
	ctx context.Context,
	slug string,
) (*PublishedMarketApplication, error) {
	if s.market == nil {
		return nil, ErrMarketUnavailable
	}
	application, err := s.market.GetApplicationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, expertmarket.ErrNotFound) {
			return nil, ErrMarketApplicationNotFound
		}
		return nil, err
	}
	if application.LatestPublishedReleaseID == nil {
		return nil, ErrMarketApplicationNotFound
	}
	release, err := s.market.GetReleaseByID(
		ctx,
		*application.LatestPublishedReleaseID,
	)
	if err != nil {
		return nil, err
	}
	if release.Status != expertmarket.ReleaseStatusPublished {
		return nil, ErrMarketApplicationNotFound
	}
	return &PublishedMarketApplication{
		Application: *application,
		Release:     *release,
	}, nil
}

func (s *Service) ListPublisherMarketReleases(
	ctx context.Context,
	organizationID int64,
	limit, offset int,
) ([]expertmarket.Release, int64, error) {
	if s.market == nil {
		return nil, 0, ErrMarketUnavailable
	}
	releases, total, err := s.market.ListReleases(ctx, expertmarket.ReleaseListFilter{
		PublisherOrganizationID: &organizationID,
		Limit:                   limit,
		Offset:                  offset,
	})
	if err != nil {
		return nil, 0, err
	}
	for index := range releases {
		application, getErr := s.market.GetApplicationByID(
			ctx,
			releases[index].ApplicationID,
		)
		if getErr != nil {
			return nil, 0, getErr
		}
		releases[index].ApplicationSlug = string(application.Slug)
	}
	return releases, total, nil
}

func (s *Service) ListPendingMarketReleases(
	ctx context.Context,
	limit, offset int,
) ([]expertmarket.Release, int64, error) {
	if s.market == nil {
		return nil, 0, ErrMarketUnavailable
	}
	status := expertmarket.ReleaseStatusPendingReview
	return s.market.ListReleases(ctx, expertmarket.ReleaseListFilter{
		Status: &status,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *Service) MarketUpgradeAvailable(
	ctx context.Context,
	organizationID, applicationID int64,
) (bool, error) {
	if s.market == nil {
		return false, ErrMarketUnavailable
	}
	installed, err := s.store.GetByMarketApplication(
		ctx,
		organizationID,
		applicationID,
	)
	if err != nil {
		return false, err
	}
	application, err := s.market.GetApplicationByID(ctx, applicationID)
	if err != nil {
		return false, err
	}
	return application.LatestPublishedReleaseID != nil &&
		(installed.SourceMarketReleaseID == nil ||
			*installed.SourceMarketReleaseID !=
				*application.LatestPublishedReleaseID), nil
}

func marketApplicationView(
	application *expertmarket.Application,
	release *expertmarket.Release,
) (MarketApplication, error) {
	snapshot, _, err := decodeMarketReleaseSnapshots(release)
	if err != nil {
		return MarketApplication{}, err
	}
	return MarketApplication{
		ID:          application.ID,
		ReleaseID:   release.ID,
		Slug:        string(application.Slug),
		Name:        release.Name,
		Summary:     release.Summary,
		Description: release.Description,
		Category:    release.Category,
		Icon:        release.Icon,
		AgentSlug:   snapshot.AgentSlug,
		SkillSlugs:  append([]string(nil), snapshot.SkillSlugs...),
		Tags:        append([]string(nil), release.Tags...),
		Outcomes:    append([]string(nil), release.Outcomes...),
		Version:     release.Version,
		Featured:    release.Featured,
	}, nil
}
