package infra

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	control "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationcontrol"
	resource "github.com/anthropics/agentsmesh/backend/internal/domain/orchestrationresource"
	"github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	controlservice "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationcontrol"
	workerplanner "github.com/anthropics/agentsmesh/backend/internal/service/orchestrationworker"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerTemplateApplyPersistsSnapshotAndRevisionAtomically(t *testing.T) {
	db, _ := orchestrationPostgresRepository(t)
	repository := NewOrchestrationResourceRepository(db)
	plan := orchestrationWorkerTemplateApplyPlan(t)
	require.NoError(t, repository.CreatePlan(context.Background(), plan))

	result, err := repository.RunWorkerTemplateApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(
			state controlservice.LockedApplyState,
			snapshotID int64,
		) (controlservice.ApplyMutation, error) {
			require.Positive(t, snapshotID)
			mutation := orchestrationCreateMutation(t, state)
			mutation.Revision.WorkerSpecSnapshotID = snapshotID
			return mutation, nil
		},
	)

	require.NoError(t, err)
	assert.Positive(t, result.WorkerSpecSnapshotID)
	assert.Equal(t, result.Head.ID, planResultResourceID(t, repository, plan))
	revision, err := repository.GetRevision(
		context.Background(),
		plan.Scope,
		result.Head.ID,
		1,
	)
	require.NoError(t, err)
	assert.Equal(t, result.WorkerSpecSnapshotID, revision.WorkerSpecSnapshotID)
	var snapshots int64
	require.NoError(t, db.Table("worker_spec_snapshots").Count(&snapshots).Error)
	assert.Equal(t, int64(1), snapshots)
}

func TestWorkerTemplateApplyRollsBackSnapshotOnBuilderFailure(t *testing.T) {
	db, _ := orchestrationPostgresRepository(t)
	repository := NewOrchestrationResourceRepository(db)
	plan := orchestrationWorkerTemplateApplyPlan(t)
	require.NoError(t, repository.CreatePlan(context.Background(), plan))
	builderErr := errors.New("mutation build failed")

	_, err := repository.RunWorkerTemplateApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(
			controlservice.LockedApplyState,
			int64,
		) (controlservice.ApplyMutation, error) {
			return controlservice.ApplyMutation{}, builderErr
		},
	)

	assert.ErrorIs(t, err, builderErr)
	var snapshots, resources int64
	require.NoError(t, db.Table("worker_spec_snapshots").Count(&snapshots).Error)
	require.NoError(t, db.Table("orchestration_resources").Count(&resources).Error)
	assert.Zero(t, snapshots)
	assert.Zero(t, resources)
	stored, loadErr := repository.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, loadErr)
	assert.Equal(t, control.PlanStatusPending, stored.Status)
}

func TestWorkerTemplateApplyRejectsWrongTargetBeforeBuilder(t *testing.T) {
	_, repository := orchestrationPostgresRepository(t)
	plan := orchestrationApplyTestCreatePlan(t)
	plan.Target.Kind = resource.KindModelBinding
	var manifest resource.Manifest
	require.NoError(t, json.Unmarshal(plan.CanonicalManifest, &manifest))
	manifest.Kind = resource.KindModelBinding
	plan.CanonicalManifest = mustCanonicalObject(t, manifest)
	var err error
	plan.DraftHash, err = control.DigestCanonicalJSON(plan.CanonicalManifest)
	require.NoError(t, err)
	plan.PlanHash, err = control.ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	require.NoError(t, repository.CreatePlan(context.Background(), plan))

	_, err = repository.RunWorkerTemplateApplyTransaction(
		context.Background(),
		plan.Scope,
		plan.ID,
		func(
			controlservice.LockedApplyState,
			int64,
		) (controlservice.ApplyMutation, error) {
			t.Fatal("wrong target must not invoke builder")
			return controlservice.ApplyMutation{}, nil
		},
	)

	assert.ErrorIs(t, err, control.ErrInvalid)
}

