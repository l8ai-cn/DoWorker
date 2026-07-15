package expert

import (
	"context"
	"sort"

	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
)

type fakeExpertMarket struct {
	applications  map[int64]expertmarket.Application
	releases      map[int64]expertmarket.Release
	nextAppID     int64
	nextReleaseID int64
}

func newFakeExpertMarket() *fakeExpertMarket {
	return &fakeExpertMarket{
		applications: map[int64]expertmarket.Application{},
		releases:     map[int64]expertmarket.Release{},
	}
}

func (market *fakeExpertMarket) CreateApplication(
	_ context.Context,
	application *expertmarket.Application,
) error {
	for _, existing := range market.applications {
		if existing.Slug == application.Slug {
			return expertmarket.ErrConflict
		}
	}
	market.nextAppID++
	application.ID = market.nextAppID
	market.applications[application.ID] = *application
	return nil
}

func (market *fakeExpertMarket) GetApplicationByID(
	_ context.Context,
	id int64,
) (*expertmarket.Application, error) {
	application, ok := market.applications[id]
	if !ok {
		return nil, expertmarket.ErrNotFound
	}
	copy := application
	return &copy, nil
}

func (market *fakeExpertMarket) GetApplicationBySlug(
	_ context.Context,
	slug string,
) (*expertmarket.Application, error) {
	for _, application := range market.applications {
		if string(application.Slug) == slug {
			copy := application
			return &copy, nil
		}
	}
	return nil, expertmarket.ErrNotFound
}

func (market *fakeExpertMarket) ListApplications(
	_ context.Context,
	filter expertmarket.ApplicationListFilter,
) ([]expertmarket.Application, int64, error) {
	rows := make([]expertmarket.Application, 0)
	for _, application := range market.applications {
		if filter.PublisherOrganizationID != nil &&
			application.PublisherOrganizationID != *filter.PublisherOrganizationID {
			continue
		}
		rows = append(rows, application)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	return rows, int64(len(rows)), nil
}

func (market *fakeExpertMarket) CreateRelease(
	_ context.Context,
	release *expertmarket.Release,
) error {
	market.nextReleaseID++
	release.ID = market.nextReleaseID
	market.releases[release.ID] = cloneMarketRelease(*release)
	return nil
}

func (market *fakeExpertMarket) CreateSubmission(
	ctx context.Context,
	application *expertmarket.Application,
	release *expertmarket.Release,
) error {
	if application.ID == 0 {
		if err := market.CreateApplication(ctx, application); err != nil {
			return err
		}
	}
	latestVersion := 0
	for _, existing := range market.releases {
		if existing.ApplicationID == application.ID &&
			existing.Status == expertmarket.ReleaseStatusPendingReview {
			return expertmarket.ErrPendingReleaseExists
		}
		if existing.ApplicationID == application.ID &&
			existing.Version > latestVersion {
			latestVersion = existing.Version
		}
	}
	release.Version = latestVersion + 1
	release.ApplicationID = application.ID
	return market.CreateRelease(ctx, release)
}

func (market *fakeExpertMarket) GetReleaseByID(
	_ context.Context,
	id int64,
) (*expertmarket.Release, error) {
	release, ok := market.releases[id]
	if !ok {
		return nil, expertmarket.ErrNotFound
	}
	copy := cloneMarketRelease(release)
	return &copy, nil
}

func (market *fakeExpertMarket) ListReleases(
	_ context.Context,
	filter expertmarket.ReleaseListFilter,
) ([]expertmarket.Release, int64, error) {
	rows := make([]expertmarket.Release, 0)
	for _, release := range market.releases {
		if filter.ApplicationID != nil && release.ApplicationID != *filter.ApplicationID {
			continue
		}
		if filter.PublisherOrganizationID != nil &&
			release.PublisherOrganizationID != *filter.PublisherOrganizationID {
			continue
		}
		if filter.Status != nil && release.Status != *filter.Status {
			continue
		}
		rows = append(rows, cloneMarketRelease(release))
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Version == rows[j].Version {
			return rows[i].ID < rows[j].ID
		}
		return rows[i].Version < rows[j].Version
	})
	total := int64(len(rows))
	start := min(filter.Offset, len(rows))
	end := len(rows)
	if filter.Limit > 0 {
		end = min(start+filter.Limit, len(rows))
	}
	return rows[start:end], total, nil
}

