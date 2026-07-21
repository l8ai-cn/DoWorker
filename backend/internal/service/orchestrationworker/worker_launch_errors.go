package orchestrationworker

import controlservice "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationcontrol"

var (
	ErrWorkerLaunchInProgress = controlservice.ErrWorkerLaunchInProgress
	ErrWorkerLaunchLeaseLost  = controlservice.ErrWorkerLaunchLeaseLost
)
