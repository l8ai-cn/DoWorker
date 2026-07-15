package orchestrationworker

import controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"

var (
	ErrWorkerLaunchInProgress = controlservice.ErrWorkerLaunchInProgress
	ErrWorkerLaunchLeaseLost  = controlservice.ErrWorkerLaunchLeaseLost
)
