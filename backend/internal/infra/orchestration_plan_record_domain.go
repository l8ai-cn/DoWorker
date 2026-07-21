package infra

import (
	"encoding/json"

	"github.com/l8ai-cn/agentcloud/backend/internal/domain/orchestrationcontrol"
)

func (record orchestrationPlanRecord) domain(
	scope orchestrationcontrol.Scope,
) (orchestrationcontrol.Plan, error) {
	if err := record.validateShape(); err != nil {
		return orchestrationcontrol.Plan{}, err
	}
	manifest, err := orchestrationcontrol.CanonicalJSONObject(record.CanonicalManifest)
	if err != nil {
		return orchestrationcontrol.Plan{}, corruptRecord("plan manifest")
	}
	artifact, err := orchestrationcontrol.CanonicalJSONObject(record.ArtifactJSON)
	if err != nil {
		return orchestrationcontrol.Plan{}, corruptRecord("plan artifact")
	}
	var resolved []orchestrationcontrol.ResolvedReference
	if err := decodeCanonicalArray(record.ResolvedReferences, &resolved); err != nil {
		return orchestrationcontrol.Plan{}, corruptRecord("plan references")
	}
	var changes []orchestrationcontrol.SemanticChange
	if err := decodeCanonicalArray(record.SemanticDiff, &changes); err != nil {
		return orchestrationcontrol.Plan{}, corruptRecord("plan diff")
	}
	var issues []orchestrationcontrol.PlanIssue
	if err := decodeCanonicalArray(record.Issues, &issues); err != nil {
		return orchestrationcontrol.Plan{}, corruptRecord("plan issues")
	}
	plan := orchestrationcontrol.Plan{
		ID: record.ID, Scope: scope, ActorID: record.ActorID,
		Operation: orchestrationcontrol.PlanOperation(record.Operation),
		Target: orchestrationcontrol.ResourceTarget{
			TypeMeta:  structTypeMeta(record.TargetAPIVersion, record.TargetKind),
			Namespace: stringSlug(record.TargetNamespace),
			Name:      stringSlug(record.TargetName),
		},
		DraftHash: record.DraftHash, PlanHash: record.PlanHash,
		CanonicalManifest: manifest, ResolvedReferences: resolved,
		SemanticChanges: changes, Issues: issues, ArtifactKind: record.ArtifactKind,
		ArtifactJSON: artifact, ArtifactDigest: record.ArtifactDigest,
		OptionsRevision: record.OptionsRevision, CreatedAt: record.CreatedAt.UTC(),
		ExpiresAt: record.ExpiresAt.UTC(), Status: orchestrationcontrol.PlanStatusPending,
	}
	record.applyBase(&plan)
	if err := record.applyConsumption(&plan); err != nil {
		return orchestrationcontrol.Plan{}, err
	}
	if err := plan.Validate(); err != nil {
		return orchestrationcontrol.Plan{}, corruptRecord("resource plan")
	}
	return plan, nil
}

func (record orchestrationPlanRecord) applyBase(plan *orchestrationcontrol.Plan) {
	if record.TargetResourceID == nil {
		return
	}
	plan.TargetResourceID = *record.TargetResourceID
	if record.BaseHeadUID != nil {
		plan.BaseUID = *record.BaseHeadUID
	}
	if record.BaseResourceVersion != nil {
		plan.BaseResourceVersion = *record.BaseResourceVersion
	}
}

func (record orchestrationPlanRecord) applyConsumption(
	plan *orchestrationcontrol.Plan,
) error {
	if record.ConsumedAt == nil {
		return nil
	}
	if record.ConsumedByID == nil || record.ConsumptionResult == nil {
		return corruptRecord("plan consumption")
	}
	resultJSON, err := orchestrationcontrol.CanonicalJSONObject(record.ResultJSON)
	if err != nil {
		return corruptRecord("plan result")
	}
	plan.Status = orchestrationcontrol.PlanStatus(*record.ConsumptionResult)
	consumedAt := record.ConsumedAt.UTC()
	plan.ConsumedAt = &consumedAt
	plan.ConsumedByID = *record.ConsumedByID
	plan.ResultJSON = resultJSON
	if plan.Status != orchestrationcontrol.PlanStatusApplied {
		return nil
	}
	if record.ResultResourceID == nil || record.ResultResourceUID == nil ||
		record.ResultResourceVersion == nil || record.ResultRevision == nil {
		return corruptRecord("plan result")
	}
	plan.ResultResourceID = *record.ResultResourceID
	plan.ResultIdentity = &orchestrationcontrol.ResourceIdentity{
		ResourceTarget: plan.Target,
		UID:            *record.ResultResourceUID,
	}
	plan.ResultResourceVersion = *record.ResultResourceVersion
	plan.ResultRevision = *record.ResultRevision
	return nil
}

func decodeCanonicalArray(raw json.RawMessage, target any) error {
	canonical, err := orchestrationcontrol.CanonicalJSONArray(raw)
	if err != nil {
		return err
	}
	return decodeStrictJSON(canonical, target)
}

func int64Pointer(value int64) *int64 { return &value }

func stringPointer(value string) *string { return &value }
