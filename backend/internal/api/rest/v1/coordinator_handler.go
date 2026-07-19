package v1

import (
	"context"
	"net/http"
	"strconv"

	"github.com/anthropics/agentsmesh/backend/internal/domain/gitprovider"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	coordinatorService "github.com/anthropics/agentsmesh/backend/internal/service/coordinator"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

// RepoQueryForCoordinator validates that a repository belongs to the org before
// a project is bound to it.
type RepoQueryForCoordinator interface {
	GetByID(ctx context.Context, id int64) (*gitprovider.Repository, error)
}

type CoordinatorHandler struct {
	service *coordinatorService.Service
	repos   RepoQueryForCoordinator
}

func NewCoordinatorHandler(service *coordinatorService.Service, repos RepoQueryForCoordinator) *CoordinatorHandler {
	return &CoordinatorHandler{service: service, repos: repos}
}

// ListProjects GET /api/v1/orgs/:slug/coordinator/projects
func (h *CoordinatorHandler) ListProjects(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	projects, err := h.service.ListProjects(c.Request.Context(), tenant.OrganizationID)
	if err != nil {
		apierr.InternalError(c, "Failed to list coordinator projects")
		return
	}
	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

// CreateProject POST /api/v1/orgs/:slug/coordinator/projects
func (h *CoordinatorHandler) CreateProject(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	var req createCoordinatorProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	repo, err := h.repos.GetByID(c.Request.Context(), req.RepositoryID)
	if err != nil || repo.OrganizationID != tenant.OrganizationID {
		apierr.ResourceNotFound(c, "Repository not found")
		return
	}

	project, err := h.service.CreateProject(c.Request.Context(), &coordinatorService.CreateProjectRequest{
		OrganizationID:       tenant.OrganizationID,
		RepositoryID:         req.RepositoryID,
		Name:                 req.Name,
		PlatformType:         req.PlatformType,
		SourceType:           req.SourceType,
		LabelFilter:          req.LabelFilter,
		ClaimPolicy:          req.ClaimPolicy,
		WorkerSpecSnapshotID: req.WorkerSpecSnapshotID,
		ScanIntervalSeconds:  req.ScanIntervalSeconds,
		MaxConcurrent:        req.MaxConcurrent,
		CreatedByID:          tenant.UserID,
	})
	if err != nil {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"project": project})
}

// GetProject GET /api/v1/orgs/:slug/coordinator/projects/:id
func (h *CoordinatorHandler) GetProject(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	id := h.parseID(c)
	if id == 0 {
		return
	}
	project, err := h.service.GetProject(c.Request.Context(), tenant.OrganizationID, id)
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"project": project})
}

// UpdateProject PATCH /api/v1/orgs/:slug/coordinator/projects/:id
func (h *CoordinatorHandler) UpdateProject(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	id := h.parseID(c)
	if id == 0 {
		return
	}
	var req updateCoordinatorProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.LabelFilter != nil {
		updates["label_filter"] = *req.LabelFilter
	}
	if req.WorkerSpecSnapshotID != nil {
		updates["worker_spec_snapshot_id"] = *req.WorkerSpecSnapshotID
	}
	if req.ScanIntervalSeconds != nil {
		updates["scan_interval_seconds"] = *req.ScanIntervalSeconds
	}
	if req.MaxConcurrent != nil {
		updates["max_concurrent"] = *req.MaxConcurrent
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if len(updates) == 0 {
		apierr.BadRequest(c, apierr.VALIDATION_FAILED, "No fields to update")
		return
	}
	if err := h.service.UpdateProject(c.Request.Context(), tenant.OrganizationID, id, updates); err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	project, err := h.service.GetProject(c.Request.Context(), tenant.OrganizationID, id)
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"project": project})
}

// DeleteProject DELETE /api/v1/orgs/:slug/coordinator/projects/:id
func (h *CoordinatorHandler) DeleteProject(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	id := h.parseID(c)
	if id == 0 {
		return
	}
	if err := h.service.DeleteProject(c.Request.Context(), tenant.OrganizationID, id); err != nil {
		apierr.InternalError(c, "Failed to delete project")
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// ListExecutions GET /api/v1/orgs/:slug/coordinator/projects/:id/executions
func (h *CoordinatorHandler) ListExecutions(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	id := h.parseID(c)
	if id == 0 {
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	executions, err := h.service.ListExecutions(c.Request.Context(), tenant.OrganizationID, id, limit)
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"executions": executions})
}

// RunNow POST /api/v1/orgs/:slug/coordinator/projects/:id/run
func (h *CoordinatorHandler) RunNow(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	id := h.parseID(c)
	if id == 0 {
		return
	}
	project, err := h.service.GetProject(c.Request.Context(), tenant.OrganizationID, id)
	if err != nil {
		h.notFoundOrInternal(c, err)
		return
	}
	result, err := h.service.RunProject(c.Request.Context(), project)
	if err != nil {
		apierr.InternalError(c, "Coordinator run failed: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}