func (market *fakeExpertMarket) UpdateReleaseLifecycle(
	_ context.Context,
	releaseID int64,
	update expertmarket.LifecycleUpdate,
) error {
	release, ok := market.releases[releaseID]
	if !ok {
		return expertmarket.ErrNotFound
	}
	if update.ExpectedStatus != nil && release.Status != *update.ExpectedStatus {
		return expertmarket.ErrLifecycleStatusConflict
	}
	applyFakeLifecycle(&release, update)
	market.releases[releaseID] = release
	return nil
}

func (market *fakeExpertMarket) UpdateReleaseLifecycleAndLatest(
	_ context.Context,
	applicationID, releaseID int64,
	update expertmarket.LifecycleUpdate,
) error {
	release, ok := market.releases[releaseID]
	if !ok || release.ApplicationID != applicationID {
		return expertmarket.ErrNotFound
	}
	if update.ExpectedStatus != nil && release.Status != *update.ExpectedStatus {
		return expertmarket.ErrLifecycleStatusConflict
	}
	applyFakeLifecycle(&release, update)
	market.releases[releaseID] = release
	application := market.applications[applicationID]
	if application.LatestPublishedReleaseID == nil ||
		market.releases[*application.LatestPublishedReleaseID].Version < release.Version {
		application.LatestPublishedReleaseID = int64Pointer(releaseID)
		market.applications[applicationID] = application
	}
	return nil
}

func (market *fakeExpertMarket) WithdrawReleaseAndRefreshLatest(
	_ context.Context,
	applicationID, releaseID int64,
	update expertmarket.LifecycleUpdate,
) error {
	release, ok := market.releases[releaseID]
	if !ok || release.ApplicationID != applicationID {
		return expertmarket.ErrNotFound
	}
	if update.ExpectedStatus != nil && release.Status != *update.ExpectedStatus {
		return expertmarket.ErrLifecycleStatusConflict
	}
	applyFakeLifecycle(&release, update)
	market.releases[releaseID] = release
	application := market.applications[applicationID]
	if application.LatestPublishedReleaseID == nil ||
		*application.LatestPublishedReleaseID != releaseID {
		return nil
	}
	var latest *expertmarket.Release
	for id, candidate := range market.releases {
		if id == releaseID || candidate.ApplicationID != applicationID ||
			candidate.Status != expertmarket.ReleaseStatusPublished {
			continue
		}
		if latest == nil || candidate.Version > latest.Version {
			copy := candidate
			latest = &copy
		}
	}
	application.LatestPublishedReleaseID = nil
	if latest != nil {
		application.LatestPublishedReleaseID = int64Pointer(latest.ID)
	}
	market.applications[applicationID] = application
	return nil
}

type fakeMarketSkills struct {
	rows []skilldom.Skill
	err  error
}

func (skills *fakeMarketSkills) ListByIDs(
	_ context.Context,
	ids []int64,
) ([]skilldom.Skill, error) {
	if skills.err != nil {
		return nil, skills.err
	}
	required := map[int64]struct{}{}
	for _, id := range ids {
		required[id] = struct{}{}
	}
	rows := make([]skilldom.Skill, 0)
	for _, skill := range skills.rows {
		if _, ok := required[skill.ID]; ok {
			rows = append(rows, skill)
		}
	}
	return rows, nil
}

func cloneMarketRelease(release expertmarket.Release) expertmarket.Release {
	release.ExpertSnapshot = append([]byte(nil), release.ExpertSnapshot...)
	release.WorkerSpecSnapshot = append([]byte(nil), release.WorkerSpecSnapshot...)
	release.SkillDependencies = append([]byte(nil), release.SkillDependencies...)
	return release
}

func applyFakeLifecycle(
	release *expertmarket.Release,
	update expertmarket.LifecycleUpdate,
) {
	release.Status = update.Status
	release.ReviewerUserID = update.ReviewerUserID
	release.RejectionReason = update.RejectionReason
	release.SubmittedAt = update.SubmittedAt
	release.ReviewedAt = update.ReviewedAt
	release.PublishedAt = update.PublishedAt
	release.RejectedAt = update.RejectedAt
	release.WithdrawnAt = update.WithdrawnAt
}

func int64Pointer(value int64) *int64 {
	return &value
}

var _ expertmarket.Repository = (*fakeExpertMarket)(nil)
