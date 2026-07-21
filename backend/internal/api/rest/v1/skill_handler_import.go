package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	skillSvc "github.com/l8ai-cn/agentcloud/backend/internal/service/skill"
	"github.com/l8ai-cn/agentcloud/backend/pkg/apierr"
)

// ImportSkills imports one or more skills from an external git repo (single
// skill repo or collection — auto-detected) into the unified catalog.
func (h *SkillHandler) ImportSkills(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var req importSkillsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}

	rows, err := h.service.ImportFromGit(c.Request.Context(), &skillSvc.ImportFromGitRequest{
		OrganizationID: tenant.OrganizationID,
		UserID:         tenant.UserID,
		URL:            req.URL,
		Branch:         req.Branch,
		Subdir:         req.Subdir,
		AgentFilter:    req.AgentFilter,
		AuthType:       req.AuthType,
		AuthCredential: req.AuthCredential,
	})
	if errors.Is(err, skillSvc.ErrInvalidTags) {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, err.Error())
		return
	}
	if err != nil && len(rows) == 0 {
		if errors.Is(err, skillSvc.ErrImportURLRequired) {
			apierr.BadRequest(c, apierr.VALIDATION_FAILED, err.Error())
			return
		}
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, "Import failed: "+err.Error())
		return
	}

	resp := gin.H{"skills": rows, "imported": len(rows)}
	if err != nil {
		resp["partial_errors"] = err.Error()
	}
	c.JSON(http.StatusCreated, resp)
}

// SyncSkillUpstream re-syncs an imported skill from its recorded upstream.
func (h *SkillHandler) SyncSkillUpstream(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	row, err := h.service.SyncFromUpstream(c.Request.Context(), tenant.OrganizationID, c.Param("skillSlug"))
	if err != nil {
		switch {
		case errors.Is(err, skilldom.ErrNotFound):
			apierr.ResourceNotFound(c, "Skill not found")
		case errors.Is(err, skillSvc.ErrNotImported),
			errors.Is(err, skillSvc.ErrInvalidTags):
			apierr.BadRequest(c, apierr.VALIDATION_FAILED, err.Error())
		default:
			apierr.BadRequest(c, apierr.VALIDATION_FAILED, "Sync failed: "+err.Error())
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"skill": row})
}
