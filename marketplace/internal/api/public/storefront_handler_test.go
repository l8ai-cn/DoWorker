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
			ListingID:        108,
			ListingVersionID: 301,
			Slug:             "listing-optimizer",
			ResourceType:     "application",
			DisplayName:      "商品优化应用",
			Tagline:          "提升商品发布效率",
			Publisher:        service.PublisherView{Slug: "commerce-lab", DisplayName: "Commerce Lab", Verified: true},
			Spaces:           []service.SpaceView{{Slug: "operations", Name: "运营"}},
			EstimatedCredits: 20_000_000,
			PublishedAt:      publishedAt,
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
	    "listing_version_id":"301",
	    "slug":"listing-optimizer",
	    "resource_type":"application",
	    "display_name":"商品优化应用",
	    "tagline":"提升商品发布效率",
	    "publisher":{"slug":"commerce-lab","display_name":"Commerce Lab","verified":true},
	    "spaces":[{"slug":"operations","name":"运营"}],
	    "tags":[],
	    "quota":{"mode":"per_install","estimated_credits_micro":"20000000"},
	    "published_at":"2026-07-11T08:00:00Z"
	  }],
	  "next_cursor":null
	}`, response.Body.String())
}

func TestMapListingDetailIncludesRuntimeAgent(t *testing.T) {
	response := mapListingDetail(service.ListingDetail{
		ListingSummary: service.ListingSummary{
			ListingID: 1, ListingVersionID: 2, Slug: "delivery",
		},
		AgentSlug: "codex-cli",
	})

	require.Equal(t, "codex-cli", response.AgentSlug)
}

func TestListListingsFiltersTaxonomyAndReturnsTags(t *testing.T) {
	storefront := service.NewStorefrontService(&repositoryStub{
		market: service.MarketView{
			MarketplaceID: 42,
			Slug:          "commerce-market",
			Status:        "published",
		},
		items: []service.ListingSummary{
			{
				ListingID:        108,
				ListingVersionID: 301,
				Slug:             "software-delivery-expert",
				ResourceType:     "application",
				DisplayName:      "软件交付专家",
				Tagline:          "交付可验证的软件变更",
				Publisher:        service.PublisherView{Slug: "do-worker", DisplayName: "Do Worker", Verified: true},
				Tags: []service.TaxonomyTagView{
					{Slug: "software-delivery", DisplayName: "软件交付", Kind: "scene"},
					{Slug: "enterprise-services", DisplayName: "企业服务", Kind: "industry"},
				},
				PublishedAt: time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC),
			},
			{
				ListingID:        109,
				ListingVersionID: 302,
				Slug:             "course-builder",
				ResourceType:     "application",
				DisplayName:      "课程构建专家",
				Tagline:          "构建教学内容",
				Publisher:        service.PublisherView{Slug: "do-worker", DisplayName: "Do Worker", Verified: true},
				PublishedAt:      time.Date(2026, 7, 12, 7, 0, 0, 0, time.UTC),
			},
		},
	})
	router := gin.New()
	NewHandler(storefront).RegisterRoutes(router.Group("/api/marketplace/v1"))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/marketplace/v1/markets/commerce-market/listings?scene=software-delivery&industry=enterprise-services",
		nil,
	)
	request.Host = "market.example.com"
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{
	  "items":[{
	    "listing_id":"108",
	    "listing_version_id":"301",
	    "slug":"software-delivery-expert",
	    "resource_type":"application",
	    "display_name":"软件交付专家",
	    "tagline":"交付可验证的软件变更",
	    "publisher":{"slug":"do-worker","display_name":"Do Worker","verified":true},
	    "spaces":[],
	    "tags":[
	      {"slug":"software-delivery","display_name":"软件交付","kind":"scene"},
	      {"slug":"enterprise-services","display_name":"企业服务","kind":"industry"}
	    ],
	    "published_at":"2026-07-12T08:00:00Z"
	  }],
	  "next_cursor":null
	}`, response.Body.String())
}

