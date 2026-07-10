package sessionapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	virtualkeysvc "github.com/anthropics/agentsmesh/backend/internal/service/virtualkey"
	"github.com/gin-gonic/gin"
)

type virtualKeyWire struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	KeyPrefix       string     `json:"key_prefix"`
	ModelResourceID int64      `json:"model_resource_id"`
	TokenBudget     *int64     `json:"token_budget,omitempty"`
	Status          string     `json:"status"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func (d *Deps) handleListVirtualKeys(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.VirtualKeys == nil {
		c.JSON(http.StatusOK, gin.H{"object": "list", "data": []virtualKeyWire{}})
		return
	}
	keys, err := d.VirtualKeys.List(c.Request.Context(), tenant.OrganizationID, tenant.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list virtual keys"})
		return
	}
	out := make([]virtualKeyWire, 0, len(keys))
	for _, k := range keys {
		out = append(out, virtualKeyWire{
			ID: k.ID, Name: k.Name, KeyPrefix: k.KeyPrefix, ModelResourceID: k.ModelResourceID,
			TokenBudget: k.TokenBudget, Status: k.Status, LastUsedAt: k.LastUsedAt, CreatedAt: k.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": out})
}

type createVirtualKeyBody struct {
	Name            string `json:"name"`
	ModelResourceID int64  `json:"model_resource_id"`
	TokenBudget     *int64 `json:"token_budget"`
}

func (d *Deps) handleCreateVirtualKey(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.VirtualKeys == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "virtual key service unavailable"})
		return
	}
	var body createVirtualKeyBody
	if err := c.ShouldBindJSON(&body); err != nil ||
		strings.TrimSpace(body.Name) == "" || body.ModelResourceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and model_resource_id are required"})
		return
	}
	created, err := d.VirtualKeys.Create(c.Request.Context(), virtualkeysvc.CreateInput{
		OrgID:           tenant.OrganizationID,
		UserID:          tenant.UserID,
		ModelResourceID: body.ModelResourceID,
		Name:            strings.TrimSpace(body.Name),
		TokenBudget:     body.TokenBudget,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create virtual key"})
		return
	}
	k := created.Key
	c.JSON(http.StatusCreated, gin.H{
		"token": created.Token,
		"key": virtualKeyWire{
			ID: k.ID, Name: k.Name, KeyPrefix: k.KeyPrefix, ModelResourceID: k.ModelResourceID,
			TokenBudget: k.TokenBudget, Status: k.Status, CreatedAt: k.CreatedAt,
		},
	})
}

func (d *Deps) handleRevokeVirtualKey(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if d.VirtualKeys == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "virtual key service unavailable"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := d.VirtualKeys.Revoke(
		c.Request.Context(),
		id,
		tenant.OrganizationID,
		tenant.UserID,
	); err != nil {
		if errors.Is(err, virtualkeysvc.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "virtual key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke virtual key"})
		return
	}
	c.Status(http.StatusNoContent)
}
