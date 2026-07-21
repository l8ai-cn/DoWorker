package v1

import (
	"context"

	podDomain "github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	workerplanner "github.com/l8ai-cn/agentcloud/backend/internal/service/orchestrationworker"
)

type QuickTaskRequest struct {
	PlanID string `json:"plan_id"`
}

type QuickTaskResponse struct {
	PodKey        string `json:"pod_key"`
	Status        string `json:"status"`
	QueuePosition int    `json:"queue_position,omitempty"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

type QuickTaskPlanApplier interface {
	Apply(
		context.Context,
		control.Scope,
		string,
	) (workerplanner.AppliedWorker, error)
}

type QuickTaskPlanAuthorizer interface {
	AuthorizeApply(context.Context, control.Scope, string) error
}

type quickTaskPodReader interface {
	GetPod(context.Context, string) (*podDomain.Pod, error)
}
