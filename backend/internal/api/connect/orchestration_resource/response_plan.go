package orchestrationresourceconnect

import (
	control "github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
	"github.com/l8ai-cn/agentcloud/backend/pkg/protoconv"
	resourcev1 "github.com/l8ai-cn/agentcloud/proto/gen/go/orchestration_resource/v1"
)

func planToProto(plan control.Plan) (*resourcev1.ResourcePlan, error) {
	operation, err := operationToProto(plan.Operation)
	if err != nil || operation == resourcev1.ResourceOperation_RESOURCE_OPERATION_UNSPECIFIED {
		return nil, control.ErrCorrupt
	}
	status, err := planStatusToProto(plan.Status)
	if err != nil {
		return nil, err
	}
	references := make([]*resourcev1.ResolvedReference, len(plan.ResolvedReferences))
	for index := range plan.ResolvedReferences {
		references[index] = resolvedReferenceToProto(plan.ResolvedReferences[index])
	}
	changes, err := semanticChangesToProto(plan.SemanticChanges)
	if err != nil {
		return nil, err
	}
	issues, err := issuesToProto(plan.Issues)
	if err != nil {
		return nil, err
	}
	result := &resourcev1.ResourcePlan{
		PlanId:              plan.ID,
		Operation:           operation,
		Target:              targetToProto(plan.Target),
		BaseResourceVersion: plan.BaseResourceVersion,
		DraftHash:           plan.DraftHash,
		PlanHash:            plan.PlanHash,
		ArtifactDigest:      plan.ArtifactDigest,
		ResolvedReferences:  references,
		SemanticDiff:        changes,
		Issues:              issues,
		ArtifactKind:        plan.ArtifactKind,
		OptionsRevision:     plan.OptionsRevision,
		CreatedAt:           protoconv.RFC3339(plan.CreatedAt),
		ExpiresAt:           protoconv.RFC3339(plan.ExpiresAt),
		Status:              status,
	}
	if plan.BaseUID != "" {
		result.Base = identityToProto(control.ResourceIdentity{
			ResourceTarget: plan.Target,
			UID:            plan.BaseUID,
		})
	}
	return result, nil
}

func resolvedReferenceToProto(
	reference control.ResolvedReference,
) *resourcev1.ResolvedReference {
	return &resourcev1.ResolvedReference{
		TypeMeta:  typeMetaToProto(reference.TypeMeta),
		Namespace: reference.Namespace.String(),
		Name:      reference.Name.String(),
		Uid:       reference.UID,
		Revision:  reference.Revision,
		Digest:    reference.Digest,
	}
}

func issuesToProto(issues []control.PlanIssue) ([]*resourcev1.PlanIssue, error) {
	result := make([]*resourcev1.PlanIssue, len(issues))
	for index := range issues {
		if err := issues[index].Validate(); err != nil {
			return nil, control.ErrCorrupt
		}
		severity, err := issueSeverityToProto(issues[index].Severity)
		if err != nil {
			return nil, err
		}
		result[index] = &resourcev1.PlanIssue{
			Severity: severity,
			Path:     issues[index].Path,
			Code:     issues[index].Code,
			Message:  issues[index].Message,
		}
	}
	return result, nil
}

func semanticChangesToProto(
	changes []control.SemanticChange,
) ([]*resourcev1.SemanticChange, error) {
	result := make([]*resourcev1.SemanticChange, len(changes))
	for index := range changes {
		if err := changes[index].Validate(); err != nil {
			return nil, control.ErrCorrupt
		}
		operation, err := changeOperationToProto(changes[index].Operation)
		if err != nil {
			return nil, err
		}
		result[index] = &resourcev1.SemanticChange{
			Operation: operation,
			Path:      changes[index].Path,
			Before:    changeValueToProto(changes[index].Before),
			After:     changeValueToProto(changes[index].After),
		}
	}
	return result, nil
}

func changeValueToProto(value control.ChangeValue) *resourcev1.ChangeValue {
	switch {
	case value.Digest != "":
		return &resourcev1.ChangeValue{
			Value: &resourcev1.ChangeValue_Digest{Digest: value.Digest},
		}
	case len(value.RedactedJSON) != 0:
		return &resourcev1.ChangeValue{
			Value: &resourcev1.ChangeValue_RedactedJson{
				RedactedJson: append([]byte(nil), value.RedactedJSON...),
			},
		}
	default:
		return nil
	}
}

func operationToProto(
	operation control.PlanOperation,
) (resourcev1.ResourceOperation, error) {
	switch operation {
	case "":
		return resourcev1.ResourceOperation_RESOURCE_OPERATION_UNSPECIFIED, nil
	case control.PlanOperationCreate:
		return resourcev1.ResourceOperation_RESOURCE_OPERATION_CREATE, nil
	case control.PlanOperationUpdate:
		return resourcev1.ResourceOperation_RESOURCE_OPERATION_UPDATE, nil
	default:
		return resourcev1.ResourceOperation_RESOURCE_OPERATION_UNSPECIFIED, control.ErrCorrupt
	}
}

func issueSeverityToProto(
	severity control.PlanIssueSeverity,
) (resourcev1.IssueSeverity, error) {
	switch severity {
	case control.PlanIssueBlocking:
		return resourcev1.IssueSeverity_ISSUE_SEVERITY_BLOCKING, nil
	case control.PlanIssueWarning:
		return resourcev1.IssueSeverity_ISSUE_SEVERITY_WARNING, nil
	default:
		return resourcev1.IssueSeverity_ISSUE_SEVERITY_UNSPECIFIED, control.ErrCorrupt
	}
}

func changeOperationToProto(
	operation control.SemanticChangeOperation,
) (resourcev1.SemanticChangeOperation, error) {
	switch operation {
	case control.SemanticChangeAdd:
		return resourcev1.SemanticChangeOperation_SEMANTIC_CHANGE_OPERATION_ADD, nil
	case control.SemanticChangeRemove:
		return resourcev1.SemanticChangeOperation_SEMANTIC_CHANGE_OPERATION_REMOVE, nil
	case control.SemanticChangeReplace:
		return resourcev1.SemanticChangeOperation_SEMANTIC_CHANGE_OPERATION_REPLACE, nil
	default:
		return resourcev1.SemanticChangeOperation_SEMANTIC_CHANGE_OPERATION_UNSPECIFIED, control.ErrCorrupt
	}
}

func planStatusToProto(status control.PlanStatus) (resourcev1.PlanStatus, error) {
	switch status {
	case control.PlanStatusPending:
		return resourcev1.PlanStatus_PLAN_STATUS_PENDING, nil
	case control.PlanStatusApplied:
		return resourcev1.PlanStatus_PLAN_STATUS_APPLIED, nil
	case control.PlanStatusCancelled:
		return resourcev1.PlanStatus_PLAN_STATUS_CANCELLED, nil
	case control.PlanStatusExpired:
		return resourcev1.PlanStatus_PLAN_STATUS_EXPIRED, nil
	default:
		return resourcev1.PlanStatus_PLAN_STATUS_UNSPECIFIED, control.ErrCorrupt
	}
}
