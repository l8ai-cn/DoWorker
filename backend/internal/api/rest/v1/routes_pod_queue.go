package v1

import (
	runnersvc "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

func registerPodQueueRoutes(rg *gin.RouterGroup, svc *Services) {
	var podOpts []PodHandlerOption
	if svc.PodCoordinator != nil {
		podOpts = append(podOpts, WithPodCoordinator(svc.PodCoordinator))
	}
	if svc.PendingQueue != nil {
		podOpts = append(podOpts, WithPendingQueue(svc.PendingQueue))
	}
	podHandler := NewPodHandler(svc.Pod, svc.Runner, svc.PodOrchestrator, podOpts...)

	rg.POST("/quick-tasks", podHandler.CreateQuickTask)
	rg.GET("/pods/queued", podHandler.ListQueuedPods)
	rg.DELETE("/pods/:key/queue", podHandler.CancelQueuedPod)
}

var _ pendingQueueReader = (*runnersvc.PendingCommandQueue)(nil)
