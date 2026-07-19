package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const orchestrationResourceService = "/proto.orchestration_resource.v1." +
	"OrchestrationResourceService/"

type AppliedOrchestrationResource struct {
	Revision             int64
	WorkerSpecSnapshotID int64
	WorkflowID           int64
}

type resourceIssueWire struct {
	Severity string `json:"severity"`
	Path     string `json:"path"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

func (r *REST) validateAndPlanResource(
	ctx context.Context,
	orgSlug string,
	manifest any,
) (string, error) {
	content, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("encode orchestration resource: %w", err)
	}
	request := map[string]any{
		"orgSlug": orgSlug,
		"source": map[string]string{
			"format":  "SOURCE_FORMAT_JSON",
			"content": base64.StdEncoding.EncodeToString(content),
		},
	}
	var validation struct {
		Issues []resourceIssueWire `json:"issues"`
	}
	if err := r.connectCall(
		ctx,
		orchestrationResourceService+"ValidateResource",
		request,
		&validation,
	); err != nil {
		return "", err
	}
	if err := rejectBlockingResourceIssues(validation.Issues); err != nil {
		return "", fmt.Errorf("validate orchestration resource: %w", err)
	}
	var planning struct {
		Issues []resourceIssueWire `json:"issues"`
		Plan   *struct {
			PlanID string `json:"planId"`
		} `json:"plan"`
	}
	if err := r.connectCall(
		ctx,
		orchestrationResourceService+"PlanResource",
		request,
		&planning,
	); err != nil {
		return "", err
	}
	if err := rejectBlockingResourceIssues(planning.Issues); err != nil {
		return "", fmt.Errorf("plan orchestration resource: %w", err)
	}
	if planning.Plan == nil || planning.Plan.PlanID == "" {
		return "", fmt.Errorf("plan orchestration resource returned no plan")
	}
	return planning.Plan.PlanID, nil
}

func rejectBlockingResourceIssues(issues []resourceIssueWire) error {
	var messages []string
	for _, issue := range issues {
		if issue.Severity != "ISSUE_SEVERITY_BLOCKING" {
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
	return fmt.Errorf("%s", strings.Join(messages, "; "))
}

func (r *REST) ApplyOrchestrationResource(
	ctx context.Context,
	orgSlug, kind string,
	manifest any,
) (AppliedOrchestrationResource, error) {
	planID, err := r.validateAndPlanResource(ctx, orgSlug, manifest)
	if err != nil {
		return AppliedOrchestrationResource{}, err
	}
	request := map[string]string{"orgSlug": orgSlug, "planId": planID}
	switch kind {
	case "ComputeTarget", "ResourceProfile":
		var resource resourceApplyWire
		err = r.connectCall(
			ctx,
			orchestrationResourceService+"ApplyBindingResourcePlan",
			request,
			&resource,
		)
		return resource.applied(0, err)
	case "Prompt":
		var resource resourceApplyWire
		err = r.connectCall(
			ctx,
			orchestrationResourceService+"ApplyPromptPlan",
			request,
			&resource,
		)
		return resource.applied(0, err)
	case "WorkerTemplate":
		var response struct {
			Resource             resourceApplyWire `json:"resource"`
			WorkerSpecSnapshotID string            `json:"workerSpecSnapshotId"`
		}
		err = r.connectCall(
			ctx,
			orchestrationResourceService+"ApplyWorkerTemplatePlan",
			request,
			&response,
		)
		snapshotID, parseErr := parsePositiveID(
			"worker spec snapshot",
			response.WorkerSpecSnapshotID,
		)
		if err == nil {
			err = parseErr
		}
		return response.Resource.applied(snapshotID, err)
	case "Workflow":
		return r.applyWorkflowResource(ctx, request)
	default:
		return AppliedOrchestrationResource{}, fmt.Errorf(
			"unsupported orchestration resource kind %q",
			kind,
		)
	}
}

func (r *REST) PlanOrchestrationResource(
	ctx context.Context,
	orgSlug string,
	manifest any,
) (string, error) {
	return r.validateAndPlanResource(ctx, orgSlug, manifest)
}

type resourceApplyWire struct {
	Revision string `json:"revision"`
}

func (wire resourceApplyWire) applied(
	snapshotID int64,
	err error,
) (AppliedOrchestrationResource, error) {
	if err != nil {
		return AppliedOrchestrationResource{}, err
	}
	revision, err := parsePositiveID("resource revision", wire.Revision)
	if err != nil {
		return AppliedOrchestrationResource{}, err
	}
	return AppliedOrchestrationResource{
		Revision:             revision,
		WorkerSpecSnapshotID: snapshotID,
	}, nil
}
