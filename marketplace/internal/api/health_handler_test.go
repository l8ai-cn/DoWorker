package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authpkg "github.com/anthropics/agentsmesh/backend/pkg/auth"
	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/stretchr/testify/require"
)

func TestLiveDoesNotDependOnDatabase(t *testing.T) {
	router := NewRouter(Dependencies{
		Ready:         func(context.Context) error { return errors.New("database unavailable") },
		Storefront:    testStorefront(),
		Identity:      healthTokenVerifier{},
		Installations: healthInstallations{},
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"status":"live"}`, response.Body.String())
}

func TestReadyReportsDatabaseFailure(t *testing.T) {
	router := NewRouter(Dependencies{
		Ready:         func(context.Context) error { return errors.New("database unavailable") },
		Storefront:    testStorefront(),
		Identity:      healthTokenVerifier{},
		Installations: healthInstallations{},
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
	require.JSONEq(t, `{
		"error": {
			"code": "SERVICE_NOT_READY",
			"message": "市场服务尚未就绪"
		}
	}`, response.Body.String())
}

func TestReadySucceedsAfterDatabaseProbe(t *testing.T) {
	router := NewRouter(Dependencies{
		Ready:         func(context.Context) error { return nil },
		Storefront:    testStorefront(),
		Identity:      healthTokenVerifier{},
		Installations: healthInstallations{},
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"status":"ready"}`, response.Body.String())
}

func TestRouterRequiresStorefront(t *testing.T) {
	require.Panics(t, func() {
		NewRouter(Dependencies{Ready: func(context.Context) error { return nil }})
	})
}

func testStorefront() *service.StorefrontService {
	return service.NewStorefrontService(healthRepository{})
}

type healthRepository struct{}

type healthTokenVerifier struct{}
type healthInstallations struct{}

func (healthTokenVerifier) Verify(context.Context, string) (*authpkg.Claims, error) {
	return nil, errors.New("not used")
}

func (healthInstallations) CreatePlan(
	context.Context,
	service.CreateInstallationPlanCommand,
) (service.InstallationPlanResult, error) {
	return service.InstallationPlanResult{}, errors.New("not used")
}

func (healthInstallations) Apply(
	context.Context,
	service.ApplyInstallationCommand,
) (service.ApplyResult, error) {
	return service.ApplyResult{}, errors.New("not used")
}

func (healthInstallations) GetOperation(
	context.Context,
	string,
	int64,
) (service.ApplyResult, error) {
	return service.ApplyResult{}, errors.New("not used")
}

func (healthRepository) ResolveMarket(
	context.Context,
	string,
	string,
) (service.MarketView, error) {
	return service.MarketView{}, service.ErrMarketNotFound
}

func (healthRepository) ListPublishedListings(
	context.Context,
	int64,
	int,
) ([]service.ListingSummary, error) {
	return nil, nil
}

func (healthRepository) GetPublishedListing(
	context.Context,
	int64,
	string,
) (service.ListingDetail, error) {
	return service.ListingDetail{}, service.ErrListingNotFound
}