func orchestrationWorkerTemplateApplyPlan(t *testing.T) control.Plan {
	t.Helper()
	plan := orchestrationApplyTestCreatePlan(t)
	plan.Target.Kind = resource.KindWorkerTemplate
	manifest := resource.Manifest{
		TypeMeta: plan.Target.TypeMeta,
		Metadata: resource.Metadata{
			Name: plan.Target.Name, Namespace: plan.Target.Namespace,
			DisplayName: "Review Worker", Labels: map[string]string{},
		},
		Spec: mustCanonicalObject(t, map[string]any{"template": "planned"}),
	}
	plan.CanonicalManifest = mustCanonicalObject(t, manifest)
	plan.ArtifactKind = "WorkerSpec"
	plan.ArtifactJSON = validApplyWorkerSpecJSON(t)
	var err error
	plan.DraftHash, err = control.DigestCanonicalJSON(plan.CanonicalManifest)
	require.NoError(t, err)
	plan.ArtifactDigest, err = control.DigestCanonicalJSON(plan.ArtifactJSON)
	require.NoError(t, err)
	plan.PlanHash, err = control.ComputePlanHash(plan.HashInput())
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	return plan
}

func validApplyWorkerSpecJSON(t *testing.T) []byte {
	t.Helper()
	spec := workerspec.NewV1(
		workerspec.Runtime{
			WorkerType: workerspec.WorkerType{
				Slug:           slugkit.MustNewForTest("codex"),
				DefinitionHash: strings.Repeat("a", 64),
			},
			Image: workerspec.RuntimeImage{
				ID: 31, Digest: "sha256:" + strings.Repeat("b", 64),
			},
		},
		workerspec.Placement{
			Policy: workerspec.PlacementPolicyExplicit,
			ComputeTarget: workerspec.ComputeTarget{
				ID: 41, Kind: workerspec.ComputeTargetKindRunnerPool,
			},
			DeploymentMode: workerspec.DeploymentModeDedicated,
			ResourceProfile: workerspec.ResourceProfile{
				ID: 51,
				Resources: workerspec.ResourceRequestsLimits{
					CPURequestMilliCPU: 500, CPULimitMilliCPU: 1000,
					MemoryRequestBytes: 512 << 20, MemoryLimitBytes: 1024 << 20,
				},
			},
		},
		workerspec.TypeConfig{
			SchemaVersion: 1, Values: map[string]any{},
			SecretRefs:      map[string]workerspec.SecretReference{},
			InteractionMode: workerspec.InteractionModeACP,
			AutomationLevel: workerspec.AutomationLevelInteractive,
		},
		workerspec.Workspace{
			SkillIDs: []int64{}, KnowledgeMounts: []workerspec.KnowledgeMount{},
			EnvBundleIDs:    []workerspec.RuntimeEnvBundleID{},
			ConfigBundleIDs: []int64{},
		},
		workerspec.Lifecycle{
			TerminationPolicy: workerspec.TerminationPolicyManual,
		},
		workerspec.Metadata{Alias: "Review Worker"},
	)
	encoded, err := workerspec.EncodeSpec(spec)
	require.NoError(t, err)
	canonical, err := control.CanonicalJSONObject(encoded)
	require.NoError(t, err)
	return canonical
}

func mustCanonicalObject(t *testing.T, value any) []byte {
	t.Helper()
	canonical, err := control.CanonicalJSONObject(value)
	require.NoError(t, err)
	return canonical
}

func planResultResourceID(
	t *testing.T,
	repository controlservice.Repository,
	plan control.Plan,
) int64 {
	t.Helper()
	stored, err := repository.GetPlan(context.Background(), plan.Scope, plan.ID)
	require.NoError(t, err)
	return stored.ResultResourceID
}

var _ workerplanner.WorkerTemplateApplyRepository = (*orchestrationResourceRepo)(nil)
