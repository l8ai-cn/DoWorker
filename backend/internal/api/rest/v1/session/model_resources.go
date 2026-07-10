package sessionapi

import (
	"context"
	"net/http"
	"slices"

	airesourcedomain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	airesourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	"github.com/gin-gonic/gin"
)

type ModelResourceLister interface {
	ListEffective(
		context.Context,
		airesourcesvc.Actor,
		int64,
		[]airesourcedomain.Modality,
	) ([]airesourcesvc.EffectiveResourceView, error)
}

type modelResourceWire struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ProviderKey string `json:"provider_key"`
	Model       string `json:"model"`
	IsDefault   bool   `json:"is_default"`
}

func (d *Deps) handleListModelResources(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	if tenant == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant context is required"})
		return
	}
	if d.AIResources == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "model resource service is unavailable"})
		return
	}
	resources, err := d.AIResources.ListEffective(
		c.Request.Context(),
		airesourcesvc.Actor{UserID: tenant.UserID},
		tenant.OrganizationID,
		[]airesourcedomain.Modality{airesourcedomain.ModalityChat},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list model resources"})
		return
	}
	out := make([]modelResourceWire, 0, len(resources))
	for _, resource := range resources {
		if !resource.Selectable || resource.Resource.ID == 0 {
			continue
		}
		out = append(out, modelResourceWire{
			ID: resource.Resource.ID, Name: resource.Resource.DisplayName,
			ProviderKey: resource.Connection.ProviderKey.String(), Model: resource.Resource.ModelID,
			IsDefault: slices.Contains(
				resource.Resource.DefaultModalities,
				airesourcedomain.ModalityChat,
			),
		})
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": out})
}
