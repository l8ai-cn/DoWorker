package v1

import (
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// registerExtPodWorkerRoutes mounts the same handlers at /pods and /workers.
// "Worker" is the product term; "Pod" remains a route alias.
func registerExtPodWorkerRoutes(rg *gin.RouterGroup, podHandler *PodHandler) {
	for _, base := range []string{"/pods", "/workers"} {
		read := rg.Group(base)
		read.Use(middleware.RequireScope("pods:read", "pods:write"))
		{
			read.GET("", podHandler.ListPods)
			read.GET("/:key", podHandler.GetPod)
		}
		write := rg.Group(base)
		write.Use(middleware.RequireScope("pods:write"))
		{
			write.POST("", podHandler.CreatePod)
			write.POST("/:key/prompt", podHandler.SendPodPrompt)
			write.POST("/:key/terminate", podHandler.TerminatePod)
		}
	}
}
