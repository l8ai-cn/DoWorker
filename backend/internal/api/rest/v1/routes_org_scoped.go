package v1

import (
	"github.com/gin-gonic/gin"
)

func RegisterOrgScopedRoutes(rg *gin.RouterGroup, svc *Services, previewPublicOrigin string) {
	registerRunnerRoutes(rg, svc)
	registerBillingRoutes(rg, svc)
	registerCoordinatorRoutes(rg, svc)
	registerPodQueueRoutes(rg, svc, previewPublicOrigin)
	registerExpertRoutes(rg, svc)
	registerSkillRoutes(rg, svc)
	registerIMBridgeRoutes(rg, svc)
}
