package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	expertservice "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	extensionservice "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
)

type PublicMarketHandler struct {
	extension *extensionservice.Service
	expert    *expertservice.Service
}

func NewPublicMarketHandler(
	extension *extensionservice.Service,
	expert *expertservice.Service,
) *PublicMarketHandler {
	return &PublicMarketHandler{extension: extension, expert: expert}
}

func (h *PublicMarketHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/applications", h.ListApplications)
	rg.GET("/skills", h.ListSkills)
}

func (h *PublicMarketHandler) ListApplications(c *gin.Context) {
	if h.expert == nil {
		apierr.InternalError(c, "Expert application market is unavailable")
		return
	}
	items := h.expert.ListMarketApplications()
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *PublicMarketHandler) ListSkills(c *gin.Context) {
	if h.extension == nil {
		apierr.InternalError(c, "Skill market is unavailable")
		return
	}
	items, err := h.extension.ListMarketSkills(c.Request.Context(), 0, c.Query("q"), c.Query("category"))
	if err != nil {
		apierr.InternalError(c, "Failed to load skill market")
		return
	}
	out := make([]publicMarketSkill, 0, len(items))
	for _, item := range items {
		if item.OrganizationID != nil {
			continue
		}
		out = append(out, publicMarketSkillFromDomain(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": out, "total": len(out)})
}

type publicMarketSkill struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	License     string `json:"license"`
	Category    string `json:"category"`
	Version     int    `json:"version"`
	PackageSize int64  `json:"package_size"`
	UpdatedAt   string `json:"updated_at"`
}

func publicMarketSkillFromDomain(item skilldom.Skill) publicMarketSkill {
	return publicMarketSkill{
		ID:          item.ID,
		Slug:        item.Slug,
		DisplayName: item.DisplayName,
		Description: item.Description,
		License:     item.License,
		Category:    item.Category,
		Version:     item.Version,
		PackageSize: item.PackageSize,
		UpdatedAt:   item.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
