package omnigent

import (
	"net/http"
	"strconv"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/gin-gonic/gin"
)

func (d *Deps) handleListSessions(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Sessions == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	limit := 100
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	rows, err := d.Sessions.ListForUser(c.Request.Context(), tenant.OrganizationID, tenant.UserID, sessionsvc.ListOptions{
		Limit:           limit,
		Project:         c.Query("project"),
		IncludeArchived: c.Query("include_archived") == "true",
		PrincipalEmail:  d.viewerEmail(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	online := d.runnerOnlineMap(tenant.OrganizationID, tenant.UserID)
	items := make([]conversationListItem, 0, len(rows))
	for i := range rows {
		pod := d.loadPod(c, rows[i].PodKey)
		item := d.listItemFrom(&rows[i], pod, online)
		d.enrichOwnership(c, &rows[i], &item)
		items = append(items, item)
	}
	c.JSON(http.StatusOK, listPageFrom(items))
}

func (d *Deps) handleGetSession(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	tenant := middleware.GetTenant(c)
	online := map[string]bool{}
	if tenant != nil {
		online = d.runnerOnlineMap(tenant.OrganizationID, tenant.UserID)
	}
	item := d.listItemFrom(row, pod, online)
	d.enrichOwnership(c, row, &item)
	c.JSON(http.StatusOK, mergeSessionGet(item, d.sessionWire(row, pod, row.RunnerNodeID)))
}

func (d *Deps) handlePatchSession(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	var body struct {
		RunnerID *string           `json:"runner_id"`
		Title    *string           `json:"title"`
		Archived *bool             `json:"archived"`
		Labels   map[string]string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Title != nil {
		if err := d.Sessions.UpdateTitle(c.Request.Context(), row.ID, body.Title); err != nil {
			if err == sessionsvc.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
		row.Title = body.Title
	}
	if body.Archived != nil {
		if err := d.Sessions.UpdateArchived(c.Request.Context(), row.ID, *body.Archived); err != nil {
			if err == sessionsvc.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
		row.Archived = *body.Archived
	}
	if body.Labels != nil {
		if v, ok := body.Labels[projectLabelKey]; ok {
			var project *string
			if v != "" {
				project = &v
			}
			if err := d.Sessions.UpdateProject(c.Request.Context(), row.ID, project); err != nil {
				if err == sessionsvc.ErrNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
				return
			}
			row.Project = project
		}
	}
	if body.RunnerID != nil && d.Runner != nil {
		r, err := d.Runner.GetByNodeIDAndOrgID(c.Request.Context(), *body.RunnerID, row.OrganizationID)
		if err != nil || r == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "runner not found", "code": "runner_not_found"})
			return
		}
		if err := d.Sessions.UpdateRunner(c.Request.Context(), row.ID, *body.RunnerID); err != nil {
			if err == sessionsvc.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
			return
		}
		row.RunnerNodeID = body.RunnerID
	}
	tenant := middleware.GetTenant(c)
	online := map[string]bool{}
	if tenant != nil {
		online = d.runnerOnlineMap(tenant.OrganizationID, tenant.UserID)
	}
	item := d.listItemFrom(row, pod, online)
	d.enrichOwnership(c, row, &item)
	if tenant != nil {
		d.enrichReadState(tenant.UserID, row.ID, &item)
	}
	c.JSON(http.StatusOK, mergeSessionGet(item, d.sessionWire(row, pod, row.RunnerNodeID)))
}

func (d *Deps) loadPod(c *gin.Context, podKey string) *podDomain.Pod {
	if d.Pod == nil || podKey == "" {
		return nil
	}
	pod, err := d.Pod.GetPod(c.Request.Context(), podKey)
	if err != nil {
		return nil
	}
	return pod
}
