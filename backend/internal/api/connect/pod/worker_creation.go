package podconnect

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
	podv1 "github.com/anthropics/agentsmesh/proto/gen/go/pod/v1"
)

func (s *Server) ListWorkerCreateOptions(
	ctx context.Context,
	req *connect.Request[podv1.ListWorkerCreateOptionsRequest],
) (*connect.Response[podv1.ListWorkerCreateOptionsResponse], error) {
	if s.workerCreation == nil {
		return nil, connect.NewError(
			connect.CodeUnavailable,
			errors.New("worker creation service not configured"),
		)
	}
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	tenant := middleware.GetTenant(ctx)
	options, err := s.workerCreation.ListOptions(
		ctx,
		specservice.Scope{OrgID: tenant.OrganizationID, UserID: tenant.UserID},
		workercreation.OptionsFilter{
			WorkerTypeSlug:  req.Msg.GetWorkerTypeSlug(),
			ComputeTargetID: optionalInt64(req.Msg.ComputeTargetId),
			DeploymentMode:  workerDeploymentMode(req.Msg.GetDeploymentMode()),
		},
	)
	if err != nil {
		return nil, mapWorkerCreationError(err)
	}
	response, err := workerCreateOptionsToProto(options)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(response), nil
}

func (s *Server) PreflightWorker(
	ctx context.Context,
	req *connect.Request[podv1.PreflightWorkerRequest],
) (*connect.Response[podv1.PreflightWorkerResponse], error) {
	if s.workerCreation == nil {
		return nil, connect.NewError(
			connect.CodeUnavailable,
			errors.New("worker creation service not configured"),
		)
	}
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	draft, err := workerDraftFromProto(req.Msg.Draft)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	tenant := middleware.GetTenant(ctx)
	result, err := s.workerCreation.Preflight(
		ctx,
		specservice.Scope{OrgID: tenant.OrganizationID, UserID: tenant.UserID},
		draft,
	)
	if err != nil {
		return nil, mapWorkerCreationError(err)
	}
	response := &podv1.PreflightWorkerResponse{
		Issues:          workerIssuesToProto(result.BlockingErrors, result.Warnings),
		OptionsRevision: result.OptionsRevision,
	}
	if result.Resolved != nil {
		resolved := string(result.Resolved.Snapshot.SpecJSON())
		if resolved == "" {
			return nil, connect.NewError(
				connect.CodeInternal,
				errors.New("worker preflight returned an empty resolved spec"),
			)
		}
		response.ResolvedSpecJson = &resolved
	}
	return connect.NewResponse(response), nil
}

func (s *Server) FillWorkerDraft(
	ctx context.Context,
	req *connect.Request[podv1.FillWorkerDraftRequest],
) (*connect.Response[podv1.FillWorkerDraftResponse], error) {
	if s.workerDraftFiller == nil {
		return nil, connect.NewError(
			connect.CodeUnavailable,
			errors.New("worker draft filler not configured"),
		)
	}
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	prompt := strings.TrimSpace(req.Msg.GetPrompt())
	if prompt == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("prompt is required"),
		)
	}
	var current *workercreation.Draft
	if req.Msg.CurrentDraft != nil {
		decoded, err := workerDraftFromProto(req.Msg.CurrentDraft)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		current = &decoded
	}
	tenant := middleware.GetTenant(ctx)
	result, err := s.workerDraftFiller.Fill(
		ctx,
		specservice.Scope{OrgID: tenant.OrganizationID, UserID: tenant.UserID},
		prompt,
		req.Msg.GetGenerationModelResourceId(),
		current,
	)
	if err != nil {
		return nil, mapWorkerCreationError(err)
	}
	draft, err := workerDraftToProto(result.Draft)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&podv1.FillWorkerDraftResponse{
		Draft:  draft,
		Issues: workerIssuesToProto(result.Issues, nil),
	}), nil
}

func mapWorkerCreationError(err error) error {
	switch {
	case errors.Is(err, specservice.ErrInvalidScope),
		errors.Is(err, specservice.ErrInvalidDraft),
		errors.Is(err, workercreation.ErrStaleOptions):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, specservice.ErrResolverUnavailable):
		return connect.NewError(connect.CodeUnavailable, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
