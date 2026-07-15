package infra

import (
	"context"
	"encoding/json"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	"github.com/anthropics/agentsmesh/backend/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGoalLoopApplyTransactionCreatesExactPinnedDraftAndReplays(
	t *testing.T,
) {
	db, repo := goalLoopApplyPostgresRepository(t)
	plan := orchestrationGoalLoopApplyPlan(t, 100)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))
	builder := goalLoopMutationBuilder(t, 100, 100)

	first, err := repo.RunGoalLoopApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	require.NoError(t, err)
	replayed, err := repo.RunGoalLoopApplyTransaction(
		context.Background(), plan.Scope, plan.ID, builder,
	)
	require.NoError(t, err)

	assert.Equal(t, first, replayed)
	assert.Equal(t, int64(100), first.WorkerSpecSnapshotID)
	assert.Equal(t, first.Head.Revision, first.ResourceRevision)
	var row goalLoopApplyRow
	require.NoError(t, db.Table("goal_loops").Take(&row).Error)
	assert.Equal(t, int64(42), row.OrganizationID)
	assert.Equal(t, int64(7), row.CreatedByID)
	assert.Equal(t, "Checkout Recovery", row.Name)
	assert.Equal(t, "checkout-recovery", row.Slug)
	require.NotNil(t, row.Description)
	assert.Equal(t, "Restore checkout reliability", *row.Description)
	assert.Equal(t, int64(100), row.WorkerSpecSnapshotID)
	assert.Equal(t, "Fix checkout", row.Objective)
	assert.JSONEq(t, `["Tests pass","Evidence recorded"]`,
		string(row.AcceptanceCriteria))
	assert.Equal(t, "go test ./...", row.VerificationCommand)
	assert.Equal(t, "draft", row.Status)
	assert.Equal(t, 100, row.MaxIterations)
	require.NotNil(t, row.TokenBudget)
	assert.Equal(t, int64(200000), *row.TokenBudget)
	assert.Equal(t, 1440, row.TimeoutMinutes)
	assert.Equal(t, 20, row.NoProgressLimit)
	assert.Equal(t, 20, row.SameErrorLimit)
	assert.Equal(t, "fail", row.EscalationPolicy)
	assert.Equal(t, first.Head.ID, row.OrchestrationResourceID)
	assert.Equal(t, first.Head.Revision, row.OrchestrationResourceRevision)
	assert.Nil(t, row.PodKey)
	assert.Nil(t, row.AutopilotControllerKey)
	assert.Zero(t, row.CurrentIteration)

	revision, err := repo.GetRevision(
		context.Background(),
		plan.Scope,
		first.Head.ID,
		first.Head.Revision,
	)
	require.NoError(t, err)
	assert.Equal(t, row.WorkerSpecSnapshotID, revision.WorkerSpecSnapshotID)
}

func TestGoalLoopApplyTransactionRejectsSnapshotMismatch(t *testing.T) {
	_, repo := goalLoopApplyPostgresRepository(t)
	plan := orchestrationGoalLoopApplyPlan(t, 100)
	require.NoError(t, repo.CreatePlan(context.Background(), plan))

	_, err := repo.RunGoalLoopApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		goalLoopMutationBuilder(t, 100, 200),
	)

	assert.ErrorIs(t, err, control.ErrInvalid)
}

func TestGoalLoopApplyTransactionRejectsCrossOrgAndMissingSnapshots(
	t *testing.T,
) {
	for _, test := range []struct {
		name       string
		snapshotID int64
	}{
		{name: "cross organization", snapshotID: 200},
		{name: "missing", snapshotID: 999},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, repo := goalLoopApplyPostgresRepository(t)
			plan := orchestrationGoalLoopApplyPlan(t, test.snapshotID)
			require.NoError(t, repo.CreatePlan(context.Background(), plan))

			_, err := repo.RunGoalLoopApplyTransaction(
				context.Background(),
				plan.Scope,
				plan.ID,
				goalLoopMutationBuilder(t, test.snapshotID, test.snapshotID),
			)

			require.Error(t, err)
			for _, table := range []string{
				"goal_loops",
				"orchestration_resources",
				"orchestration_resource_revisions",
			} {
				var count int64
				require.NoError(t, db.Table(table).Count(&count).Error)
				assert.Zero(t, count, table)
			}
			stored, getErr := repo.GetPlan(
				context.Background(),
				plan.Scope,
				plan.ID,
			)
			require.NoError(t, getErr)
			assert.Equal(t, control.PlanStatusPending, stored.Status)
		})
	}
}

