package infra

import "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"

func (record orchestrationPlanRecord) validateShape() error {
	if err := record.validateBaseShape(); err != nil {
		return err
	}
	if record.ConsumedAt == nil {
		if record.ConsumedByID != nil || record.ConsumptionResult != nil ||
			record.ResultResourceID != nil || record.ResultResourceUID != nil ||
			record.ResultResourceVersion != nil || record.ResultRevision != nil ||
			len(record.ResultJSON) != 0 {
			return corruptRecord("pending plan consumption")
		}
		return nil
	}
	if record.ConsumedByID == nil || record.ConsumptionResult == nil ||
		len(record.ResultJSON) == 0 {
		return corruptRecord("plan consumption")
	}
	if *record.ConsumptionResult == string(orchestrationcontrol.PlanStatusApplied) {
		if record.ResultResourceID == nil || record.ResultResourceUID == nil ||
			record.ResultResourceVersion == nil || record.ResultRevision == nil {
			return corruptRecord("applied plan result")
		}
		return nil
	}
	if record.ResultResourceID != nil || record.ResultResourceUID != nil ||
		record.ResultResourceVersion != nil || record.ResultRevision != nil {
		return corruptRecord("non-applied plan result")
	}
	return nil
}

func (record orchestrationPlanRecord) validateBaseShape() error {
	switch orchestrationcontrol.PlanOperation(record.Operation) {
	case orchestrationcontrol.PlanOperationCreate:
		if record.TargetResourceID != nil || record.BaseHeadUID != nil ||
			record.BaseResourceVersion != nil {
			return corruptRecord("create plan base")
		}
	case orchestrationcontrol.PlanOperationUpdate:
		if record.TargetResourceID == nil || record.BaseHeadUID == nil ||
			record.BaseResourceVersion == nil {
			return corruptRecord("update plan base")
		}
	default:
		return corruptRecord("plan operation")
	}
	return nil
}
