package orchestrationcontrol

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlanApplyReturnsNewTerminalStateAndRejectsReplay(t *testing.T) {
	plan := validUpdatePlan(t)
	consumedAt := plan.CreatedAt.Add(time.Minute)

	applied, err := plan.Apply(
		consumedAt,
		plan.ActorID,
		plan.TargetResourceID,
		validIdentity(),
		plan.BaseResourceVersion+1,
		3,
	)
	require.NoError(t, err)
	require.Equal(t, PlanStatusPending, plan.Status)
	require.Nil(t, plan.ConsumedAt)
	require.Equal(t, PlanStatusApplied, applied.Status)
	require.Equal(t, consumedAt, *applied.ConsumedAt)
	require.Equal(t, plan.ActorID, applied.ConsumedByID)
	require.Equal(t, plan.TargetResourceID, applied.ResultResourceID)
	require.Equal(t, validIdentity(), *applied.ResultIdentity)
	require.Equal(t, int64(9), applied.ResultResourceVersion)
	require.Equal(t, int64(3), applied.ResultRevision)
	require.JSONEq(t, `{}`, string(applied.ResultJSON))
	require.NoError(t, applied.Validate())

	_, err = applied.Cancel(consumedAt.Add(time.Second), plan.ActorID)
	require.ErrorIs(t, err, ErrConsumed)
	_, err = applied.Expire(plan.ExpiresAt, plan.ActorID)
	require.ErrorIs(t, err, ErrConsumed)
	_, err = applied.Apply(
		consumedAt.Add(time.Second),
		plan.ActorID,
		plan.TargetResourceID,
		validIdentity(),
		10,
		4,
	)
	require.ErrorIs(t, err, ErrConsumed)
}

func TestPlanApplyRequiresActorTargetAndFreshExpiry(t *testing.T) {
	plan := validUpdatePlan(t)

	_, err := plan.Apply(
		plan.CreatedAt.Add(time.Minute),
		99,
		plan.TargetResourceID,
		validIdentity(),
		9,
		3,
	)
	require.ErrorIs(t, err, ErrInvalid)

	wrongIdentity := validIdentity()
	wrongIdentity.Name = "worker-two"
	_, err = plan.Apply(
		plan.CreatedAt.Add(time.Minute),
		plan.ActorID,
		plan.TargetResourceID,
		wrongIdentity,
		9,
		3,
	)
	require.ErrorIs(t, err, ErrInvalid)

	newUID := validIdentity()
	newUID.UID = "55555555-5555-4555-8555-555555555555"
	_, err = plan.Apply(
		plan.CreatedAt.Add(time.Minute),
		plan.ActorID,
		plan.TargetResourceID,
		newUID,
		9,
		3,
	)
	require.ErrorIs(t, err, ErrInvalid)

	_, err = plan.Apply(
		plan.ExpiresAt,
		plan.ActorID,
		plan.TargetResourceID,
		validIdentity(),
		9,
		3,
	)
	require.ErrorIs(t, err, ErrExpired)

	_, err = plan.Cancel(plan.ExpiresAt, plan.ActorID)
	require.ErrorIs(t, err, ErrExpired)
}

