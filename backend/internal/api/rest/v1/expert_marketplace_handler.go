package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	expertsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/expert"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
)

func (h *ExpertHandler) SubmitMarketApplication(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var request submitMarketApplicationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	if err := request.validate(); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	source, err := h.service.GetBySlug(
		c.Request.Context(),
		tenant.OrganizationID,
		c.Param("expertSlug"),
	)
	if err != nil {
		h.marketplaceError(c, err, "Failed to find source expert")
		return
	}
	submission, err := h.service.SubmitMarketApplication(
		c.Request.Context(),
		expertsvc.SubmitMarketApplicationRequest{
			OrganizationID: tenant.OrganizationID,
			UserID:         tenant.UserID,
			SourceExpertID: source.ID,
			Slug:           request.Slug,
			Summary:        request.Summary,
			Description:    request.Description,
			Category:       request.Category,
			Icon:           request.Icon,
			Tags:           request.Tags,
			Outcomes:       request.Outcomes,
		},
	)
	if err != nil {
		h.marketplaceError(c, err, "Failed to submit expert application")
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"application": submission.Application,
		"release":     submission.Release,
	})
}

func (h *ExpertHandler) ListMarketSubmissions(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	limit, offset, err := marketplacePagination(c)
	if err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	releases, total, err := h.service.ListPublisherMarketReleases(
		c.Request.Context(),
		tenant.OrganizationID,
		limit,
		offset,
	)
	if err != nil {
		h.marketplaceError(c, err, "Failed to list market submissions")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"releases": releases,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *ExpertHandler) WithdrawMarketRelease(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	releaseID, err := positivePathID(c, "releaseID")
	if err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	release, err := h.service.WithdrawMarketRelease(
		c.Request.Context(),
		expertsvc.WithdrawMarketReleaseRequest{
			PublisherOrganizationID: tenant.OrganizationID,
			ReleaseID:               releaseID,
		},
	)
	if err != nil {
		h.marketplaceError(c, err, "Failed to withdraw market release")
		return
	}
	c.JSON(http.StatusOK, gin.H{"release": release})
}

func (h *ExpertHandler) UpgradeMarketApplication(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	expert, err := h.service.GetBySlug(
		c.Request.Context(),
		tenant.OrganizationID,
		c.Param("expertSlug"),
	)
	if err != nil {
		h.marketplaceError(c, err, "Failed to find installed expert")
		return
	}
	upgraded, changed, err := h.service.UpgradeMarketApplication(
		c.Request.Context(),
		expertsvc.UpgradeMarketApplicationRequest{
			OrganizationID:   tenant.OrganizationID,
			OrganizationSlug: tenant.OrganizationSlug,
			UserID:           tenant.UserID,
			ExpertID:         expert.ID,
		},
	)
	if err != nil {
		h.marketplaceError(c, err, "Failed to upgrade market expert")
		return
	}
	c.JSON(http.StatusOK, gin.H{"expert": upgraded, "upgraded": changed})
}

func (h *ExpertHandler) GetMarketUpgradeAvailability(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	expert, err := h.service.GetBySlug(
		c.Request.Context(),
		tenant.OrganizationID,
		c.Param("expertSlug"),
	)
	if err != nil || expert.SourceMarketApplicationID == nil {
		if err == nil {
			apierr.ResourceNotFound(c, "Marketplace installation not found")
			return
		}
		h.marketplaceError(c, err, "Failed to find installed expert")
		return
	}
	available, err := h.service.MarketUpgradeAvailable(
		c.Request.Context(),
		tenant.OrganizationID,
		*expert.SourceMarketApplicationID,
	)
	if err != nil {
		h.marketplaceError(c, err, "Failed to check market upgrade")
		return
	}
	c.JSON(http.StatusOK, gin.H{"upgrade_available": available})
}
