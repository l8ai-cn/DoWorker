package autopilotconnect

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"

	"github.com/l8ai-cn/agentcloud/backend/internal/api/connect/interceptors"
	"github.com/l8ai-cn/agentcloud/backend/internal/domain/agentpod"
	"github.com/l8ai-cn/agentcloud/backend/internal/infra/eventbus"
	"github.com/l8ai-cn/agentcloud/backend/internal/middleware"
	agentpodsvc "github.com/l8ai-cn/agentcloud/backend/internal/service/agentpod"
	apv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/autopilot/v1"
	eventsv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/events/v1"
	runnerv1 "github.com/l8ai-cn/agentcloud/proto/gen/go/runner/v1"
)

// CreateAutopilotController — REST analogue: POST /autopilot-controllers.
func (s *Server) CreateAutopilotController(
	ctx context.Context, req *connect.Request[apv1.CreateAutopilotControllerRequest],
) (*connect.Response[apv1.AutopilotController], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req.Msg, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.svc == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("autopilot service not configured"))
	}
	if s.podSvc == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("pod service not configured"))
	}
	if req.Msg.GetPodKey() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("pod_key is required"))
	}

	tenant := middleware.GetTenant(ctx)
	pod, err := s.podSvc.GetPod(ctx, req.Msg.GetPodKey())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("target pod not found"))
	}
	if pod.OrganizationID != tenant.OrganizationID {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("forbidden"))
	}

	controller, err := s.svc.CreateAndStart(ctx, &agentpodsvc.CreateAndStartRequest{
		OrganizationID:        tenant.OrganizationID,
		Pod:                   pod,
		Prompt:                req.Msg.GetPrompt(),
		MaxIterations:         req.Msg.GetMaxIterations(),
		IterationTimeoutSec:   req.Msg.GetIterationTimeoutSec(),
		NoProgressThreshold:   req.Msg.GetNoProgressThreshold(),
		SameErrorThreshold:    req.Msg.GetSameErrorThreshold(),
		ApprovalTimeoutMin:    req.Msg.GetApprovalTimeoutMin(),
		ControlAgentSlug:      req.Msg.GetControlAgentSlug(),
		ControlPromptTemplate: req.Msg.GetControlPromptTemplate(),
		MCPConfigJSON:         req.Msg.GetMcpConfigJson(),
		KeyPrefix:             "autopilot",
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	s.publishAutopilotCreated(ctx, tenant.OrganizationID, controller, pod.PodKey)
	return connect.NewResponse(toProtoController(controller)), nil
}

func (s *Server) publishAutopilotCreated(ctx context.Context, orgID int64, c *agentpod.AutopilotController, podKey string) {
	if s.eventBus == nil || c == nil {
		return
	}
	data := &eventsv1.AutopilotCreatedEventData{
		AutopilotControllerKey: c.AutopilotControllerKey,
		PodKey:                 podKey,
	}
	event, err := eventbus.NewEntityEvent(eventbus.EventAutopilotCreated, orgID, "autopilot_controller", c.AutopilotControllerKey, data)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build autopilot:created event", "key", c.AutopilotControllerKey, "error", err)
		return
	}
	if err := s.eventBus.Publish(ctx, event); err != nil {
		slog.ErrorContext(ctx, "failed to publish autopilot:created event", "key", c.AutopilotControllerKey, "error", err)
	}
}

// PauseAutopilotController — REST analogue: POST /autopilot-controllers/:key/pause.
func (s *Server) PauseAutopilotController(
	ctx context.Context, req *connect.Request[apv1.ActionRequest],
) (*connect.Response[apv1.ActionResponse], error) {
	return s.sendControl(ctx, req.Msg, "pause", func(c *runnerv1.AutopilotControlCommand) {
		c.Action = &runnerv1.AutopilotControlCommand_Pause{Pause: &runnerv1.AutopilotPauseAction{}}
	})
}

