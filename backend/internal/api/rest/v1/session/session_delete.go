package sessionapi

import (
	"errors"
	"net/http"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/grant"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	runnerservice "github.com/l8ai-cn/agentcloud/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

const projectLabelKey = "omni_project"

func (d *Deps) handleDeleteSession(c *gin.Context) {
	row, pod, ok := d.authorizeSession(c, c.Param("id"))
	if !ok {
		return
	}
	deleteBranch := c.Query("delete_branch") == "true"
	if pod != nil {
		if d.PodCoordinator != nil {
			var err error
			if deleteBranch {
				err = d.PodCoordinator.TerminatePodDeleteBranch(c.Request.Context(), pod.PodKey)
			} else {
				err = d.PodCoordinator.TerminatePod(c.Request.Context(), pod.PodKey)
			}
			if err != nil && !errors.Is(err, runnerservice.ErrPodAlreadyTerminated) {
				_ = err
			}
		}
		if d.Grants != nil {
			_ = d.Grants.CleanupByResource(c.Request.Context(), grant.TypePod, pod.PodKey)
		}
	}
	if err := d.Sessions.SoftDelete(c.Request.Context(), row.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if d.Stream != nil {
		d.Stream.PublishSessionStatus(row.ID, "idle")
	} else if d.Hub != nil {
		d.Hub.Publish(row.ID, formatSSE("session.status", map[string]any{
			"conversation_id": row.ID, "status": "idle",
		}))
	}
	if d.Hub != nil {
		d.Hub.RemoveSession(row.ID)
	}
	if d.Elicitations != nil {
		d.Elicitations.RemoveSession(row.ID)
	}
	if d.Updates != nil {
		d.Updates.NotifyChanged(row.ID)
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) handleListProjects(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Sessions == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	names, err := d.Sessions.ListProjects(c.Request.Context(), tenant.OrganizationID, tenant.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list projects failed"})
		return
	}
	if names == nil {
		names = []string{}
	}
	c.JSON(http.StatusOK, names)
}

func sessionLabels(project *string) map[string]string {
	labels := map[string]string{}
	if project != nil && *project != "" {
		labels[projectLabelKey] = *project
	}
	return labels
}
