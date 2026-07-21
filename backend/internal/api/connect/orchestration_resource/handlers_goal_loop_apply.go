package orchestrationresourceconnect

import (
	"context"

	"connectrpc.com/connect"

	resourcev1 "github.com/l8ai-cn/agentcloud/proto/gen/go/orchestration_resource/v1"
)

func (server *Server) CreateGoalLoopFromPlan(
	ctx context.Context,
	request *connect.Request[resourcev1.CreateGoalLoopFromPlanRequest],
) (*connect.Response[resourcev1.CreateGoalLoopFromPlanResponse], error) {
	ctx, scope, err := server.resolveScope(ctx, request.Msg)
	if err != nil {
		return nil, err
	}
	planID, err := planIDFromProto(request.Msg.GetPlanId())
	if err != nil {
		return nil, err
	}
	if err := server.authorizeApply(ctx, scope, planID); err != nil {
		return nil, err
	}
	applied, err := server.goalLoopApply.Apply(ctx, scope, planID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&resourcev1.CreateGoalLoopFromPlanResponse{
		Resource:             resourceToProto(applied.Head),
		GoalLoopId:           applied.GoalLoopID,
		WorkerSpecSnapshotId: applied.WorkerSpecSnapshotID,
		ResourceRevision:     applied.ResourceRevision,
	}), nil
}
