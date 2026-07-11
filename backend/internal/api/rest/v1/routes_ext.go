package v1

import (
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterExtRoutes registers third-party API key-authenticated routes.
// These routes reuse existing handler instances with scope-based access control.
func RegisterExtRoutes(rg *gin.RouterGroup, svc *Services) {
	// Pod routes
	var podOpts []PodHandlerOption
	if svc.PodCoordinator != nil {
		podOpts = append(podOpts, WithPodCoordinator(svc.PodCoordinator))
	}
	if svc.PodCoordinator != nil {
		if sender := svc.PodCoordinator.GetCommandSender(); sender != nil {
			podOpts = append(podOpts, WithCommandSender(sender))
		}
	}
	podHandler := NewPodHandler(svc.Pod, svc.Runner, svc.PodOrchestrator, podOpts...)
	registerExtPodWorkerRoutes(rg, podHandler)

	// Ticket routes
	ticketHandler := NewTicketHandler(svc.Ticket)

	ticketsRead := rg.Group("/tickets")
	ticketsRead.Use(middleware.RequireScope("tickets:read", "tickets:write"))
	{
		ticketsRead.GET("", ticketHandler.ListTickets)
		ticketsRead.GET("/board", ticketHandler.GetBoard)
		ticketsRead.GET("/:ticket_slug", ticketHandler.GetTicket)
	}
	ticketsWrite := rg.Group("/tickets")
	ticketsWrite.Use(middleware.RequireScope("tickets:write"))
	{
		ticketsWrite.POST("", ticketHandler.CreateTicket)
		ticketsWrite.PUT("/:ticket_slug", ticketHandler.UpdateTicket)
		ticketsWrite.PATCH("/:ticket_slug/status", ticketHandler.UpdateTicketStatus)
		ticketsWrite.DELETE("/:ticket_slug", ticketHandler.DeleteTicket)
	}

	// Channel routes
	channelHandler := NewChannelHandler(svc.Channel, svc.Ticket)

	channelsRead := rg.Group("/channels")
	channelsRead.Use(middleware.RequireScope("channels:read", "channels:write"))
	{
		channelsRead.GET("", channelHandler.ListChannels)
		channelsRead.GET("/:id", channelHandler.GetChannel)
		channelsRead.GET("/:id/messages", channelHandler.ListMessages)
	}
	channelsWrite := rg.Group("/channels")
	channelsWrite.Use(middleware.RequireScope("channels:write"))
	{
		channelsWrite.POST("", channelHandler.CreateChannel)
		channelsWrite.PUT("/:id", channelHandler.UpdateChannel)
		channelsWrite.POST("/:id/messages", channelHandler.SendMessage)
	}

	// Runner routes (read-only)
	var runnerOpts []RunnerHandlerOption
	if svc.Pod != nil {
		runnerOpts = append(runnerOpts, WithPodServiceForRunner(svc.Pod))
	}
	if svc.PodCoordinator != nil {
		runnerOpts = append(runnerOpts, WithPodCoordinatorForRunner(svc.PodCoordinator))
	}
	runnerHandler := NewRunnerHandler(svc.Runner, runnerOpts...)

	runnersRead := rg.Group("/runners")
	runnersRead.Use(middleware.RequireScope("runners:read"))
	{
		runnersRead.GET("", runnerHandler.ListRunners)
		runnersRead.GET("/:id", runnerHandler.GetRunner)
		runnersRead.GET("/available", runnerHandler.ListAvailableRunners)
		runnersRead.GET("/:id/pods", runnerHandler.ListRunnerPods)
	}

	// Repository routes (read-only)
	repositoryHandler := NewRepositoryHandler(svc.Repository)

	reposRead := rg.Group("/repositories")
	reposRead.Use(middleware.RequireScope("repos:read"))
	{
		reposRead.GET("", repositoryHandler.ListRepositories)
		reposRead.GET("/:id", repositoryHandler.GetRepository)
		reposRead.GET("/:id/branches", repositoryHandler.ListBranches)
		reposRead.GET("/:id/merge-requests", repositoryHandler.ListRepositoryMergeRequests)
	}

	registerExtExpertRoutes(rg, svc)

	// Workflow routes
	if svc.Workflow != nil && svc.WorkflowOrchestrator != nil {
		workflowHandler := NewWorkflowHandler(svc.Workflow, svc.WorkflowRun, svc.WorkflowOrchestrator, svc.PodCoordinator)

		workflowsRead := rg.Group("/workflows")
		workflowsRead.Use(middleware.RequireScope("workflows:read", "workflows:write"))
		{
			workflowsRead.GET("", workflowHandler.ListWorkflows)
			workflowsRead.GET("/:workflow_slug", workflowHandler.GetWorkflow)
			workflowsRead.GET("/:workflow_slug/runs", workflowHandler.ListWorkflowRuns)
			workflowsRead.GET("/:workflow_slug/runs/:run_id", workflowHandler.GetRun)
		}
		workflowsWrite := rg.Group("/workflows")
		workflowsWrite.Use(middleware.RequireScope("workflows:write"))
		{
			workflowsWrite.POST("/:workflow_slug/trigger", workflowHandler.TriggerWorkflow)
			workflowsWrite.POST("/:workflow_slug/runs/:run_id/cancel", workflowHandler.CancelRun)
		}
	}
}
