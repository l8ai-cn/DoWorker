package sessionapi

import (
	"context"
	"fmt"
	"strings"

	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	workercreation "github.com/anthropics/agentsmesh/backend/internal/service/workercreation"
	specservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

type sessionWorkerSpecBody struct {
	OptionsRevision        string                             `json:"options_revision"`
	RuntimeImageID         int64                              `json:"runtime_image_id"`
	PlacementPolicy        string                             `json:"placement_policy"`
	ComputeTargetID        int64                              `json:"compute_target_id"`
	DeploymentMode         string                             `json:"deployment_mode"`
	ResourceProfileID      int64                              `json:"resource_profile_id"`
	ToolModelResourceIDs   map[string]int64                   `json:"tool_model_resource_ids"`
	ConfigDocumentBindings []specdomain.ConfigDocumentBinding `json:"config_document_bindings"`
}

type sessionWorkerPlanInput struct {
	WorkerSpec      *sessionWorkerSpecBody
	WorkerTypeSlug  string
	ModelResourceID *int64
	RepositoryID    *int64
	Branch          *string
	Alias           *string
	AgentfileLayer  *string
	AutomationLevel string
	Perpetual       bool
}

func (d *Deps) buildFreshWorkerPlan(
	ctx context.Context,
	orgID, userID int64,
	orgSlug string,
	input sessionWorkerPlanInput,
) (*workercreation.Draft, error) {
	if d.WorkerCreation == nil {
		return nil, specservice.ErrResolverUnavailable
	}
	if input.WorkerSpec == nil {
		return nil, invalidSessionWorkerPlan("worker_spec", "is required")
	}
	automation, err := workerPlanAutomation(input.AutomationLevel)
	if err != nil {
		return nil, err
	}
	draft, err := d.WorkerCreation.NewFreshPodDraft(
		ctx,
		specservice.Scope{OrgID: orgID, UserID: userID},
		workercreation.FreshPodDraftInput{
			OptionsRevision:        input.WorkerSpec.OptionsRevision,
			OrganizationSlug:       orgSlug,
			WorkerTypeSlug:         input.WorkerTypeSlug,
			ModelResourceID:        input.ModelResourceID,
			ToolModelResourceIDs:   input.WorkerSpec.ToolModelResourceIDs,
			ConfigDocumentBindings: input.WorkerSpec.ConfigDocumentBindings,
			Runtime: specservice.RuntimeSelection{
				RuntimeImageID:    input.WorkerSpec.RuntimeImageID,
				PlacementPolicy:   specdomain.PlacementPolicy(input.WorkerSpec.PlacementPolicy),
				ComputeTargetID:   input.WorkerSpec.ComputeTargetID,
				DeploymentMode:    specdomain.DeploymentMode(input.WorkerSpec.DeploymentMode),
				ResourceProfileID: input.WorkerSpec.ResourceProfileID,
			},
			RepositoryID:    input.RepositoryID,
			Branch:          workerPlanString(input.Branch),
			Alias:           workerPlanString(input.Alias),
			AgentfileLayer:  workerPlanString(input.AgentfileLayer),
			AutomationLevel: automation,
			Perpetual:       input.Perpetual,
		},
	)
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

func workerPlanString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func workerPlanAutomation(value string) (specdomain.AutomationLevel, error) {
	switch strings.TrimSpace(value) {
	case "":
		return "", invalidSessionWorkerPlan("automation_level", "is required")
	case string(specdomain.AutomationLevelInteractive):
		return specdomain.AutomationLevelInteractive, nil
	case string(specdomain.AutomationLevelAutoEdit):
		return specdomain.AutomationLevelAutoEdit, nil
	case string(specdomain.AutomationLevelAutonomous):
		return specdomain.AutomationLevelAutonomous, nil
	}
	return "", invalidSessionWorkerPlan(
		"automation_level",
		fmt.Sprintf("unsupported value %q", value),
	)
}

func invalidSessionWorkerPlan(field, reason string) error {
	return &specservice.InvalidDraftFieldError{Field: field, Reason: reason}
}
