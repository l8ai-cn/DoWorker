package v1

import (
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerExpertRoutes(rg *gin.RouterGroup, svc *Services) {
	if svc.Expert == nil {
		return
	}
	h := NewExpertHandler(svc.Expert)
	experts := rg.Group("/experts")
	{
		experts.GET("", h.ListExperts)
		experts.POST("", h.CreateExpert)
		experts.GET("/:expertSlug", h.GetExpert)
		experts.PATCH("/:expertSlug", h.UpdateExpert)
		experts.DELETE("/:expertSlug", h.DeleteExpert)
		experts.POST("/:expertSlug/run", h.RunExpert)
	}
	rg.POST("/pods/:pod_key/publish-expert", h.PublishFromPod)
}

func registerExtExpertRoutes(rg *gin.RouterGroup, svc *Services) {
	if svc.Expert == nil {
		return
	}
	h := NewExpertHandler(svc.Expert)
	read := rg.Group("/experts")
	read.Use(middleware.RequireScope("experts:read", "experts:write", "pods:read", "pods:write"))
	{
		read.GET("", h.ListExperts)
		read.GET("/:expertSlug", h.GetExpert)
	}
	write := rg.Group("/experts")
	write.Use(middleware.RequireScope("experts:write", "pods:write"))
	{
		write.POST("", h.CreateExpert)
		write.PATCH("/:expertSlug", h.UpdateExpert)
		write.DELETE("/:expertSlug", h.DeleteExpert)
		write.POST("/:expertSlug/run", h.RunExpert)
	}
}
