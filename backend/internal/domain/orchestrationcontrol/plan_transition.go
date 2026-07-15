package orchestrationcontrol

import (
	"bytes"
	"encoding/json"
	"time"
)

func (plan Plan) Apply(
	consumedAt time.Time,
	actorID int64,
	resultResourceID int64,
	resultIdentity ResourceIdentity,
	resourceVersion int64,
	revision int64,
) (Plan, error) {
	if err := plan.validateTransition(consumedAt, actorID); err != nil {
		return Plan{}, err
	}
	if !consumedAt.Before(plan.ExpiresAt) {
		return Plan{}, ErrExpired
	}
	if err := plan.validateApplyResult(
		resultResourceID,
		resultIdentity,
		resourceVersion,
		revision,
	); err != nil {
		return Plan{}, err
	}

	next := plan
	next.Status = PlanStatusApplied
	next.ConsumedAt = timePointer(consumedAt)
	next.ConsumedByID = actorID
	next.ResultResourceID = resultResourceID
	next.ResultIdentity = &resultIdentity
	next.ResultResourceVersion = resourceVersion
	next.ResultRevision = revision
	next.ResultJSON = json.RawMessage(`{}`)
	if err := next.Validate(); err != nil {
		return Plan{}, err
	}
	return next, nil
}

func (plan Plan) Cancel(consumedAt time.Time, actorID int64) (Plan, error) {
	if err := plan.validateTransition(consumedAt, actorID); err != nil {
		return Plan{}, err
	}
	if !consumedAt.Before(plan.ExpiresAt) {
		return Plan{}, ErrExpired
	}
	next := plan
	next.Status = PlanStatusCancelled
	next.ConsumedAt = timePointer(consumedAt)
	next.ConsumedByID = actorID
	next.ResultResourceID = 0
	next.ResultIdentity = nil
	next.ResultResourceVersion = 0
	next.ResultRevision = 0
	next.ResultJSON = json.RawMessage(`{}`)
	if err := next.Validate(); err != nil {
		return Plan{}, err
	}
	return next, nil
}

func (plan Plan) Expire(consumedAt time.Time, actorID int64) (Plan, error) {
	if err := plan.validateTransition(consumedAt, actorID); err != nil {
		return Plan{}, err
	}
	if consumedAt.Before(plan.ExpiresAt) {
		return Plan{}, invalid("consumedAt", "must be at or after expiry")
	}
	next := plan
	next.Status = PlanStatusExpired
	next.ConsumedAt = timePointer(consumedAt)
	next.ConsumedByID = actorID
	next.ResultResourceID = 0
	next.ResultIdentity = nil
	next.ResultResourceVersion = 0
	next.ResultRevision = 0
	next.ResultJSON = json.RawMessage(`{}`)
	if err := next.Validate(); err != nil {
		return Plan{}, err
	}
	return next, nil
}

func (plan Plan) validateTransition(consumedAt time.Time, actorID int64) error {
	if err := plan.Validate(); err != nil {
		return err
	}
	if plan.Status != PlanStatusPending {
		return ErrConsumed
	}
	if actorID <= 0 || actorID != plan.ActorID {
		return invalid("consumedById", "must equal the plan actor")
	}
	if consumedAt.IsZero() || consumedAt.Before(plan.CreatedAt) {
		return invalid("consumedAt", "must not precede plan creation")
	}
	if _, err := consumedAt.MarshalJSON(); err != nil {
		return invalid("consumedAt", "must be encodable")
	}
	return nil
}

func (plan Plan) validateConsumption() error {
	switch plan.Status {
	case PlanStatusPending:
		if plan.ConsumedAt != nil || plan.ConsumedByID != 0 ||
			plan.ResultResourceID != 0 || plan.ResultIdentity != nil ||
			plan.ResultResourceVersion != 0 ||
			plan.ResultRevision != 0 || len(plan.ResultJSON) != 0 {
			return corrupt("plan.consumption", "pending plans must be unconsumed")
		}
		return nil
	case PlanStatusApplied:
		return plan.validateAppliedConsumption()
	case PlanStatusCancelled, PlanStatusExpired:
		return plan.validateNonResourceConsumption()
	default:
		return invalid("plan.status", "is unsupported")
	}
}

func (plan Plan) validateAppliedConsumption() error {
	if err := plan.validateConsumptionEnvelope(); err != nil {
		return err
	}
	if plan.ConsumedAt == nil || !plan.ConsumedAt.Before(plan.ExpiresAt) {
		return corrupt("plan.consumedAt", "applied plans must be consumed before expiry")
	}
	if plan.ResultIdentity == nil {
		return corrupt("plan.result", "must contain a resource identity")
	}
	if err := plan.validateApplyResult(
		plan.ResultResourceID,
		*plan.ResultIdentity,
		plan.ResultResourceVersion,
		plan.ResultRevision,
	); err != nil {
		return corrupt("plan.result", "must match the planned resource")
	}
	return nil
}

func (plan Plan) validateNonResourceConsumption() error {
	if err := plan.validateConsumptionEnvelope(); err != nil {
		return err
	}
	if plan.ResultResourceID != 0 || plan.ResultIdentity != nil ||
		plan.ResultResourceVersion != 0 ||
		plan.ResultRevision != 0 {
		return corrupt("plan.result", "non-applied plans must not contain a resource result")
	}
	if plan.Status == PlanStatusCancelled &&
		(plan.ConsumedAt == nil || !plan.ConsumedAt.Before(plan.ExpiresAt)) {
		return corrupt("plan.consumedAt", "cancelled plans must be consumed before expiry")
	}
	if plan.Status == PlanStatusExpired &&
		(plan.ConsumedAt == nil || plan.ConsumedAt.Before(plan.ExpiresAt)) {
		return corrupt("plan.consumedAt", "expired plans must be consumed at or after expiry")
	}
	return nil
}

func (plan Plan) validateConsumptionEnvelope() error {
	if plan.ConsumedAt == nil || plan.ConsumedByID != plan.ActorID ||
		plan.ConsumedAt.Before(plan.CreatedAt) {
		return corrupt("plan.consumption", "must be actor-bound and time ordered")
	}
	canonical, err := CanonicalJSONObject(plan.ResultJSON)
	if err != nil || !bytes.Equal(canonical, plan.ResultJSON) {
		return corrupt("plan.resultJson", "must be canonical object JSON")
	}
	if err := rejectRawSecretJSON(plan.ResultJSON); err != nil {
		return corrupt("plan.resultJson", "must not contain raw secrets")
	}
	return nil
}

func timePointer(value time.Time) *time.Time { return &value }
