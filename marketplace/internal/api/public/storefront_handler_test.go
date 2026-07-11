package public

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestListListingsReturnsPublicContract(t *testing.T) {
	publishedAt := time.Date(2026, 7, 11, 16, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	storefront := service.NewStorefrontService(&repositoryStub{
		market: service.MarketView{
			MarketplaceID: 42,
			Slug:          "commerce-market",
			Name:          "跨境电商市场",
			Summary:       "开箱即用的 AI 应用",
			Status:        "published",
			DefaultLocale: "zh-CN",
		},
		items: []service.ListingSummary{{
			ListingID:    108,
			Slug:         "listing-optimizer",
			ResourceType: "application",
			DisplayName:  "商品优化应用",
			Tagline:      "提升商品发布效率",
			Publisher:    service.PublisherView{Slug: "commerce-lab", DisplayName: "Commerce Lab", Verified: true},
			Spaces:       []service.SpaceView{{Slug: "operations", Name: "运营"}},
			PublishedAt:  publishedAt,
		}},
	})
	router := gin.New()
	NewHandler(storefront).RegisterRoutes(router.Group("/api/marketplace/v1"))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/marketplace/v1/markets/commerce-market/listings", nil)
	request.Host = "market.example.com"
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{
	  "items":[{
	    "listing_id":"108",
	    "slug":"listing-optimizer",
	    "resource_type":"application",
	    "display_name":"商品优化应用",
	    "tagline":"提升商品发布效率",
	    "publisher":{"slug":"commerce-lab","display_name":"Commerce Lab","verified":true},
	    "spaces":[{"slug":"operations","name":"运营"}],
	    "published_at":"2026-07-11T08:00:00Z"
	  }],
	  "next_cursor":null
	}`, response.Body.String())
}

func TestMarketNotFoundUsesStableChineseError(t *testing.T) {
	storefront := service.NewStorefrontService(&repositoryStub{marketErr: service.ErrMarketNotFound})
	router := gin.New()
	NewHandler(storefront).RegisterRoutes(router.Group("/api/marketplace/v1"))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/marketplace/v1/markets/missing", nil)
	request.Host = "missing.example.com"
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNotFound, response.Code)
	require.JSONEq(t, `{"error":{"code":"MARKET_NOT_FOUND","message":"找不到这个市场"}}`, response.Body.String())
}

type repositoryStub struct {
	market    service.MarketView
	marketErr error
	items     []service.ListingSummary
}

func (r *repositoryStub) ResolveMarket(context.Context, string, string) (service.MarketView, error) {
	return r.market, r.marketErr
}

func (r *repositoryStub) ListPublishedListings(context.Context, int64, int) ([]service.ListingSummary, error) {
	return r.items, nil
}

func (r *repositoryStub) GetPublishedListing(context.Context, int64, string) (service.ListingDetail, error) {
	return service.ListingDetail{}, service.ErrListingNotFound
}
