package orchestrationresourceconnect

import (
	"context"

	"connectrpc.com/connect"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	service "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	resourcev1 "github.com/anthropics/agentsmesh/proto/gen/go/orchestration_resource/v1"
)

func (server *Server) ValidateResource(
	ctx context.Context,
	request *connect.Request[resourcev1.ValidateResourceRequest],
) (*connect.Response[resourcev1.ValidateResourceResponse], error) {
	ctx, scope, err := server.resolveScope(ctx, request.Msg)
	if err != nil {
		return nil, err
	}
	source, err := sourceFromProto(request.Msg.GetSource())
	if err != nil {
		return nil, err
	}
	result, err := server.service.Validate(ctx, service.ValidateRequest{
		Scope:  scope,
		Source: source,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	operation, err := operationToProto(result.Operation)
	if err != nil {
		return nil, mapServiceError(err)
	}
	issues, err := issuesToProto(result.Issues)
	if err != nil {
		return nil, mapServiceError(err)
	}
	response := &resourcev1.ValidateResourceResponse{
		Operation:     operation,
		CanonicalJson: append([]byte(nil), result.CanonicalManifest...),
		Issues:        issues,
	}
	if result.Target.TypeMeta.APIVersion != "" {
		response.Target = targetToProto(result.Target)
	}
	return connect.NewResponse(response), nil
}

func (server *Server) PlanResource(
	ctx context.Context,
	request *connect.Request[resourcev1.PlanResourceRequest],
) (*connect.Response[resourcev1.PlanResourceResponse], error) {
	ctx, scope, err := server.resolveScope(ctx, request.Msg)
	if err != nil {
		return nil, err
	}
	source, err := sourceFromProto(request.Msg.GetSource())
	if err != nil {
		return nil, err
	}
	result, err := server.service.Plan(ctx, service.PlanRequest{
		Scope:  scope,
		Source: source,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	operation, err := operationToProto(result.Operation)
	if err != nil {
		return nil, mapServiceError(err)
	}
	issues, err := issuesToProto(result.Issues)
	if err != nil {
		return nil, mapServiceError(err)
	}
	response := &resourcev1.PlanResourceResponse{
		Operation:     operation,
		CanonicalJson: append([]byte(nil), result.CanonicalManifest...),
		Issues:        issues,
	}
	if result.Target.TypeMeta.APIVersion != "" {
		response.Target = targetToProto(result.Target)
	}
	if result.Plan != nil && !hasBlockingIssue(result.Issues) {
		response.Plan, err = planToProto(*result.Plan)
		if err != nil {
			return nil, mapServiceError(err)
		}
	}
	return connect.NewResponse(response), nil
}

func hasBlockingIssue(issues []control.PlanIssue) bool {
	for index := range issues {
		if issues[index].Severity == control.PlanIssueBlocking {
			return true
		}
	}
	return false
}
