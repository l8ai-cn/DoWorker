package goalloopconnect

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	"github.com/l8ai-cn/agentcloud/backend/internal/loopscript"
	goalloopv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/goalloop/v1"
)

func (s *Server) CompileLoopProgram(
	ctx context.Context,
	req *connect.Request[goalloopv1.CompileLoopProgramRequest],
) (*connect.Response[goalloopv1.CompileLoopProgramResponse], error) {
	if _, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc); err != nil {
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
	if _, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc); err != nil {
		return nil, err
	}
	return nil, connect.NewError(
		connect.CodeFailedPrecondition,
		errors.New("loop programs must use orchestration validate-plan-apply before explicit Start"),
	)
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