func TestListListingsPassesAllQueryParametersToStorefront(t *testing.T) {
	query := service.ListingQuery{
		Q: "delivery", Scene: "software-delivery", Industry: "enterprise-services",
		Audience: "delivery-engineer", Type: "application", Capability: "code-review", Integration: "github",
		Readiness: "runner-required", Space: "software-delivery", Sort: "relevance",
	}
	cursor, err := service.EncodeListingCursor(service.ListingCursor{
		Sort: "relevance", QueryFingerprint: service.ListingQueryFingerprint("commerce-market", query),
		PublishedAt: time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC), ListingID: 108,
	})
	require.NoError(t, err)
	repository := &repositoryStub{
		market: service.MarketView{MarketplaceID: 42, Slug: "commerce-market", Status: "published"},
		items: []service.ListingSummary{{
			ListingID: 108, ListingVersionID: 301, Slug: "software-delivery-expert",
			ResourceType: "application", DisplayName: "软件交付专家",
		}},
	}
	router := gin.New()
	NewHandler(service.NewStorefrontService(repository)).RegisterRoutes(router.Group("/api/marketplace/v1"))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/marketplace/v1/markets/commerce-market/listings?q=delivery&scene=software-delivery&industry=enterprise-services&audience=delivery-engineer&type=application&capability=code-review&integration=github&readiness=runner-required&space=software-delivery&sort=relevance&cursor="+cursor,
		nil,
	)
	request.Host = "market.example.com"
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, service.ListingQuery{
		Q: "delivery", Scene: "software-delivery", Industry: "enterprise-services",
		Audience: "delivery-engineer", Type: "application", Capability: "code-review", Integration: "github",
		Readiness: "runner-required", Space: "software-delivery", Sort: "relevance",
		Cursor: &service.ListingCursor{
			Sort: "relevance", QueryFingerprint: service.ListingQueryFingerprint("commerce-market", query),
			PublishedAt: time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC), ListingID: 108,
		},
	}, repository.listingQuery)
}

func TestListListingsRejectsCursorFromAnotherQuery(t *testing.T) {
	cursor, err := service.EncodeListingCursor(service.ListingCursor{
		Sort: "relevance", PublishedAt: time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC), ListingID: 108,
	})
	require.NoError(t, err)
	router := gin.New()
	NewHandler(service.NewStorefrontService(&repositoryStub{
		market: service.MarketView{MarketplaceID: 42, Slug: "commerce-market", Status: "published"},
	})).RegisterRoutes(router.Group("/api/marketplace/v1"))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/marketplace/v1/markets/commerce-market/listings?q=delivery&sort=relevance&cursor="+cursor,
		nil,
	)
	request.Host = "market.example.com"
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
	require.JSONEq(t, `{
	  "error":{"code":"INVALID_LISTING_QUERY","message":"无效的列表查询参数"}
	}`, response.Body.String())
}

func TestListListingsRejectsCursorFromAnotherMarket(t *testing.T) {
	query := service.ListingQuery{Sort: "latest"}
	cursor, err := service.EncodeListingCursor(service.ListingCursor{
		Sort: "latest", QueryFingerprint: service.ListingQueryFingerprint("commerce-market", query),
		PublishedAt: time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC), ListingID: 108,
	})
	require.NoError(t, err)
	router := gin.New()
	NewHandler(service.NewStorefrontService(&repositoryStub{
		market: service.MarketView{MarketplaceID: 42, Status: "published"},
	})).RegisterRoutes(router.Group("/api/marketplace/v1"))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/marketplace/v1/markets/campus-market/listings?sort=latest&cursor="+cursor,
		nil,
	)
	request.Host = "market.example.com"
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code)
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

func TestForwardedHostSelectsMarketplaceDomain(t *testing.T) {
	repository := &repositoryStub{
		market: service.MarketView{
			MarketplaceID: 42,
			Slug:          "commerce-market",
			Status:        "published",
		},
	}
	router := gin.New()
	NewHandler(service.NewStorefrontService(repository)).
		RegisterRoutes(router.Group("/api/marketplace/v1"))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/marketplace/v1/markets/commerce-market", nil)
	request.Host = "marketplace:8080"
	request.Header.Set("X-Forwarded-Host", "market.example.com, edge.example.com")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "market.example.com", repository.resolvedHost)
}

type repositoryStub struct {
	market       service.MarketView
	marketErr    error
	items        []service.ListingSummary
	resolvedHost string
	listingQuery service.ListingQuery
}

func (r *repositoryStub) ResolveMarket(_ context.Context, _, host string) (service.MarketView, error) {
	r.resolvedHost = host
	return r.market, r.marketErr
}

func (r *repositoryStub) ListPublishedListings(
	ctx context.Context,
	_ int64,
	_ int,
) ([]service.ListingSummary, error) {
	query := service.ListingQueryFromContext(ctx)
	r.listingQuery = query
	if query.Scene == "software-delivery" && query.Industry == "enterprise-services" {
		return r.items[:1], nil
	}
	return r.items, nil
}

func (r *repositoryStub) GetPublishedListing(context.Context, int64, string) (service.ListingDetail, error) {
	return service.ListingDetail{}, service.ErrListingNotFound
}
