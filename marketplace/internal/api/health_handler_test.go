package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/stretchr/testify/require"
)

func TestLiveDoesNotDependOnDatabase(t *testing.T) {
	router := NewRouter(Dependencies{
		Ready:      func(context.Context) error { return errors.New("database unavailable") },
		Storefront: testStorefront(),
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"status":"live"}`, response.Body.String())
}

func TestReadyReportsDatabaseFailure(t *testing.T) {
	router := NewRouter(Dependencies{
		Ready:      func(context.Context) error { return errors.New("database unavailable") },
		Storefront: testStorefront(),
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
		Ready:      func(context.Context) error { return nil },
		Storefront: testStorefront(),
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
