package goalloopconnect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/anthropics/agentsmesh/backend/internal/api/connect/interceptors"
	"github.com/anthropics/agentsmesh/backend/internal/loopscript"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	goalloopsvc "github.com/anthropics/agentsmesh/backend/internal/service/goalloop"
	goalloopv1 "github.com/anthropics/agentsmesh/proto/gen/go/goalloop/v1"
)

func (s *Server) CompileLoopProgram(
	ctx context.Context,
	req *connect.Request[goalloopv1.CompileLoopProgramRequest],
) (*connect.Response[goalloopv1.CompileLoopProgramResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	program, canonical, diagnostics := compileLoopSource(req.Msg.GetSource())
	response := &goalloopv1.CompileLoopProgramResponse{
		CanonicalSource: canonical,
		Diagnostics:     loopDiagnosticsToProto(diagnostics),
		Revision:        req.Msg.GetRevision(),
	}
	if program != nil {
		response.Program = loopProgramToProto(program)
	}
	return connect.NewResponse(response), nil
}

func (s *Server) RunLoopProgram(
	ctx context.Context,
	req *connect.Request[goalloopv1.RunLoopProgramRequest],
) (*connect.Response[goalloopv1.GoalLoop], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.service == nil {
		return nil, unavailable()
	}
	workerSnapshotID := req.Msg.GetWorkerSpecSnapshotId()
	if workerSnapshotID < 1 {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("worker_spec_snapshot_id is required"),
		)
	}
	program, _, diagnostics := compileLoopSource(req.Msg.GetSource())
	if len(diagnostics) != 0 {
		return nil, loopInvalid(diagnostics[0])
	}
	spec, diagnostics := loopscript.CompileGoalLoopV1(program)
	if len(diagnostics) != 0 {
		return nil, loopInvalid(diagnostics[0])
	}
	tenant := middleware.GetTenant(ctx)
	if err := s.service.ValidateExecutionReady(); err != nil {
		return nil, mapServiceError(err)
	}
	if err := s.service.ValidateWorkerSnapshotForExecution(
		ctx,
		tenant.OrganizationID,
		tenant.UserID,
		workerSnapshotID,
	); err != nil {
		return nil, mapServiceError(err)
	}
	tokenBudget := spec.TokenBudget
	loop, err := s.service.Create(ctx, goalloopsvc.CreateRequest{
		OrganizationID:       tenant.OrganizationID,
		CreatedByID:          tenant.UserID,
		Name:                 spec.Name,
		WorkerSpecSnapshotID: workerSnapshotID,
		Objective:            spec.Objective,
		AcceptanceCriteria:   spec.AcceptanceCriteria,
		VerificationCommand:  spec.VerificationCommand,
		MaxIterations:        spec.MaxIterations,
		TokenBudget:          &tokenBudget,
		TimeoutMinutes:       spec.TimeoutMinutes,
		NoProgressLimit:      spec.NoProgressLimit,
		SameErrorLimit:       spec.SameErrorLimit,
		EscalationPolicy:     spec.EscalationPolicy,
	})
	if err != nil {
		return nil, mapServiceError(err)
	}
	if loop == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("goal loop create returned nil"))
	}
	started, err := s.service.Start(ctx, tenant.OrganizationID, tenant.UserID, loop.Slug)
	if err != nil {
		mapped := mapServiceError(err)
		return nil, connect.NewError(
			connect.CodeOf(mapped),
			fmt.Errorf("goal loop %s created but start failed: %v", loop.Slug, err),
		)
	}
	if started == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("goal loop start returned nil"))
	}
	item, err := toProto(started)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(item), nil
}

func compileLoopSource(source string) (*loopscript.Program, string, []loopscript.Diagnostic) {
	program, diagnostics := loopscript.Parse(source)
	if len(diagnostics) != 0 {
		return nil, "", diagnostics
	}
	canonical, diagnostics := loopscript.Format(program)
	if len(diagnostics) != 0 {
		return nil, "", diagnostics
	}
	return program, canonical, nil
}

func loopInvalid(diagnostic loopscript.Diagnostic) error {
	return connect.NewError(
		connect.CodeInvalidArgument,
		fmt.Errorf("%s: %s", diagnostic.Code, diagnostic.Message),
	)
}
