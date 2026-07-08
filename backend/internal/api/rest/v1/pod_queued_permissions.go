package v1

import (
	"github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
)

func (h *PodHandler) canCancelQueuedPod(tenant *middleware.TenantContext, pod *agentpod.Pod) bool {
	if pod.CreatedByID == tenant.UserID {
		return true
	}
	return tenant.UserRole == "owner" || tenant.UserRole == "admin"
}
