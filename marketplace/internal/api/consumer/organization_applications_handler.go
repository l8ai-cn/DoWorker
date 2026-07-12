package consumer

import (
	"context"
	"net/http"
	"strconv"

	actorapi "github.com/anthropics/agentsmesh/marketplace/internal/api/actor"
	"github.com/anthropics/agentsmesh/marketplace/internal/service"
	"github.com/gin-gonic/gin"
)

type OrganizationApplicationsReader interface {
	ListOrganizationApplications(
		context.Context,
		int64,
		int64,
	) ([]service.OrganizationApplication, error)
}

type OrganizationApplicationsHandler struct {
	applications OrganizationApplicationsReader
}

func NewOrganizationApplicationsHandler(
	applications OrganizationApplicationsReader,
) *OrganizationApplicationsHandler {
	return &OrganizationApplicationsHandler{applications: applications}
}

func (h *OrganizationApplicationsHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/organizations/:organizationID/applications", h.list)
}

func (h *OrganizationApplicationsHandler) list(c *gin.Context) {
	current, ok := actorapi.FromContext(c)
	if !ok {
		writeInstallationError(c, service.ErrInvalidInstallationRequest)
		return
	}
	organizationID, err := strconv.ParseInt(c.Param("organizationID"), 10, 64)
	if err != nil || organizationID <= 0 {
		writeInstallationError(c, service.ErrInvalidInstallationRequest)
		return
	}
	applications, err := h.applications.ListOrganizationApplications(
		c,
		organizationID,
		current.UserID,
	)
	if err != nil {
		writeInstallationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"applications": applications})
}
