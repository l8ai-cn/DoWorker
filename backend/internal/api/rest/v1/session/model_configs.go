package sessionapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/anthropics/agentsmesh/backend/internal/domain/aimodel"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	aimodelsvc "github.com/anthropics/agentsmesh/backend/internal/service/aimodel"
	"github.com/gin-gonic/gin"
)

type modelConfigWire struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
	Model        string `json:"model"`
	BaseURL      string `json:"base_url"`
	IsDefault    bool   `json:"is_default"`
	Scope        string `json:"scope"`
	TokenBudget  *int64 `json:"token_budget,omitempty"`
}

func toModelConfigWire(m *aimodel.AIModel) modelConfigWire {
	scope := "user"
	if m.OrganizationID != nil && m.UserID == nil {
		scope = "org"
	}
	return modelConfigWire{
		ID: m.ID, Name: m.Name, ProviderType: m.ProviderType, Model: m.Model,
		BaseURL: m.BaseURL, IsDefault: m.IsDefault, Scope: scope, TokenBudget: m.TokenBudget,
	}
}

func (d *Deps) handleListModelConfigs(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.AIModels == nil {
		c.JSON(http.StatusOK, gin.H{"object": "list", "data": []modelConfigWire{}})
		return
	}
	models, err := d.AIModels.ListVisible(c.Request.Context(), tenant.UserID, tenant.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list models"})
		return
	}
	out := make([]modelConfigWire, 0, len(models))
	for _, m := range models {
		out = append(out, toModelConfigWire(m))
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": out})
}

type createModelConfigBody struct {
	Name         string            `json:"name"`
	ProviderType string            `json:"provider_type"`
	Model        string            `json:"model"`
	BaseURL      string            `json:"base_url"`
	Credentials  map[string]string `json:"credentials"`
	IsDefault    bool              `json:"is_default"`
	TokenBudget  *int64            `json:"token_budget"`
	Scope        string            `json:"scope"` // "org" | "user" (default user)
}

func (d *Deps) handleCreateModelConfig(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil || d.AIModels == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "model service unavailable"})
		return
	}
	var body createModelConfigBody
	if err := c.ShouldBindJSON(&body); err != nil ||
		strings.TrimSpace(body.Name) == "" || strings.TrimSpace(body.ProviderType) == "" ||
		strings.TrimSpace(body.Model) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, provider_type, model are required"})
		return
	}
	in := aimodelsvc.CreateInput{
		Name:         strings.TrimSpace(body.Name),
		ProviderType: strings.TrimSpace(body.ProviderType),
		Model:        strings.TrimSpace(body.Model),
		BaseURL:      strings.TrimSpace(body.BaseURL),
		Credentials:  body.Credentials,
		IsDefault:    body.IsDefault,
		TokenBudget:  body.TokenBudget,
	}
	if body.Scope == "org" {
		orgID := tenant.OrganizationID
		in.OrgID = &orgID
	} else {
		userID := tenant.UserID
		in.UserID = &userID
	}
	m, err := d.AIModels.Create(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create model"})
		return
	}
	c.JSON(http.StatusOK, toModelConfigWire(m))
}

func (d *Deps) handleDeleteModelConfig(c *gin.Context) {
	if d.AIModels == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "model service unavailable"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := d.AIModels.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete model"})
		return
	}
	c.Status(http.StatusNoContent)
}
