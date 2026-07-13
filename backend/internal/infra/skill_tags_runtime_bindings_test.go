package infra

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	extensionsvc "github.com/anthropics/agentsmesh/backend/internal/service/extension"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
	skillsvc "github.com/anthropics/agentsmesh/backend/internal/service/skill"
	workerspecservice "github.com/anthropics/agentsmesh/backend/internal/service/workerspec"
)

func TestSkillTagUpdatePreservesExpertAndWorkerSpecBindings(t *testing.T) {
	ctx := context.Background()
	db := workerSpecSnapshotDBForContract(t)
	require.NoError(t, db.Exec(
		"ALTER TABLE skills ADD COLUMN tags TEXT NOT NULL DEFAULT '{}'",
	).Error)
	createRuntimeBindingExpertsTable(t, db)

	skillRepo := NewSkillCatalogRepository(db)
	skillService := skillsvc.NewService(skillsvc.Deps{
		Store:    skillRepo,
		Gitops:   gitops.NewFake("am-skills"),
		Packager: runtimeBindingPackager{},
	})
	skillRow, err := skillService.Create(ctx, &skillsvc.CreateSkillRequest{
		OrganizationID: 77,
		UserID:         7,
		Name:           "Video Editing",
		Instructions:   "Edit the video.",
		Tags:           []string{"video"},
	})
	require.NoError(t, err)

	snapshotRepo := NewWorkerSpecSnapshotRepository(db)
	snapshot, err := snapshotRepo.Create(
		ctx,
		workerSpecSnapshotForSkill(t, 77, skillRow.ID),
	)
	require.NoError(t, err)

	expertRepo := NewExpertRepository(db)
	expertRow := &expertdom.Expert{
		OrganizationID:       77,
		Slug:                 "video-editor",
		Name:                 "Video Editor",
		AgentSlug:            "codex-cli",
		InteractionMode:      expertdom.InteractionModePTY,
		AutomationLevel:      expertdom.AutomationLevelAutonomous,
		SkillSlugs:           pq.StringArray{"video-editing", "video-delivery-qa"},
		KnowledgeMounts:      []byte("[]"),
		ConfigOverrides:      []byte("{}"),
		Metadata:             []byte("{}"),
		WorkerSpecSnapshotID: &snapshot.ID,
		CreatedByID:          7,
	}
	require.NoError(t, expertRepo.Create(ctx, expertRow))

	beforeExpert, err := expertRepo.GetByID(ctx, 77, expertRow.ID)
	require.NoError(t, err)
	beforeSnapshot, err := snapshotRepo.GetByID(ctx, 77, snapshot.ID)
	require.NoError(t, err)
	beforeJSON := readWorkerSpecJSON(t, db, snapshot.ID)

	tags := []string{"Editing", " Motion "}
	_, err = skillService.Update(ctx, &skillsvc.UpdateSkillRequest{
		OrganizationID: 77,
		SkillID:        skillRow.ID,
		Tags:           &tags,
	})
	require.NoError(t, err)

	updatedSkill, err := skillRepo.GetByID(ctx, 77, skillRow.ID)
	require.NoError(t, err)
	afterExpert, err := expertRepo.GetByID(ctx, 77, expertRow.ID)
	require.NoError(t, err)
	afterSnapshot, err := snapshotRepo.GetByID(ctx, 77, snapshot.ID)
	require.NoError(t, err)
	afterJSON := readWorkerSpecJSON(t, db, snapshot.ID)

	assert.Equal(t, []string{"editing", "motion"}, []string(updatedSkill.Tags))
	assert.Equal(t, beforeExpert.SkillSlugs, afterExpert.SkillSlugs)
	assert.Equal(t, beforeExpert.WorkerSpecSnapshotID, afterExpert.WorkerSpecSnapshotID)
	assert.Equal(t, beforeSnapshot.Spec, afterSnapshot.Spec)
	assert.Equal(t, beforeSnapshot.Summary, afterSnapshot.Summary)
	assert.Equal(t, beforeJSON, afterJSON)
}

type runtimeBindingPackager struct{}

func (runtimeBindingPackager) DeletePackage(context.Context, string) error {
	return nil
}

