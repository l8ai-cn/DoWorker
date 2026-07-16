package orchestrationresourceconnect

import (
	"context"

	"connectrpc.com/connect"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resourcev1 "github.com/anthropics/agentsmesh/proto/gen/go/orchestration_resource/v1"
)

func (server *Server) ApplyBindingResourcePlan(
	ctx context.Context,
	request *connect.Request[resourcev1.ApplyBindingResourcePlanRequest],
) (*connect.Response[resourcev1.Resource], error) {
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
	head, err := server.bindingApply.Apply(ctx, scope, planID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(resourceToProto(head)), nil
}

func (server *Server) ApplyWorkerTemplatePlan(
	ctx context.Context,
	request *connect.Request[resourcev1.ApplyWorkerTemplatePlanRequest],
) (*connect.Response[resourcev1.ApplyWorkerTemplatePlanResponse], error) {
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
	applied, err := server.workerTemplateApply.Apply(ctx, scope, planID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&resourcev1.ApplyWorkerTemplatePlanResponse{
		Resource:             resourceToProto(applied.Head),
		WorkerSpecSnapshotId: applied.WorkerSpecSnapshotID,
	}), nil
}

func (server *Server) CreateWorkerFromPlan(
	ctx context.Context,
	request *connect.Request[resourcev1.CreateWorkerFromPlanRequest],
) (*connect.Response[resourcev1.CreateWorkerFromPlanResponse], error) {
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
	applied, err := server.workerApply.Apply(ctx, scope, planID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&resourcev1.CreateWorkerFromPlanResponse{
		Resource:             resourceToProto(applied.Head),
		LaunchId:             applied.LaunchID,
		PodId:                applied.PodID,
		PodKey:               applied.PodKey,
		WorkerSpecSnapshotId: applied.WorkerSpecSnapshotID,
		ResourceRevision:     applied.ResourceRevision,
		RunnerId:             applied.RunnerID,
	}), nil
}

func (server *Server) ApplyPromptPlan(
	ctx context.Context,
	request *connect.Request[resourcev1.ApplyPromptPlanRequest],
) (*connect.Response[resourcev1.Resource], error) {
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
	head, err := server.promptApply.Apply(ctx, scope, planID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(resourceToProto(head)), nil
}

func (server *Server) ApplyExpertPlan(
	ctx context.Context,
	request *connect.Request[resourcev1.ApplyExpertPlanRequest],
) (*connect.Response[resourcev1.ApplyExpertPlanResponse], error) {
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
	applied, err := server.expertApply.Apply(ctx, scope, planID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&resourcev1.ApplyExpertPlanResponse{
		Resource:             resourceToProto(applied.Head),
		ExpertId:             applied.ExpertID,
		WorkerSpecSnapshotId: applied.WorkerSpecSnapshotID,
		ResourceRevision:     applied.ResourceRevision,
	}), nil
}

func (server *Server) ApplyWorkflowPlan(
	ctx context.Context,
	request *connect.Request[resourcev1.ApplyWorkflowPlanRequest],
) (*connect.Response[resourcev1.ApplyWorkflowPlanResponse], error) {
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
	applied, err := server.workflowApply.Apply(ctx, scope, planID)
	if err != nil {
		return nil, mapServiceError(err)
	}
	return connect.NewResponse(&resourcev1.ApplyWorkflowPlanResponse{
		Resource:             resourceToProto(applied.Head),
		WorkflowId:           applied.WorkflowID,
		WorkerSpecSnapshotId: applied.WorkerSpecSnapshotID,
		ResourceRevision:     applied.ResourceRevision,
	}), nil
}

func (server *Server) authorizeApply(
	ctx context.Context,
	scope control.Scope,
	planID string,
) error {
	if err := server.service.AuthorizeApply(ctx, scope, planID); err != nil {
		return mapServiceError(err)
	}
	return nil
}
