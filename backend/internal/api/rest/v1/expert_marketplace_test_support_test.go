package v1

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/gin-gonic/gin"
)

type expertMarketplaceAPIFixture struct {
	handler *ExpertHandler
	experts *marketplaceExpertRepository
	market  *marketplaceRepository
}

func newExpertMarketplaceAPIFixture() *expertMarketplaceAPIFixture {
	gin.SetMode(gin.TestMode)
	experts := &marketplaceExpertRepository{rows: map[int64]*expertdom.Expert{}}
	market := &marketplaceRepository{
		applications: map[int64]*expertmarket.Application{},
		releases:     map[int64]*expertmarket.Release{},
	}
	service := expertsvc.NewService(expertsvc.Deps{
		Store:             experts,
		WorkerSpecs:       marketplaceWorkerSpecs{},
		MarketInstallLock: marketplaceLock{},
		Market:            market,
		MarketSkills:      marketplaceSkills{},
	})
	return &expertMarketplaceAPIFixture{
		handler: NewExpertHandler(service),
		experts: experts,
		market:  market,
	}
}

func (fixture *expertMarketplaceAPIFixture) perform(
	method, target, body string,
) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	context.Request = httptest.NewRequest(method, target, reader)
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set("tenant", &middleware.TenantContext{
		OrganizationID: 41,
		UserID:         73,
	})
	path := context.Request.URL.Path
	switch {
	case strings.HasSuffix(path, "/market-submissions"):
		context.Params = gin.Params{{Key: "expertSlug", Value: pathSegment(path, 1)}}
		fixture.handler.SubmitMarketApplication(context)
	case path == "/marketplace/submissions":
		fixture.handler.ListMarketSubmissions(context)
	case strings.HasSuffix(path, "/withdraw"):
		context.Params = gin.Params{{Key: "releaseID", Value: pathSegment(path, 2)}}
		fixture.handler.WithdrawMarketRelease(context)
	case strings.HasSuffix(path, "/market-upgrade") && method == http.MethodGet:
		context.Params = gin.Params{{Key: "expertSlug", Value: pathSegment(path, 1)}}
		fixture.handler.GetMarketUpgradeAvailability(context)
	default:
		context.Params = gin.Params{{Key: "expertSlug", Value: pathSegment(path, 1)}}
		fixture.handler.UpgradeMarketApplication(context)
	}
	return recorder
}

func (fixture *expertMarketplaceAPIFixture) errorResponse(
	err error,
) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	fixture.handler.marketplaceError(context, err, "request failed")
	return recorder
}

func pathSegment(path string, index int) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if index >= len(parts) {
		return ""
	}
	return parts[index]
}

type marketplaceExpertRepository struct {
	rows               map[int64]*expertdom.Expert
	lastOrganizationID int64
	lastExpertID       int64
	lastSlug           string
}

func (repository *marketplaceExpertRepository) GetByID(
	_ context.Context, organizationID, expertID int64,
) (*expertdom.Expert, error) {
	repository.lastOrganizationID = organizationID
	repository.lastExpertID = expertID
	row := repository.rows[expertID]
	if row == nil || row.OrganizationID != organizationID {
		return nil, expertdom.ErrNotFound
	}
	return row, nil
}

func (repository *marketplaceExpertRepository) GetBySlug(
	_ context.Context, organizationID int64, slug string,
) (*expertdom.Expert, error) {
	repository.lastOrganizationID = organizationID
	repository.lastSlug = slug
	for _, row := range repository.rows {
		if row.OrganizationID == organizationID && row.Slug == slug {
			return row, nil
		}
	}
	return nil, expertdom.ErrNotFound
}

func (repository *marketplaceExpertRepository) GetByMarketApplication(
	_ context.Context, organizationID, applicationID int64,
) (*expertdom.Expert, error) {
	for _, row := range repository.rows {
		if row.OrganizationID == organizationID &&
			row.SourceMarketApplicationID != nil &&
			*row.SourceMarketApplicationID == applicationID {
			return row, nil
		}
	}
	return nil, expertdom.ErrNotFound
}

