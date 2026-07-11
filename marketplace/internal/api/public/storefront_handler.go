package public

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/gin-gonic/gin"
)

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
	item, err := h.storefront.GetMarket(c, c.Param("marketSlug"), c.Request.Host)
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
	items, err := h.storefront.ListListings(
		c,
		c.Param("marketSlug"),
		c.Request.Host,
		20,
	)
	if err != nil {
		writeStorefrontError(c, err)
		return
	}
	response := make([]listingSummaryResponse, 0, len(items))
	for _, item := range items {
		response = append(response, mapListingSummary(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": response, "next_cursor": nil})
}

func (h *Handler) getListing(c *gin.Context) {
	item, err := h.storefront.GetListing(
		c,
		c.Param("marketSlug"),
		c.Request.Host,
		c.Param("listingSlug"),
	)
	if err != nil {
		writeStorefrontError(c, err)
		return
	}
	c.JSON(http.StatusOK, mapListingDetail(item))
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