func TestPlanCancelAndExpireHaveNoResourceResult(t *testing.T) {
	plan := validCreatePlan(t)
	cancelledAt := plan.CreatedAt.Add(time.Minute)
	cancelled, err := plan.Cancel(cancelledAt, plan.ActorID)
	require.NoError(t, err)
	require.Equal(t, PlanStatusCancelled, cancelled.Status)
	require.Nil(t, cancelled.ResultIdentity)
	require.Zero(t, cancelled.ResultResourceID)
	require.Zero(t, cancelled.ResultResourceVersion)
	require.Zero(t, cancelled.ResultRevision)
	require.JSONEq(t, `{}`, string(cancelled.ResultJSON))
	require.NoError(t, cancelled.Validate())

	expired, err := plan.Expire(plan.ExpiresAt, plan.ActorID)
	require.NoError(t, err)
	require.Equal(t, PlanStatusExpired, expired.Status)
	require.Nil(t, expired.ResultIdentity)
	require.NoError(t, expired.Validate())

	_, err = plan.Expire(plan.ExpiresAt.Add(-time.Nanosecond), plan.ActorID)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestPlanValidationRejectsInconsistentConsumptionState(t *testing.T) {
	plan := validCreatePlan(t)
	consumedAt := plan.CreatedAt.Add(time.Minute)
	plan.Status = PlanStatusApplied
	plan.ConsumedAt = &consumedAt
	plan.ConsumedByID = plan.ActorID
	plan.ResultIdentity = nil
	plan.ResultResourceID = 0
	plan.ResultResourceVersion = 0
	plan.ResultRevision = 0
	plan.ResultJSON = json.RawMessage(`{}`)

	require.ErrorIs(t, plan.Validate(), ErrCorrupt)

	plan.Status = PlanStatusCancelled
	plan.ResultIdentity = &ResourceIdentity{
		ResourceTarget: validTarget(),
		UID:            testTargetID,
	}
	plan.ResultResourceID = 101
	require.ErrorIs(t, plan.Validate(), ErrCorrupt)

	plan.Status = PlanStatusExpired
	plan.ResultIdentity = nil
	plan.ResultResourceID = 0
	plan.ConsumedAt = &consumedAt
	require.ErrorIs(t, plan.Validate(), ErrCorrupt)
}

func TestPlanApplyRequiresPersistedResultResource(t *testing.T) {
	plan := validCreatePlan(t)
	_, err := plan.Apply(
		plan.CreatedAt.Add(time.Minute),
		plan.ActorID,
		0,
		validIdentity(),
		1,
		1,
	)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestPlanApplyRequiresResultToMatchPlannedResource(t *testing.T) {
	update := validUpdatePlan(t)
	_, err := update.Apply(
		update.CreatedAt.Add(time.Minute),
		update.ActorID,
		update.TargetResourceID+1,
		validIdentity(),
		update.BaseResourceVersion+1,
		3,
	)
	require.ErrorIs(t, err, ErrInvalid)

	create := validCreatePlan(t)
	_, err = create.Apply(
		create.CreatedAt.Add(time.Minute),
		create.ActorID,
		101,
		validIdentity(),
		2,
		1,
	)
	require.ErrorIs(t, err, ErrInvalid)
	_, err = create.Apply(
		create.CreatedAt.Add(time.Minute),
		create.ActorID,
		101,
		validIdentity(),
		1,
		2,
	)
	require.ErrorIs(t, err, ErrInvalid)
}

func TestPlanValidationRejectsContradictoryAppliedResult(t *testing.T) {
	plan := validUpdatePlan(t)
	applied, err := plan.Apply(
		plan.CreatedAt.Add(time.Minute),
		plan.ActorID,
		plan.TargetResourceID,
		validIdentity(),
		plan.BaseResourceVersion+1,
		3,
	)
	require.NoError(t, err)

	applied.ResultResourceID++
	require.ErrorIs(t, applied.Validate(), ErrCorrupt)
	applied.ResultResourceID = plan.TargetResourceID
	applied.ResultRevision = applied.ResultResourceVersion + 1
	require.ErrorIs(t, applied.Validate(), ErrCorrupt)
}

func TestPlanIssueAndDiffSecretErrorsDoNotEchoRawValues(t *testing.T) {
	const raw = "password=phase2a-secret"
	plan := validCreatePlan(t)
	plan.Issues[0].Message = raw
	err := plan.Validate()
	require.Error(t, err)
	require.NotContains(t, err.Error(), raw)

	plan = validCreatePlan(t)
	plan.SemanticChanges[0].After = ChangeValue{
		RedactedJSON: json.RawMessage(`{"password":"phase2a-secret"}`),
	}
	err = plan.Validate()
	require.Error(t, err)
	require.NotContains(t, err.Error(), "phase2a-secret")
	require.NotContains(t, strings.ToLower(err.Error()), "password=phase2a-secret")
}

func TestSecretGuardAllowsNonSecretTokenCounters(t *testing.T) {
	require.NoError(t, rejectRawSecretJSON(json.RawMessage(
		`{"maxTokens":4096,"inputTokens":128,"outputTokens":64}`,
	)))
}
