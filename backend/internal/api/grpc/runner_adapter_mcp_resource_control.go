package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

type MCPResourceControl interface {
	Validate(
		context.Context,
		controlservice.ValidateRequest,
	) (controlservice.ValidationResult, error)
	Plan(
		context.Context,
		controlservice.PlanRequest,
	) (controlservice.PlanResult, error)
	AuthorizeApply(context.Context, control.Scope, string) error
}

type MCPWorkerPlanApplier interface {
	Apply(
		context.Context,
		control.Scope,
		string,
	) (workerplanner.AppliedWorker, error)
}

type MCPWorkflowPlanApplier interface {
	ApplyWithStatus(
		context.Context,
		control.Scope,
		string,
		string,
	) (workerplanner.AppliedWorkflow, error)
}

type mcpResourceApplyRequest struct {
	Resource json.RawMessage `json:"resource"`
}

func (a *GRPCRunnerAdapter) planMCPResource(
	ctx context.Context,
	tc *middleware.TenantContext,
	source []byte,
	expectedKind string,
) (control.Scope, string, *mcpError) {
	if a.resourceControl == nil {
		return control.Scope{}, "", newMcpError(
			503,
			"orchestration resource service unavailable",
		)
	}
	scope, request, requestErr := mcpResourceValidateRequest(tc, source)
	if requestErr != nil {
		return control.Scope{}, "", requestErr
	}
	validated, err := a.resourceControl.Validate(ctx, request)
	if err != nil {
		return control.Scope{}, "", mapResourceControlError(err)
	}
	if issueErr := mcpBlockingIssues(validated.Issues); issueErr != nil {
		return control.Scope{}, "", issueErr
	}
	if validated.Target.Kind != expectedKind {
		return control.Scope{}, "", newMcpErrorf(
			400,
			"resource kind must be %s",
			expectedKind,
		)
	}
	planned, err := a.resourceControl.Plan(ctx, request)
	if err != nil {
		return control.Scope{}, "", mapResourceControlError(err)
	}
	if issueErr := mcpBlockingIssues(planned.Issues); issueErr != nil {
		return control.Scope{}, "", issueErr
	}
	if planned.Plan == nil || planned.Plan.ID == "" {
		return control.Scope{}, "", newMcpError(
			409,
			"orchestration resource plan is unavailable",
		)
	}
	if err := a.resourceControl.AuthorizeApply(
		ctx,
		scope,
		planned.Plan.ID,
	); err != nil {
		return control.Scope{}, "", mapResourceControlError(err)
	}
	return scope, planned.Plan.ID, nil
}

func mcpResourceValidateRequest(
	tc *middleware.TenantContext,
	source []byte,
) (control.Scope, controlservice.ValidateRequest, *mcpError) {
	if tc == nil {
		return control.Scope{}, controlservice.ValidateRequest{}, newMcpError(
			500,
			"tenant context unavailable",
		)
	}
	if len(source) == 0 {
		return control.Scope{}, controlservice.ValidateRequest{}, newMcpError(
			400,
			"resource is required",
		)
	}
	scope := control.Scope{
		OrganizationID:   tc.OrganizationID,
		OrganizationSlug: slugkit.Slug(tc.OrganizationSlug),
		ActorID:          tc.UserID,
	}
	if err := scope.Validate(); err != nil {
		return control.Scope{}, controlservice.ValidateRequest{}, newMcpError(
			500,
			"tenant resource scope unavailable",
		)
	}
	request := controlservice.ValidateRequest{
		Scope: scope,
		Source: controlservice.ResourceSource{
			Format:  controlservice.SourceFormatJSON,
			Content: append([]byte(nil), source...),
		},
	}
	return scope, request, nil
}

func mcpBlockingIssues(issues []control.PlanIssue) *mcpError {
	messages := make([]string, 0, len(issues))
	for _, issue := range issues {
		if issue.Severity != control.PlanIssueBlocking {
			continue
		}
		messages = append(
			messages,
			fmt.Sprintf("%s %s: %s", issue.Path, issue.Code, issue.Message),
		)
	}
	if len(messages) == 0 {
		return nil
	}
	return newMcpErrorf(
		400,
		"orchestration resource blocked: %s",
		strings.Join(messages, "; "),
	)
}

func mapResourceControlError(err error) *mcpError {
	switch {
	case errors.Is(err, control.ErrInvalid):
		return newMcpError(400, "invalid orchestration resource request")
	case errors.Is(err, controlservice.ErrForbidden):
		return newMcpError(403, "orchestration resource access forbidden")
	case errors.Is(err, control.ErrNotFound):
		return newMcpError(404, "orchestration resource not found")
	case errors.Is(err, control.ErrConflict),
		errors.Is(err, control.ErrStale),
		errors.Is(err, control.ErrExpired),
		errors.Is(err, control.ErrConsumed),
		errors.Is(err, controlservice.ErrStaleOptions),
		errors.Is(err, controlservice.ErrWorkerLaunchInProgress):
		return newMcpError(409, "orchestration resource state changed")
	case errors.Is(err, controlservice.ErrUnavailable):
		return newMcpError(503, "orchestration resource service unavailable")
	default:
		return newMcpError(500, "orchestration resource operation failed")
	}
}

func mcpAppliedResource(
	head control.ResourceHead,
	workerSpecSnapshotID int64,
) map[string]interface{} {
	return map[string]interface{}{
		"kind":                    head.Identity.Kind,
		"name":                    head.Identity.Name,
		"uid":                     head.Identity.UID,
		"revision":                head.Revision,
		"worker_spec_snapshot_id": workerSpecSnapshotID,
	}
}
