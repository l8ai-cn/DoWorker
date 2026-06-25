package v1

import "github.com/gin-gonic/gin"

func registerCoordinatorRoutes(rg *gin.RouterGroup, svc *Services) {
	if svc.Coordinator == nil {
		return
	}
	h := NewCoordinatorHandler(svc.Coordinator, svc.Repository)

	projects := rg.Group("/coordinator/projects")
	{
		projects.GET("", h.ListProjects)
		projects.POST("", h.CreateProject)
		projects.GET("/:id", h.GetProject)
		projects.PATCH("/:id", h.UpdateProject)
		projects.DELETE("/:id", h.DeleteProject)
		projects.GET("/:id/executions", h.ListExecutions)
		projects.POST("/:id/run", h.RunNow)
	}
}
