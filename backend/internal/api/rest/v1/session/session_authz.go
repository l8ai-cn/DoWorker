package sessionapi

import (
	"net/http"

	podDomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	domain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	sessionsvc "github.com/anthropics/agentsmesh/backend/internal/service/agentsession"
	"github.com/gin-gonic/gin"
)

const (
	levelRead   = 1
	levelEdit   = 2
	levelManage = 3
	levelOwner  = 4
)

func (d *Deps) viewerEmail(c *gin.Context) string {
	if v, ok := c.Get("email"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if d.User == nil {
		return ""
	}
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		return ""
	}
	u, err := d.User.GetByID(c.Request.Context(), tenant.UserID)
	if err != nil || u == nil {
		return ""
	}
	return u.Email
}

func (d *Deps) sessionAccessLevel(c *gin.Context, row *domain.Session) int {
	tenant := middleware.GetTenant(c)
	if tenant == nil || row == nil {
		return 0
	}
	if row.OrganizationID != tenant.OrganizationID {
		return 0
	}
	if row.UserID == tenant.UserID {
		return levelOwner
	}
	if d.SessionPermissions == nil {
		return 0
	}
	email := d.viewerEmail(c)
	principals := []string{"__public__"}
	if email != "" {
		principals = append(principals, email)
	}
	level, ok := d.SessionPermissions.EffectiveLevel(c.Request.Context(), row.ID, principals...)
	if !ok {
		return 0
	}
	return level
}

func (d *Deps) authorizeSessionByPodKey(c *gin.Context, podKey string) (*domain.Session, *podDomain.Pod, bool) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Sessions == nil || podKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, nil, false
	}
	row, err := d.Sessions.GetByPodKey(c.Request.Context(), podKey)
	if err != nil {
		if err == sessionsvc.ErrNotFound {
			c.Status(http.StatusNoContent)
			return nil, nil, false
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return nil, nil, false
	}
	if row == nil || d.sessionAccessLevel(c, row) < levelRead {
		c.Status(http.StatusNoContent)
		return nil, nil, false
	}
	return row, d.loadPod(c, row.PodKey), true
}

func (d *Deps) authorizeSession(c *gin.Context, id string) (*domain.Session, *podDomain.Pod, bool) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.Sessions == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, nil, false
	}
	row, err := d.Sessions.GetActive(c.Request.Context(), id)
	if err == sessionsvc.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found", "code": "session_not_found"})
		return nil, nil, false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return nil, nil, false
	}
	if d.sessionAccessLevel(c, row) < levelRead {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found", "code": "session_not_found"})
		return nil, nil, false
	}
	return row, d.loadPod(c, row.PodKey), true
}

func (d *Deps) requireSessionLevel(c *gin.Context, row *domain.Session, min int) bool {
	if d.sessionAccessLevel(c, row) >= min {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	return false
}
