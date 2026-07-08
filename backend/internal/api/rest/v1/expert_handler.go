package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	expertSvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
)

type ExpertHandler struct {
	service *expertSvc.Service
}

func NewExpertHandler(service *expertSvc.Service) *ExpertHandler {
	return &ExpertHandler{service: service}
}

func (h *ExpertHandler) ListExperts(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, total, err := h.service.List(c.Request.Context(), tenant.OrganizationID, limit, offset)
	if err != nil {
		apierr.InternalError(c, "Failed to list experts")
		return
	}
	c.JSON(http.StatusOK, gin.H{"experts": items, "total": total, "limit": limit, "offset": offset})
}

func (h *ExpertHandler) GetExpert(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	row, err := h.service.GetBySlug(c.Request.Context(), tenant.OrganizationID, c.Param("expertSlug"))
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"expert": row})
}

func (h *ExpertHandler) CreateExpert(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var req createExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	row, err := h.service.Create(c.Request.Context(), &expertSvc.CreateExpertRequest{
		OrganizationID: tenant.OrganizationID, UserID: tenant.UserID,
		Name: req.Name, Slug: req.Slug, Description: req.Description,
		AgentSlug: req.AgentSlug, RunnerID: req.RunnerID, RepositoryID: req.RepositoryID,
		BranchName: req.BranchName, Prompt: req.Prompt, InteractionMode: req.InteractionMode,
		Perpetual: req.Perpetual, UsedEnvBundles: req.UsedEnvBundles, SkillSlugs: req.SkillSlugs,
		KnowledgeMounts: req.KnowledgeMounts, ConfigOverrides: req.ConfigOverrides,
		AgentfileLayer: req.AgentfileLayer,
	})
	if err != nil {
		h.validationOrInternal(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"expert": row})
}

func (h *ExpertHandler) UpdateExpert(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	row, err := h.service.GetBySlug(c.Request.Context(), tenant.OrganizationID, c.Param("expertSlug"))
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	var req updateExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	updated, err := h.service.Update(c.Request.Context(), &expertSvc.UpdateExpertRequest{
		OrganizationID: tenant.OrganizationID, ExpertID: row.ID,
		Name: req.Name, Description: req.Description, AgentSlug: req.AgentSlug,
		RunnerID: req.RunnerID, RepositoryID: req.RepositoryID, BranchName: req.BranchName,
		Prompt: req.Prompt, InteractionMode: req.InteractionMode, Perpetual: req.Perpetual,
		UsedEnvBundles: req.UsedEnvBundles, SkillSlugs: req.SkillSlugs,
		KnowledgeMounts: req.KnowledgeMounts, ConfigOverrides: req.ConfigOverrides,
		AgentfileLayer: req.AgentfileLayer,
	})
	if err != nil {
		h.validationOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"expert": updated})
}

func (h *ExpertHandler) DeleteExpert(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	row, err := h.service.GetBySlug(c.Request.Context(), tenant.OrganizationID, c.Param("expertSlug"))
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), tenant.OrganizationID, row.ID); err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Expert deleted"})
}

func (h *ExpertHandler) RunExpert(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var req runExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil && c.Request.ContentLength > 0 {
		apierr.ValidationError(c, err.Error())
		return
	}
	result, err := h.service.Run(c.Request.Context(), &expertSvc.RunExpertRequest{
		OrganizationID: tenant.OrganizationID, UserID: tenant.UserID,
		ExpertSlug: c.Param("expertSlug"), Alias: req.Alias, PromptOverride: req.PromptOverride,
		RunnerID: req.RunnerID, Cols: req.Cols, Rows: req.Rows,
	})
	if err != nil {
		mapOrchestratorErrorToHTTP(c, err)
		return
	}
	if result.Warning != "" {
		c.JSON(http.StatusCreated, gin.H{"pod": result.Pod, "warning": result.Warning})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"pod": result.Pod})
}

func (h *ExpertHandler) PublishFromPod(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var req publishExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	row, err := h.service.PublishFromPod(c.Request.Context(), &expertSvc.PublishFromPodRequest{
		OrganizationID: tenant.OrganizationID, UserID: tenant.UserID,
		PodKey: c.Param("pod_key"), Name: req.Name, Slug: req.Slug, Description: req.Description,
		AgentfileLayer: req.AgentfileLayer, UsedEnvBundles: req.UsedEnvBundles,
		SkillSlugs: req.SkillSlugs, KnowledgeMounts: req.KnowledgeMounts,
	})
	if err != nil {
		if errors.Is(err, expertSvc.ErrPodAccessDenied) {
			apierr.Forbidden(c, apierr.SOURCE_POD_ACCESS_DENIED, "Pod belongs to a different organization")
			return
		}
		h.validationOrInternal(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"expert": row})
}

func (h *ExpertHandler) notFoundOrInternal(c *gin.Context, err error) {
	if errors.Is(err, expertdom.ErrNotFound) {
		apierr.ResourceNotFound(c, "Expert not found")
		return
	}
	apierr.InternalError(c, "Expert request failed")
}

func (h *ExpertHandler) validationOrInternal(c *gin.Context, err error) {
	switch {
	case errors.Is(err, expertSvc.ErrExpertNameRequired),
		errors.Is(err, expertSvc.ErrExpertAgentRequired):
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, err.Error())
	default:
		apierr.InternalError(c, "Expert request failed")
	}
}
