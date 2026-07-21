package public

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/l8ai-cn/agentcloud/marketplace/internal/service"
	"github.com/gin-gonic/gin"
)

const listingsPageSize = 20

type Handler struct {
	storefront *service.StorefrontService
}

func NewHandler(storefront *service.StorefrontService) *Handler {
	return &Handler{storefront: storefront}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	markets := router.Group("/markets/:marketSlug")
	markets.GET("", h.getMarket)
	markets.GET("/listings", h.listListings)
	markets.GET("/listings/:listingSlug", h.getListing)
}

func (h *Handler) getMarket(c *gin.Context) {
	item, err := h.storefront.GetMarket(c, c.Param("marketSlug"), requestHost(c))
	if err != nil {
		writeStorefrontError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"marketplace_id": strconv.FormatInt(item.MarketplaceID, 10),
		"slug":           item.Slug,
		"name":           item.Name,
		"summary":        item.Summary,
		"status":         item.Status,
		"default_locale": item.DefaultLocale,
	})
}

func (h *Handler) listListings(c *gin.Context) {
	query, err := parseListingQuery(c, c.Param("marketSlug"))
	if err != nil {
		writeInvalidListingQuery(c)
		return
	}
	queryFingerprint := service.ListingQueryFingerprint(c.Param("marketSlug"), query)
	items, err := h.storefront.ListListings(
		service.WithListingQuery(c, query),
		c.Param("marketSlug"),
		requestHost(c),
		listingsPageSize+1,
	)
	if err != nil {
		writeStorefrontError(c, err)
		return
	}
	nextCursor := any(nil)
	if len(items) > listingsPageSize {
		items = items[:listingsPageSize]
		items[len(items)-1].PageCursor.QueryFingerprint = queryFingerprint
		nextCursor, err = service.EncodeListingCursor(items[len(items)-1].PageCursor)
		if err != nil {
			writeStorefrontError(c, err)
			return
		}
	}
	response := make([]listingSummaryResponse, 0, len(items))
	for _, item := range items {
		response = append(response, mapListingSummary(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": response, "next_cursor": nextCursor})
}

func (h *Handler) getListing(c *gin.Context) {
	item, err := h.storefront.GetListing(
		c,
		c.Param("marketSlug"),
		requestHost(c),
		c.Param("listingSlug"),
	)
	if err != nil {
		writeStorefrontError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapListingDetail(item))
}

func requestHost(c *gin.Context) string {
	if forwarded := c.GetHeader("X-Forwarded-Host"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	return c.Request.Host
}

func parseListingQuery(c *gin.Context, marketSlug string) (service.ListingQuery, error) {
	query := service.ListingQuery{
		Q:           strings.TrimSpace(c.Query("q")),
		Scene:       strings.TrimSpace(c.Query("scene")),
		Industry:    strings.TrimSpace(c.Query("industry")),
		Audience:    strings.TrimSpace(c.Query("audience")),
		Type:        strings.TrimSpace(c.Query("type")),
		Capability:  strings.TrimSpace(c.Query("capability")),
		Integration: strings.TrimSpace(c.Query("integration")),
		Readiness:   strings.TrimSpace(c.Query("readiness")),
		Space:       strings.TrimSpace(c.Query("space")),
		Sort:        strings.TrimSpace(c.DefaultQuery("sort", "featured")),
	}
	if len(query.Q) > 200 || !validListingQueryValue(query) {
		return service.ListingQuery{}, errors.New("invalid listing query")
	}
	if cursor := strings.TrimSpace(c.Query("cursor")); cursor != "" {
		value, err := service.DecodeListingCursor(cursor)
		if err != nil || value.Sort != query.Sort ||
			value.QueryFingerprint != service.ListingQueryFingerprint(marketSlug, query) {
			return service.ListingQuery{}, errors.New("invalid listing cursor")
		}
		query.Cursor = &value
	}
	return query, nil
}

func validListingQueryValue(query service.ListingQuery) bool {
	if query.Sort != "featured" && query.Sort != "latest" && query.Sort != "relevance" {
		return false
	}
	if query.Type != "" && query.Type != "application" && query.Type != "skill" &&
		query.Type != "mcp_connector" && query.Type != "resource" {
		return false
	}
	for _, value := range []string{
		query.Scene, query.Industry, query.Audience, query.Capability, query.Integration, query.Readiness, query.Space,
	} {
		if value != "" && !isIdentifier(value) {
			return false
		}
	}
	return true
}

func isIdentifier(value string) bool {
	if len(value) < 2 || len(value) > 100 || value[0] == '-' || value[len(value)-1] == '-' {
		return false
	}
	for _, char := range value {
		if char != '-' && (char < 'a' || char > 'z') && (char < '0' || char > '9') {
			return false
		}
	}
	return !strings.Contains(value, "--")
}

func writeInvalidListingQuery(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": gin.H{"code": "INVALID_LISTING_QUERY", "message": "无效的列表查询参数"},
	})
}

func writeStorefrontError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "市场服务暂时不可用"
	switch {
	case errors.Is(err, service.ErrMarketNotFound):
		status, code, message = http.StatusNotFound, "MARKET_NOT_FOUND", "找不到这个市场"
	case errors.Is(err, service.ErrMarketSuspended):
		status, code, message = http.StatusServiceUnavailable, "MARKET_SUSPENDED", "市场暂时停止服务"
	case errors.Is(err, service.ErrListingNotFound):
		status, code, message = http.StatusNotFound, "LISTING_NOT_AVAILABLE", "此内容当前不可获取"
	}
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