// ResumeAutopilotController — REST analogue: POST /autopilot-controllers/:key/resume.
func (s *Server) ResumeAutopilotController(
	ctx context.Context, req *connect.Request[apv1.ActionRequest],
) (*connect.Response[apv1.ActionResponse], error) {
	return s.sendControl(ctx, req.Msg, "resume", func(c *runnerv1.AutopilotControlCommand) {
		c.Action = &runnerv1.AutopilotControlCommand_Resume{Resume: &runnerv1.AutopilotResumeAction{}}
	})
}

// StopAutopilotController — REST analogue: POST /autopilot-controllers/:key/stop.
func (s *Server) StopAutopilotController(
	ctx context.Context, req *connect.Request[apv1.ActionRequest],
) (*connect.Response[apv1.ActionResponse], error) {
	return s.sendControl(ctx, req.Msg, "stop", func(c *runnerv1.AutopilotControlCommand) {
		c.Action = &runnerv1.AutopilotControlCommand_Stop{Stop: &runnerv1.AutopilotStopAction{}}
	})
}

// TakeoverAutopilotController — REST analogue: POST /autopilot-controllers/:key/takeover.
func (s *Server) TakeoverAutopilotController(
	ctx context.Context, req *connect.Request[apv1.ActionRequest],
) (*connect.Response[apv1.ActionResponse], error) {
	return s.sendControl(ctx, req.Msg, "takeover", func(c *runnerv1.AutopilotControlCommand) {
		c.Action = &runnerv1.AutopilotControlCommand_Takeover{Takeover: &runnerv1.AutopilotTakeoverAction{}}
	})
}

// HandbackAutopilotController — REST analogue: POST /autopilot-controllers/:key/handback.
func (s *Server) HandbackAutopilotController(
	ctx context.Context, req *connect.Request[apv1.ActionRequest],
) (*connect.Response[apv1.ActionResponse], error) {
	return s.sendControl(ctx, req.Msg, "handback", func(c *runnerv1.AutopilotControlCommand) {
		c.Action = &runnerv1.AutopilotControlCommand_Handback{Handback: &runnerv1.AutopilotHandbackAction{}}
	})
}

// ApproveAutopilotController — REST analogue: POST /autopilot-controllers/:key/approve.
// Carries continue_execution (default true) + additional_iterations.
func (s *Server) ApproveAutopilotController(
	ctx context.Context, req *connect.Request[apv1.ApproveRequest],
) (*connect.Response[apv1.ActionResponse], error) {
	actionReq := &apv1.ActionRequest{
		OrgSlug: req.Msg.GetOrgSlug(),
		Key:     req.Msg.GetKey(),
	}
	continueExec := true
	if req.Msg.ContinueExecution != nil {
		continueExec = req.Msg.GetContinueExecution()
	}
	return s.sendControl(ctx, actionReq, "approve", func(c *runnerv1.AutopilotControlCommand) {
		c.Action = &runnerv1.AutopilotControlCommand_Approve{
			Approve: &runnerv1.AutopilotApproveAction{
				ContinueExecution:    continueExec,
				AdditionalIterations: req.Msg.GetAdditionalIterations(),
			},
		}
	})
}

// sendControl resolves org + loads the AP record + sends the runner command.
// `applyAction` sets the specific oneof variant on the command so the
// handler stays agnostic to which control was issued.
func (s *Server) sendControl(
	ctx context.Context,
	req *apv1.ActionRequest,
	action string,
	applyAction func(*runnerv1.AutopilotControlCommand),
) (*connect.Response[apv1.ActionResponse], error) {
	ctx, _, err := interceptors.ResolveOrgScope(ctx, req, s.orgSvc)
	if err != nil {
		return nil, err
	}
	if s.svc == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("autopilot service not configured"))
	}
	if s.commandSender == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("command sender not configured"))
	}
	if req.GetKey() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("key is required"))
	}
	tenant := middleware.GetTenant(ctx)
	ap, err := s.svc.GetAutopilotController(ctx, tenant.OrganizationID, req.GetKey())
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("autopilot pod not found"))
	}
	cmd := &runnerv1.AutopilotControlCommand{
		AutopilotKey: ap.AutopilotControllerKey,
	}
	applyAction(cmd)
	if err := s.commandSender.SendAutopilotControl(ap.RunnerID, cmd); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&apv1.ActionResponse{Status: "ok", Action: action}), nil
}
