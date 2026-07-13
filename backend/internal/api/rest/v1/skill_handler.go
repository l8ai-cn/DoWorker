package v1

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	skillSvc "github.com/anthropics/agentsmesh/backend/internal/service/skill"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
)

// SkillHandler serves the git-backed, author-in-platform skill routes. This is
// additive to the external-import/marketplace skill flow (which is served over
// Connect-RPC and is untouched).
type SkillHandler struct {
	service skillHandlerService
}

func NewSkillHandler(service skillHandlerService) *SkillHandler {
	return &SkillHandler{service: service}
}

func (h *SkillHandler) ListSkills(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if c.Query("all") == "true" {
		items, err := h.service.ListAll(c.Request.Context(), tenant.OrganizationID)
		if err != nil {
			apierr.InternalError(c, "Failed to list skills")
			return
		}
		c.JSON(http.StatusOK, gin.H{"skills": items, "total": len(items)})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.service.List(c.Request.Context(), tenant.OrganizationID, limit, offset)
	if err != nil {
		apierr.InternalError(c, "Failed to list skills")
		return
	}
	c.JSON(http.StatusOK, gin.H{"skills": items, "total": total, "limit": limit, "offset": offset})
}

func (h *SkillHandler) GetSkill(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	row, err := h.service.Get(c.Request.Context(), tenant.OrganizationID, c.Param("skillSlug"))
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"skill": row})
}

func (h *SkillHandler) CreateSkill(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var req createSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	row, err := h.service.Create(c.Request.Context(), &skillSvc.CreateSkillRequest{
		OrganizationID: tenant.OrganizationID, UserID: tenant.UserID,
		Slug: req.Slug, Name: req.Name, Description: req.Description,
		License: req.License, Instructions: req.Instructions, Tags: req.Tags,
	})
	if err != nil {
		h.validationOrInternal(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"skill": row})
}

func (h *SkillHandler) UpdateSkill(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	row, err := h.service.Get(c.Request.Context(), tenant.OrganizationID, c.Param("skillSlug"))
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	var req updateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	updated, err := h.service.Update(c.Request.Context(), &skillSvc.UpdateSkillRequest{
		OrganizationID: tenant.OrganizationID, SkillID: row.ID,
		Name: req.Name, Description: req.Description,
		License: req.License, Instructions: req.Instructions, Tags: req.Tags,
	})
	if err != nil {
		h.validationOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"skill": updated})
}

func (h *SkillHandler) DeleteSkill(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	row, err := h.service.Get(c.Request.Context(), tenant.OrganizationID, c.Param("skillSlug"))
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), tenant.OrganizationID, row.ID); err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Skill deleted"})
}

// GetSkillFile returns a single file from the authored skill's backing repo.
// The *path param is sanitized (shared sanitizeRepoPath) to reject traversal.
func (h *SkillHandler) GetSkillFile(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	rel, err := sanitizeRepoPath(c.Param("path"))
	if err != nil {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, err.Error())
		return
	}
	data, entry, err := h.service.ReadSkillFile(c.Request.Context(), tenant.OrganizationID, c.Param("skillSlug"), rel)
	if err != nil {
		h.gitReadError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"path":    entry.Path,
		"name":    entry.Name,
		"size":    entry.Size,
		"sha":     entry.SHA,
		"content": base64.StdEncoding.EncodeToString(data),
	})
}

// GetSkillTree returns the file tree of the authored skill's backing repo.
func (h *SkillHandler) GetSkillTree(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	entries, err := h.service.ListSkillTree(c.Request.Context(), tenant.OrganizationID, c.Param("skillSlug"))
	if err != nil {
		h.gitReadError(c, err)
		return
	}
	out := make([]gin.H, 0, len(entries))
	for _, e := range entries {
		out = append(out, gin.H{"path": e.Path, "name": e.Name, "type": e.Type, "size": e.Size, "sha": e.SHA})
	}
	c.JSON(http.StatusOK, gin.H{"entries": out})
}

func (h *SkillHandler) notFoundOrInternal(c *gin.Context, err error) {
	if errors.Is(err, skilldom.ErrNotFound) {
		apierr.ResourceNotFound(c, "Skill not found")
		return
	}
	apierr.InternalError(c, "Skill request failed")
}

func (h *SkillHandler) validationOrInternal(c *gin.Context, err error) {
	switch {
	case errors.Is(err, skillSvc.ErrNameRequired),
		errors.Is(err, skillSvc.ErrInstructionsRequired),
		errors.Is(err, skillSvc.ErrInvalidTags):
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, err.Error())
	case errors.Is(err, skilldom.ErrNotFound):
		apierr.ResourceNotFound(c, "Skill not found")
	default:
		apierr.InternalError(c, "Skill request failed")
	}
}

func (h *SkillHandler) gitReadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, skilldom.ErrNotFound),
		errors.Is(err, gitops.ErrNotFound):
		apierr.ResourceNotFound(c, "Not found")
	default:
		apierr.InternalError(c, "Skill request failed")
	}
}