type goalLoopApplyRow struct {
	OrganizationID                int64
	CreatedByID                   int64
	Name                          string
	Slug                          string
	Description                   *string
	WorkerSpecSnapshotID          int64
	Objective                     string
	AcceptanceCriteria            json.RawMessage
	VerificationCommand           string
	Status                        string
	PodKey                        *string
	AutopilotControllerKey        *string
	MaxIterations                 int
	CurrentIteration              int
	TokenBudget                   *int64
	TimeoutMinutes                int
	NoProgressLimit               int
	SameErrorLimit                int
	EscalationPolicy              string
	OrchestrationResourceID       int64
	OrchestrationResourceRevision int64
}

func goalLoopApplyPostgresRepository(
	t *testing.T,
) (*gorm.DB, *orchestrationResourceRepo) {
	t.Helper()
	db, repo := orchestrationPostgresRepository(t)
	applyOrchestrationDomainLinkFixtures(t, db)
	require.NoError(t, db.Exec(`
INSERT INTO worker_spec_snapshots
	(id, organization_id, version, spec_json, summary_json)
VALUES (200, 99, 1, '{}', '{}')`).Error)
	for _, name := range []string{
		"000202_add_goal_loops.up.sql",
		"000217_orchestration_goal_loop_link.up.sql",
	} {
		content, err := migrations.FS.ReadFile(name)
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(content)).Error)
	}
	return db, repo
}

func goalLoopMutationBuilder(
	t *testing.T,
	revisionSnapshotID, projectionSnapshotID int64,
) workerplanner.GoalLoopApplyBuilder {
	t.Helper()
	return func(state controlservice.LockedApplyState) (
		workerplanner.GoalLoopApplyMutation,
		error,
	) {
		mutation := orchestrationCreateMutation(t, state)
		mutation.Revision.WorkerSpecSnapshotID = revisionSnapshotID
		tokenBudget := int64(200000)
		return workerplanner.GoalLoopApplyMutation{
			ApplyMutation: mutation,
			Projection: workerplanner.GoalLoopApplyProjection{
				Name:        "Checkout Recovery",
				Description: "Restore checkout reliability",
				Objective:   "Fix checkout",
				AcceptanceCriteria: []string{
					"Tests pass",
					"Evidence recorded",
				},
				VerificationCommand:  "go test ./...",
				MaxIterations:        100,
				TokenBudget:          &tokenBudget,
				TimeoutMinutes:       1440,
				NoProgressLimit:      20,
				SameErrorLimit:       20,
				EscalationPolicy:     "fail",
				WorkerSpecSnapshotID: projectionSnapshotID,
			},
		}, nil
	}
}

func orchestrationGoalLoopApplyPlan(
	t *testing.T,
	snapshotID int64,
) control.Plan {
	t.Helper()
	plan := orchestrationApplyTestCreatePlan(t)
	tokenBudget := int64(200000)
	manifest := resource.Manifest{
		TypeMeta: resource.TypeMeta{
			APIVersion: resource.APIVersionV1Alpha1,
			Kind:       resource.KindGoalLoop,
		},
		Metadata: resource.Metadata{
			Name: "checkout-recovery", Namespace: "team-alpha",
			DisplayName: "Checkout Recovery", Labels: map[string]string{},
		},
		Spec: expertCanonicalJSON(t, resource.GoalLoopResourceSpec{
			WorkerTemplateRef: resource.Reference{
				Kind: resource.KindWorkerTemplate, Name: "review-worker",
			},
			Description: "Restore checkout reliability",
			Objective:   "Fix checkout",
			AcceptanceCriteria: []string{
				"Tests pass",
				"Evidence recorded",
			},
			VerificationCommand: "go test ./...",
			MaxIterations:       100,
			TokenBudget:         &tokenBudget,
			TimeoutMinutes:      1440,
			NoProgressLimit:     20,
			SameErrorLimit:      20,
			EscalationPolicy:    "fail",
		}),
	}
	plan.Target = control.ResourceTarget{
		TypeMeta: manifest.TypeMeta, Namespace: manifest.Metadata.Namespace,
		Name: manifest.Metadata.Name,
	}
	plan.CanonicalManifest = expertCanonicalJSON(t, manifest)
	plan.DraftHash = expertDigestJSON(t, plan.CanonicalManifest)
	plan.ArtifactKind = resource.KindGoalLoop + "Apply"
	plan.ArtifactJSON = expertCanonicalJSON(
		t,
		workerplanner.DefinitionApplyArtifact{
			WorkerSpecSnapshotID: snapshotID,
		},
	)
	plan.ArtifactDigest = expertDigestJSON(t, plan.ArtifactJSON)
	plan.ResolvedReferences = []control.ResolvedReference{}
	plan.PlanHash = expertPlanHash(t, plan)
	return plan
}
