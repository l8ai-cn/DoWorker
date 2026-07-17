package v1

import (
	runnersvc "github.com/anthropics/agentsmesh/backend/internal/service/runner"
	"github.com/gin-gonic/gin"
)

func registerPodQueueRoutes(rg *gin.RouterGroup, svc *Services, previewPublicOrigin string) {
	var podOpts []PodHandlerOption
	if svc.PodCoordinator != nil {
		podOpts = append(podOpts, WithPodCoordinator(svc.PodCoordinator))
		if sender := svc.PodCoordinator.GetCommandSender(); sender != nil {
			podOpts = append(podOpts, WithCommandSender(sender))
		}
	}
	if svc.PendingQueue != nil {
		podOpts = append(podOpts, WithPendingQueue(svc.PendingQueue))
	}
	if svc.Grant != nil {
		podOpts = append(podOpts, WithGrantServiceForPod(svc.Grant))
	}
	if svc.RelayManager != nil && svc.RelayTokenGenerator != nil {
		podOpts = append(podOpts, WithRelayPreview(svc.RelayManager, svc.RelayTokenGenerator, previewPublicOrigin))
	}
	if svc.SandboxFsService != nil {
		podOpts = append(podOpts, WithPodWorkspaceSandbox(svc.SandboxFsService))
	}
	if svc.File != nil {
		podOpts = append(podOpts, WithPodWorkspaceArtifactTransfer(svc.File))
	}
	podHandler := NewPodHandler(svc.Pod, svc.Runner, svc.PodOrchestrator, podOpts...)

	rg.POST("/quick-tasks", podHandler.CreateQuickTask)
	rg.DELETE("/pods/:key", podHandler.DeletePod)
	rg.GET("/pods/queued", podHandler.ListQueuedPods)
	rg.DELETE("/pods/:key/queue", podHandler.CancelQueuedPod)
	rg.GET("/pods/:key/preview", podHandler.GetPodPreview)
	rg.GET("/pods/:key/resources/workspace/changes", podHandler.ListWorkspaceArtifacts)
	rg.GET("/pods/:key/resources/workspace/filesystem/*filepath", podHandler.ReadWorkspaceArtifact)
	rg.GET("/pods/:key/resources/workspace/artifacts/*filepath", podHandler.TransferWorkspaceArtifact)
}

var _ pendingQueueReader = (*runnersvc.PendingCommandQueue)(nil)
