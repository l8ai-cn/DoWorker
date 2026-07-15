package orchestrationcontrol

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanPinsReferencesComputesDiffAndPersistsImmutablePlan(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	service := fixture.service(t)

	result, err := service.Plan(context.Background(), PlanRequest{
		Scope:  fixture.scope,
		Source: ResourceSource{Format: SourceFormatJSON, Content: []byte(testResourceJSON)},
	})
	require.NoError(t, err)
	require.NotNil(t, result.Plan)
	require.Empty(t, result.Issues)
	require.Equal(t, 1, fixture.repository.createPlanCalls)
	stored := fixture.repository.createdPlans[0]

	assert.Equal(t, fixture.planID, stored.ID)
	assert.Equal(t, fixture.now, stored.CreatedAt)
	assert.Equal(t, fixture.now.Add(15*time.Minute), stored.ExpiresAt)
	assert.Equal(t, orchestrationcontrol.PlanStatusPending, stored.Status)
	assert.Len(t, stored.ResolvedReferences, 2)
	assert.Equal(t, "ModelBinding", stored.ResolvedReferences[0].Kind)
	assert.Equal(t, "Prompt", stored.ResolvedReferences[1].Kind)
	assert.NotEmpty(t, stored.SemanticChanges)
	assert.Equal(t, "WorkerSpec", stored.ArtifactKind)
	assert.JSONEq(t, `{"snapshotVersion":2}`, string(stored.ArtifactJSON))
	assert.NoError(t, stored.Validate())
	assert.Equal(t, stored, *result.Plan)
}

func TestPlanRejectsUnauthorizedReferenceWithoutPersistence(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	fixture.references.errByKind["Prompt"] = ErrForbidden
	service := fixture.service(t)

	_, err := service.Plan(context.Background(), PlanRequest{
		Scope:  fixture.scope,
		Source: ResourceSource{Format: SourceFormatJSON, Content: []byte(testResourceJSON)},
	})
	assert.ErrorIs(t, err, ErrForbidden)
	assert.Zero(t, fixture.repository.createPlanCalls)
	assert.Equal(t, []string{"ModelBinding", "Prompt"}, fixture.references.kinds)
}

func TestPlanReturnsStaleOptionsIssueWithoutPersistence(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	fixture.planner.planErr = ErrStaleOptions
	service := fixture.service(t)

	result, err := service.Plan(context.Background(), PlanRequest{
		Scope:  fixture.scope,
		Source: ResourceSource{Format: SourceFormatJSON, Content: []byte(testResourceJSON)},
	})
	require.NoError(t, err)
	require.Nil(t, result.Plan)
	require.Len(t, result.Issues, 1)
	assert.Equal(t, "stale-options", result.Issues[0].Code)
	assert.Zero(t, fixture.repository.createPlanCalls)
}

func TestPlanRejectsHeadRaceBeforePersistence(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	head := orchestrationServiceHead()
	fixture.repository.resourceSequence = []resourceRead{
		{head: head},
		{head: func() orchestrationcontrol.ResourceHead {
			changed := head
			changed.ResourceVersion++
			changed.UpdatedAt = changed.UpdatedAt.Add(time.Second)
			return changed
		}()},
	}
	fixture.repository.revisions[head.Revision] = orchestrationServiceRevision(t, head)
	service := fixture.service(t)

	_, err := service.Plan(context.Background(), PlanRequest{
		Scope:  fixture.scope,
		Source: ResourceSource{Format: SourceFormatJSON, Content: []byte(testResourceJSON)},
	})
	assert.ErrorIs(t, err, orchestrationcontrol.ErrStale)
	assert.Zero(t, fixture.repository.createPlanCalls)
}

func TestPlanSortsPlannerIssuesDeterministically(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	fixture.planner.output.Issues = []orchestrationcontrol.PlanIssue{
		{Severity: orchestrationcontrol.PlanIssueWarning, Path: "/spec/promptRef", Code: "z-warning", Message: "Z"},
		{Severity: orchestrationcontrol.PlanIssueWarning, Path: "/spec/modelRef", Code: "a-warning", Message: "A"},
	}
	service := fixture.service(t)

	result, err := service.Plan(context.Background(), PlanRequest{
		Scope:  fixture.scope,
		Source: ResourceSource{Format: SourceFormatJSON, Content: []byte(testResourceJSON)},
	})
	require.NoError(t, err)
	require.NotNil(t, result.Plan)
	require.Len(t, result.Plan.Issues, 2)
	assert.Equal(t, "/spec/modelRef", result.Plan.Issues[0].Path)
	assert.Equal(t, "/spec/promptRef", result.Plan.Issues[1].Path)
}

func TestPlanRejectsSecretBearingArtifact(t *testing.T) {
	fixture := newOrchestrationServiceFixture(t)
	fixture.planner.output.ArtifactJSON = json.RawMessage(
		`{"password":{"value":"plaintext"}}`,
	)
	service := fixture.service(t)

	_, err := service.Plan(context.Background(), PlanRequest{
		Scope:  fixture.scope,
		Source: ResourceSource{Format: SourceFormatJSON, Content: []byte(testResourceJSON)},
	})
	assert.ErrorIs(t, err, orchestrationcontrol.ErrInvalid)
	assert.Zero(t, fixture.repository.createPlanCalls)
}
