package goalloopconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	airesourcesvc "github.com/anthropics/agentsmesh/backend/internal/service/airesource"
	goalloopsvc "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
	goalloopv1 "github.com/anthropics/agentsmesh/proto/gen/go/goalloop/v1"
)

func (s *Server) GenerateLoopProgram(
	ctx context.Context,
	req *connect.Request[goalloopv1.GenerateLoopProgramRequest],
) (*connect.Response[goalloopv1.CompileLoopProgramResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.aiDrafts == nil {
		return nil, unavailableDraftGeneration()
	}
	tenant := middleware.GetTenant(ctx)
	proposal, err := s.aiDrafts.Generate(
		ctx,
		goalloopsvc.DraftGenerationScope{
			OrganizationID: tenant.OrganizationID,
			UserID:         tenant.UserID,
		},
		goalloopsvc.DraftGenerationInput{
			Prompt:          req.Msg.GetPrompt(),
			CurrentSource:   req.Msg.GetCurrentSource(),
			ModelResourceID: req.Msg.GetModelResourceId(),
			Locale:          req.Msg.GetLocale(),
		},
	)
	if err != nil {
		return nil, mapDraftGenerationError(err)
	}
	if proposal.Program == nil || proposal.CanonicalSource == "" {
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("Loop draft generator returned an empty proposal"),
		)
	}
	return connect.NewResponse(&goalloopv1.CompileLoopProgramResponse{
		CanonicalSource: proposal.CanonicalSource,
		Program:         loopProgramToProto(proposal.Program),
		Revision:        req.Msg.GetRevision(),
	}), nil
}

func (s *Server) RepairLoopProgram(
	ctx context.Context,
	req *connect.Request[goalloopv1.RepairLoopProgramRequest],
) (*connect.Response[goalloopv1.RepairLoopProgramResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.aiDrafts == nil {
		return nil, unavailableDraftGeneration()
	}
	tenant := middleware.GetTenant(ctx)
	proposal, err := s.aiDrafts.Repair(
		ctx,
		goalloopsvc.DraftGenerationScope{
			OrganizationID: tenant.OrganizationID,
			UserID:         tenant.UserID,
		},
		goalloopsvc.DraftRepairInput{
			Source:          req.Msg.GetSource(),
			ModelResourceID: req.Msg.GetModelResourceId(),
			Locale:          req.Msg.GetLocale(),
			DiagnosticCode:  req.Msg.GetDiagnosticCode(),
			NodeID:          req.Msg.GetNodeId(),
			FieldPath:       req.Msg.GetFieldPath(),
			Prompt:          req.Msg.GetPrompt(),
		},
	)
	if err != nil {
		return nil, mapDraftGenerationError(err)
	}
	if proposal.Program == nil || proposal.CanonicalSource == "" {
		return nil, connect.NewError(
			connect.CodeInternal,
			errors.New("Loop AI repair proposal is empty"),
		)
	}
	return connect.NewResponse(&goalloopv1.RepairLoopProgramResponse{
		Proposal: &goalloopv1.CompileLoopProgramResponse{
			CanonicalSource: proposal.CanonicalSource,
			Program:         loopProgramToProto(proposal.Program),
			Revision:        req.Msg.GetRevision(),
		},
		Patch: &goalloopv1.LoopIntegerPatch{
			NodeId: proposal.Patch.NodeID, FieldPath: proposal.Patch.FieldPath,
			OldValue: proposal.Patch.OldValue, NewValue: proposal.Patch.NewValue,
		},
	}), nil
}

func mapDraftGenerationError(err error) error {
	code, message := connect.CodeInternal, "Loop AI generation failed"
	switch {
	case errors.Is(err, goalloopsvc.ErrInvalidDraftGenerationInput),
		errors.Is(err, goalloopsvc.ErrDraftContainsSecret),
		errors.Is(err, airesourcesvc.ErrInvalidOwner),
		errors.Is(err, airesourcesvc.ErrInvalidProvider),
		errors.Is(err, airesourcesvc.ErrInvalidEndpoint),
		errors.Is(err, airesourcesvc.ErrInvalidRequirements),
		errors.Is(err, airesourcesvc.ErrIncompatibleModality),
		errors.Is(err, airesourcesvc.ErrIncompatibleCapability),
		errors.Is(err, airesourcesvc.ErrIncompatibleProtocolAdapter):
		code, message = connect.CodeInvalidArgument, "invalid Loop AI generation request"
	case errors.Is(err, airesourcesvc.ErrNotFound):
		code, message = connect.CodeNotFound, "AI resource not found"
	case errors.Is(err, airesourcesvc.ErrForbidden):
		code, message = connect.CodePermissionDenied, "AI resource access forbidden"
	case errors.Is(err, airesourcesvc.ErrDisabled),
		errors.Is(err, airesourcesvc.ErrUnhealthy),
		errors.Is(err, airesourcesvc.ErrUnchecked),
		errors.Is(err, goalloopsvc.ErrDraftSourceInvalid),
		errors.Is(err, goalloopsvc.ErrDraftRepairUnsupported),
		errors.Is(err, goalloopsvc.ErrDraftRepairTargetStale),
		errors.Is(err, goalloopsvc.ErrGeneratedDraftInvalid):
		code, message = connect.CodeFailedPrecondition, "Loop AI generation prerequisites are not satisfied"
	case errors.Is(err, goalloopsvc.ErrDraftGenerationUnavailable),
		errors.Is(err, goalloopsvc.ErrDraftProviderUnavailable):
		code, message = connect.CodeUnavailable, "Loop AI generation is unavailable"
	}
	return connect.NewError(code, errors.New(message))
}

func unavailableDraftGeneration() error {
	return connect.NewError(
		connect.CodeUnavailable,
		errors.New("Loop AI generation is not configured"),
	)
}