func (*marketplaceExpertRepository) Create(context.Context, *expertdom.Expert) error {
	return nil
}
func (*marketplaceExpertRepository) Update(context.Context, *expertdom.Expert) error {
	return nil
}
func (*marketplaceExpertRepository) Delete(context.Context, int64, int64) error {
	return nil
}
func (*marketplaceExpertRepository) UpdateMarketRelease(
	context.Context, int64, int64, int64, expertdom.MarketReleaseUpdate,
) error {
	return nil
}
func (*marketplaceExpertRepository) SlugExists(
	context.Context, int64, string, int64,
) (bool, error) {
	return false, nil
}
func (*marketplaceExpertRepository) List(
	context.Context, int64, int, int,
) ([]expertdom.Expert, int64, error) {
	return nil, 0, nil
}
func (*marketplaceExpertRepository) RecordRun(
	context.Context, int64, int64, time.Time,
) error {
	return nil
}

type marketplaceRepository struct {
	applications map[int64]*expertmarket.Application
	releases     map[int64]*expertmarket.Release
}

func (*marketplaceRepository) CreateApplication(
	context.Context, *expertmarket.Application,
) error {
	return nil
}
func (*marketplaceRepository) CreateRelease(context.Context, *expertmarket.Release) error {
	return nil
}
func (*marketplaceRepository) CreateSubmission(
	context.Context, *expertmarket.Application, *expertmarket.Release,
) error {
	return nil
}

func (repository *marketplaceRepository) GetApplicationByID(
	_ context.Context, id int64,
) (*expertmarket.Application, error) {
	row := repository.applications[id]
	if row == nil {
		return nil, expertmarket.ErrNotFound
	}
	return row, nil
}

func (repository *marketplaceRepository) GetApplicationBySlug(
	_ context.Context, slug string,
) (*expertmarket.Application, error) {
	for _, row := range repository.applications {
		if string(row.Slug) == slug {
			return row, nil
		}
	}
	return nil, expertmarket.ErrNotFound
}

func (repository *marketplaceRepository) GetApplicationBySourceExpert(
	_ context.Context, organizationID, sourceExpertID int64,
) (*expertmarket.Application, error) {
	for _, row := range repository.applications {
		if row.PublisherOrganizationID == organizationID &&
			row.SourceExpertID == sourceExpertID {
			return row, nil
		}
	}
	return nil, expertmarket.ErrNotFound
}

func (repository *marketplaceRepository) ListApplications(
	context.Context, expertmarket.ApplicationListFilter,
) ([]expertmarket.Application, int64, error) {
	return nil, 0, nil
}

func (repository *marketplaceRepository) GetReleaseByID(
	_ context.Context, id int64,
) (*expertmarket.Release, error) {
	row := repository.releases[id]
	if row == nil {
		return nil, expertmarket.ErrNotFound
	}
	return row, nil
}

func (repository *marketplaceRepository) ListReleases(
	_ context.Context, filter expertmarket.ReleaseListFilter,
) ([]expertmarket.Release, int64, error) {
	rows := make([]expertmarket.Release, 0)
	for _, row := range repository.releases {
		if filter.PublisherOrganizationID == nil ||
			row.PublisherOrganizationID == *filter.PublisherOrganizationID {
			rows = append(rows, *row)
		}
	}
	return rows, int64(len(rows)), nil
}

func (repository *marketplaceRepository) UpdateReleaseLifecycle(
	_ context.Context, id int64, update expertmarket.LifecycleUpdate,
) error {
	repository.releases[id].Status = update.Status
	return nil
}
func (*marketplaceRepository) UpdateReleaseLifecycleAndLatest(
	context.Context, int64, int64, expertmarket.LifecycleUpdate,
) error {
	return nil
}
func (repository *marketplaceRepository) WithdrawReleaseAndRefreshLatest(
	_ context.Context, _ int64, releaseID int64, update expertmarket.LifecycleUpdate,
) error {
	repository.releases[releaseID].Status = update.Status
	return nil
}

type marketplaceWorkerSpecs struct{}

func (marketplaceWorkerSpecs) GetByID(
	context.Context, int64, int64,
) (specdomain.Snapshot, error) {
	return specdomain.Snapshot{}, expertdom.ErrNotFound
}

type marketplaceSkills struct{}

func (marketplaceSkills) ListByIDs(
	context.Context, []int64,
) ([]skilldom.Skill, error) {
	return nil, nil
}
func (marketplaceSkills) ListActivePlatformBySlugs(
	context.Context, []string,
) ([]skilldom.Skill, error) {
	return nil, nil
}

type marketplaceLock struct{}

func (marketplaceLock) WithinMarketApplicationLock(
	_ context.Context, _ int64, run func() error,
) error {
	return run()
}
func (marketplaceLock) WithinMarketInstallationLock(
	_ context.Context, _, _ int64, run func() error,
) error {
	return run()
}