func (runtimeBindingPackager) PrepareCatalogFromDir(
	_ context.Context,
	dir, _ string,
) (*extensionsvc.PreparedSkill, error) {
	content, err := os.ReadFile(filepath.Join(dir, "skill.json"))
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(content)
	return &extensionsvc.PreparedSkill{
		Slug:        "video-editing",
		ContentSha:  hex.EncodeToString(sum[:]),
		StorageKey:  "skills/video-editing.tar.gz",
		PackageSize: int64(len(content)),
		Data:        content,
	}, nil
}

func (runtimeBindingPackager) StorePrepared(
	_ context.Context,
	prepared *extensionsvc.PreparedSkill,
) (*extensionsvc.PackagedSkill, error) {
	return &extensionsvc.PackagedSkill{
		Slug:        prepared.Slug,
		ContentSha:  prepared.ContentSha,
		StorageKey:  prepared.StorageKey,
		PackageSize: prepared.PackageSize,
	}, nil
}

type storedWorkerSpecJSON struct {
	Spec    []byte `gorm:"column:spec_json"`
	Summary []byte `gorm:"column:summary_json"`
}

func readWorkerSpecJSON(t *testing.T, db *gorm.DB, snapshotID int64) storedWorkerSpecJSON {
	t.Helper()
	var stored storedWorkerSpecJSON
	require.NoError(t, db.Raw(
		"SELECT spec_json, summary_json FROM worker_spec_snapshots WHERE id = ?",
		snapshotID,
	).Scan(&stored).Error)
	return stored
}

func createRuntimeBindingExpertsTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE experts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		organization_id INTEGER NOT NULL,
		slug TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		agent_slug TEXT NOT NULL,
		runner_id INTEGER,
		repository_id INTEGER,
		branch_name TEXT,
		prompt TEXT,
		interaction_mode TEXT NOT NULL,
		automation_level TEXT NOT NULL,
		perpetual INTEGER NOT NULL DEFAULT 0,
		used_env_bundles TEXT NOT NULL DEFAULT '{}',
		skill_slugs TEXT NOT NULL DEFAULT '{}',
		knowledge_mounts BLOB NOT NULL DEFAULT '[]',
		config_overrides BLOB NOT NULL DEFAULT '{}',
		agentfile_layer TEXT,
		source_pod_key TEXT,
		worker_spec_snapshot_id INTEGER,
		source_market_application_id INTEGER,
		source_market_release_id INTEGER,
		git_repo_path TEXT,
		default_branch TEXT NOT NULL DEFAULT 'main',
		http_clone_url TEXT,
		metadata BLOB NOT NULL DEFAULT '{}',
		created_by_id INTEGER NOT NULL,
		run_count INTEGER NOT NULL DEFAULT 0,
		last_run_at DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error)
}

func workerSpecSnapshotForSkill(
	t *testing.T,
	organizationID, skillID int64,
) workerspecservice.ResolvedSnapshot {
	t.Helper()
	spec := workerSpecForRepoContract()
	spec.Workspace.SkillIDs = []int64{skillID}
	ports := &workerSpecResolutionPorts{spec: spec}
	resolver := workerspecservice.NewResolver(workerspecservice.ResolverDeps{
		WorkerTypes: ports,
		Runtime:     ports,
		Models:      ports,
		Secrets:     ports,
		Workspaces:  ports,
	})
	snapshot, err := resolver.Resolve(
		context.Background(),
		workerspecservice.Scope{OrgID: organizationID, UserID: 7},
		workerspecservice.Draft{
			ModelResourceID: spec.Runtime.ModelBinding.ResourceID,
			WorkerTypeSlug:  spec.Runtime.WorkerType.Slug,
			Runtime: workerspecservice.RuntimeSelection{
				RuntimeImageID:    spec.Runtime.Image.ID,
				PlacementPolicy:   spec.Placement.Policy,
				ComputeTargetID:   spec.Placement.ComputeTarget.ID,
				DeploymentMode:    spec.Placement.DeploymentMode,
				ResourceProfileID: spec.Placement.ResourceProfile.ID,
			},
			TypeConfig: spec.TypeConfig,
			Workspace:  spec.Workspace,
			Lifecycle:  spec.Lifecycle,
			Metadata:   spec.Metadata,
		},
	)
	require.NoError(t, err)
	return snapshot
}
