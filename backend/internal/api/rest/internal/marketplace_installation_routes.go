package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	"github.com/gin-gonic/gin"
)

type MarketplaceExpertInstaller interface {
	InstallMarketplaceExpert(
		context.Context,
		expertsvc.MarketplaceInstallationRequest,
	) (*expertdom.Expert, bool, error)
}

type MarketplaceOrganizationAuthorizer interface {
	IsMember(context.Context, int64, int64) (bool, error)
}

type MarketplaceInstallationDeps struct {
	Installer      MarketplaceExpertInstaller
	Authorizer     MarketplaceOrganizationAuthorizer
	InternalSecret string
}

type marketplaceInstallRequest struct {
	InstallationID       string          `json:"installation_id"`
	PlatformResourceType string          `json:"platform_resource_type"`
	PlatformResourceID   int64           `json:"platform_resource_id"`
	SourceReleaseID      int64           `json:"source_release_id"`
	TargetOrganizationID int64           `json:"target_platform_organization_id"`
	ActorUserID          int64           `json:"actor_platform_user_id"`
	RuntimeSnapshot      json.RawMessage `json:"runtime_snapshot"`
	Configuration        json.RawMessage `json:"configuration"`
}

type marketplaceAuthorizeRequest struct {
	TargetOrganizationID int64 `json:"target_platform_organization_id"`
	ActorUserID          int64 `json:"actor_platform_user_id"`
}

func RegisterMarketplaceInstallationRoutes(
	router *gin.RouterGroup,
	deps MarketplaceInstallationDeps,
) {
	if deps.Installer == nil || deps.Authorizer == nil {
		panic("marketplace installer and organization authorizer are required")
	}
	router.Use(InternalAPIAuth(deps.InternalSecret))
	router.POST("/authorize", func(c *gin.Context) {
		var request marketplaceAuthorizeRequest
		if c.ShouldBindJSON(&request) != nil ||
			request.TargetOrganizationID <= 0 ||
			request.ActorUserID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "INVALID_MARKETPLACE_AUTHORIZATION"},
			})
			return
		}
		allowed, err := deps.Authorizer.IsMember(
			c,
			request.TargetOrganizationID,
			request.ActorUserID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "MARKETPLACE_AUTHORIZATION_FAILED"},
			})
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "TARGET_ORGANIZATION_FORBIDDEN"},
			})
			return
		}
		c.Status(http.StatusNoContent)
	})
	router.POST("/apply", func(c *gin.Context) {
		var request marketplaceInstallRequest
		var configuration struct {
			ModelResourceID      int64            `json:"model_resource_id"`
			ToolModelResourceIDs map[string]int64 `json:"tool_model_resource_ids"`
		}
		if c.ShouldBindJSON(&request) != nil ||
			request.PlatformResourceType != "expert" ||
			json.Unmarshal(request.Configuration, &configuration) != nil ||
			configuration.ModelResourceID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "INVALID_RUNTIME_INSTALLATION"},
			})
			return
		}
		row, existing, err := deps.Installer.InstallMarketplaceExpert(
			c,
			expertsvc.MarketplaceInstallationRequest{
				InstallationID:            request.InstallationID,
				TargetOrganizationID:      request.TargetOrganizationID,
				ActorUserID:               request.ActorUserID,
				ModelResourceID:           configuration.ModelResourceID,
				ToolModelResourceIDs:      configuration.ToolModelResourceIDs,
				SourceMarketApplicationID: request.PlatformResourceID,
				SourceMarketReleaseID:     request.SourceReleaseID,
				RuntimeSnapshot:           request.RuntimeSnapshot,
			},
		)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": gin.H{
					"code":    "RUNTIME_INSTALLATION_FAILED",
					"message": "应用运行时安装失败",
				},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"runtime_ref": "expert:" + strconv.FormatInt(row.ID, 10),
			"result": gin.H{
				"expert_id":         strconv.FormatInt(row.ID, 10),
				"already_installed": existing,
			},
		})
	})
}
