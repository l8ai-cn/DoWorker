package infra

import (
	"encoding/json"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
)

type orchestrationPlanRecord struct {
	ID                    string          `gorm:"column:id;primaryKey"`
	OrganizationID        int64           `gorm:"column:organization_id"`
	ActorID               int64           `gorm:"column:actor_id"`
	TargetResourceID      *int64          `gorm:"column:target_resource_id"`
	TargetAPIVersion      string          `gorm:"column:target_api_version"`
	TargetKind            string          `gorm:"column:target_kind"`
	TargetNamespace       string          `gorm:"column:target_namespace"`
	TargetName            string          `gorm:"column:target_name"`
	Operation             string          `gorm:"column:operation"`
	BaseHeadUID           *string         `gorm:"column:base_head_uid"`
	BaseResourceVersion   *int64          `gorm:"column:base_resource_version"`
	DraftHash             string          `gorm:"column:draft_hash"`
	PlanHash              string          `gorm:"column:plan_hash"`
	CanonicalManifest     json.RawMessage `gorm:"column:canonical_manifest;type:jsonb"`
	ResolvedReferences    json.RawMessage `gorm:"column:resolved_refs;type:jsonb"`
	SemanticDiff          json.RawMessage `gorm:"column:semantic_diff;type:jsonb"`
	Issues                json.RawMessage `gorm:"column:issues;type:jsonb"`
	ArtifactKind          string          `gorm:"column:artifact_kind"`
	ArtifactJSON          json.RawMessage `gorm:"column:artifact_json;type:jsonb"`
	ArtifactDigest        string          `gorm:"column:artifact_digest"`
	OptionsRevision       string          `gorm:"column:options_revision"`
	CreatedAt             time.Time       `gorm:"column:created_at"`
	ExpiresAt             time.Time       `gorm:"column:expires_at"`
	ConsumedAt            *time.Time      `gorm:"column:consumed_at"`
	ConsumedByID          *int64          `gorm:"column:consumed_by_id"`
	ConsumptionResult     *string         `gorm:"column:consumption_result"`
	ResultResourceID      *int64          `gorm:"column:result_resource_id"`
	ResultResourceUID     *string         `gorm:"column:result_resource_uid"`
	ResultResourceVersion *int64          `gorm:"column:result_resource_version"`
	ResultRevision        *int64          `gorm:"column:result_revision"`
	ResultJSON            json.RawMessage `gorm:"column:result_json;type:jsonb"`
}

func (orchestrationPlanRecord) TableName() string {
	return "orchestration_resource_plans"
}

func orchestrationPlanRecordFromDomain(
	plan orchestrationcontrol.Plan,
) (orchestrationPlanRecord, error) {
	if err := plan.Validate(); err != nil {
		return orchestrationPlanRecord{}, err
	}
	resolved, err := orchestrationcontrol.CanonicalJSONArray(plan.ResolvedReferences)
	if err != nil {
		return orchestrationPlanRecord{}, err
	}
	diff, err := orchestrationcontrol.CanonicalJSONArray(plan.SemanticChanges)
	if err != nil {
		return orchestrationPlanRecord{}, err
	}
	issues, err := orchestrationcontrol.CanonicalJSONArray(plan.Issues)
	if err != nil {
		return orchestrationPlanRecord{}, err
	}
	record := orchestrationPlanRecord{
		ID: plan.ID, OrganizationID: plan.Scope.OrganizationID,
		ActorID: plan.ActorID, TargetAPIVersion: plan.Target.APIVersion,
		TargetKind: plan.Target.Kind, TargetNamespace: plan.Target.Namespace.String(),
		TargetName: plan.Target.Name.String(), Operation: string(plan.Operation),
		DraftHash: plan.DraftHash, PlanHash: plan.PlanHash,
		CanonicalManifest: plan.CanonicalManifest, ResolvedReferences: resolved,
		SemanticDiff: diff, Issues: issues, ArtifactKind: plan.ArtifactKind,
		ArtifactJSON: plan.ArtifactJSON, ArtifactDigest: plan.ArtifactDigest,
		OptionsRevision: plan.OptionsRevision, CreatedAt: plan.CreatedAt,
		ExpiresAt: plan.ExpiresAt,
	}
	record.setBase(plan)
	record.setConsumption(plan)
	return record, nil
}

func (record *orchestrationPlanRecord) setBase(plan orchestrationcontrol.Plan) {
	if plan.TargetResourceID > 0 {
		record.TargetResourceID = int64Pointer(plan.TargetResourceID)
		record.BaseHeadUID = stringPointer(plan.BaseUID)
		record.BaseResourceVersion = int64Pointer(plan.BaseResourceVersion)
	}
}

func (record *orchestrationPlanRecord) setConsumption(plan orchestrationcontrol.Plan) {
	if plan.ConsumedAt == nil {
		return
	}
	record.ConsumedAt = plan.ConsumedAt
	record.ConsumedByID = int64Pointer(plan.ConsumedByID)
	status := string(plan.Status)
	record.ConsumptionResult = &status
	record.ResultJSON = plan.ResultJSON
	if plan.Status == orchestrationcontrol.PlanStatusApplied {
		record.ResultResourceID = int64Pointer(plan.ResultResourceID)
		record.ResultResourceUID = stringPointer(plan.ResultIdentity.UID)
		record.ResultResourceVersion = int64Pointer(plan.ResultResourceVersion)
		record.ResultRevision = int64Pointer(plan.ResultRevision)
	}
}
