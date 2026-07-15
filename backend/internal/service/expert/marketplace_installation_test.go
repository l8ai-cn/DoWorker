package expert

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/stretchr/testify/require"
)

func TestInstallMarketplaceExpertClonesTemplateIdempotently(t *testing.T) {
	store := newFakeStore()
	store.nextID = 200
	snapshots := &fakeMarketSnapshots{}
	skills := &fakeMarketSkills{
		rows: []skilldom.Skill{
			marketSkill("delivery-worktree", nil, true),
			marketSkill("delivery-e2e", nil, true),
			marketSkill("delivery-github-merge", nil, true),
			marketSkill("delivery-gitlab-merge", nil, true),
		},
	}
	svc := NewService(Deps{
		Store:             store,
		WorkerSpecWriter:  snapshots,
		MarketWorkerSpecs: snapshots,
		MarketSkills:      skills,
	})
	source := expertWorkerSpecSnapshot(0, 0).Spec
	source.Workspace.Instructions = "负责把明确需求转化为经过测试、评审并可合并的代码交付。"
	source.Workspace.InitialTask = ""
	source.Workspace.SkillIDs = nil
	source.Workspace.SkillPackages = make(
		[]specdomain.SkillPackageBinding,
		0,
		len(skills.rows),
	)
	for _, skill := range skills.rows {
		source.Workspace.SkillPackages = append(
			source.Workspace.SkillPackages,
			specdomain.SkillPackageBinding{
				SkillID: skill.ID, Slug: skill.Slug, Version: 1,
				ContentSHA:  "approved-" + skill.Slug,
				StorageKey:  "approved/" + skill.Slug,
				PackageSize: 123,
			},
		)
	}
	sourceJSON, err := json.Marshal(source)
	require.NoError(t, err)
	request := MarketplaceInstallationRequest{
		InstallationID:       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		TargetOrganizationID: 9, ActorUserID: 14,
		ModelResourceID:           301,
		SourceMarketApplicationID: 101,
		SourceMarketReleaseID:     201,
		RuntimeSnapshot: append([]byte(`{"version":1,"expert":{
			"version":1,
			"slug":"software-delivery-expert",
			"name":"软件交付专家",
			"description":"适用于功能开发、缺陷修复和版本交付。",
			"agent_slug":"codex-cli",
			"prompt":"负责把明确需求转化为经过测试、评审并可合并的代码交付。",
			"interaction_mode":"acp",
			"automation_level":"autonomous",
			"perpetual":false,
			"used_env_bundles":[],
			"skill_slugs":["delivery-worktree","delivery-e2e","delivery-github-merge","delivery-gitlab-merge"],
			"knowledge_mounts":[],
			"config_overrides":{},
			"metadata":{}
		},
			"worker_spec":`), append(sourceJSON, '}')...),
	}

	first, existing, err := svc.InstallMarketplaceExpert(context.Background(), request)
	require.NoError(t, err)
	require.False(t, existing)
	require.Equal(t, int64(9), first.OrganizationID)
	require.Equal(t, "market-aaaaaaaaaaaa4aaa8aaaaaaaaaaaaaaa", first.Slug)
	require.Equal(t, "软件交付专家", first.Name)
	require.Equal(t, "codex-cli", first.AgentSlug)
	require.Equal(t, int64(101), *first.SourceMarketApplicationID)
	require.Equal(t, int64(201), *first.SourceMarketReleaseID)
	require.Nil(t, first.RunnerID)
	require.Nil(t, first.RepositoryID)
	require.Nil(t, first.BranchName)
	require.Equal(t, []string{
		"delivery-worktree",
		"delivery-e2e",
		"delivery-github-merge",
		"delivery-gitlab-merge",
	}, []string(first.SkillSlugs))
	require.NotNil(t, first.WorkerSpecSnapshotID)
	require.Equal(t, int64(301), snapshots.preparedModels[0])
	require.Equal(t, int64(9), snapshots.created[0].OrganizationID)
	require.Equal(
		t,
		source.Workspace.SkillPackages,
		snapshots.created[0].Spec.Workspace.SkillPackages,
	)

	secondRequest := request
	secondRequest.InstallationID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	second, existing, err := svc.InstallMarketplaceExpert(
		context.Background(),
		secondRequest,
	)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, first.ID, second.ID)
	require.Len(t, store.rows, 1)
}

func TestInstallMarketplaceExpertRejectsIncompleteSnapshot(t *testing.T) {
	svc := NewService(Deps{Store: newFakeStore()})

	_, _, err := svc.InstallMarketplaceExpert(context.Background(), MarketplaceInstallationRequest{
		InstallationID:       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		TargetOrganizationID: 9,
		ActorUserID:          14,
		ModelResourceID:      301,
		RuntimeSnapshot:      []byte(`{"name":"缺少运行时"}`),
	})

	require.ErrorIs(t, err, ErrMarketplaceInstallationInvalid)
}

func TestInstallMarketplaceExpertCleansSnapshotAfterRequestCancellation(t *testing.T) {
	store := newFakeStore()
	store.createErr = errors.New("create failed")
	snapshots := &fakeMarketSnapshots{}
	svc := NewService(Deps{
		Store:             store,
		WorkerSpecWriter:  snapshots,
		MarketWorkerSpecs: snapshots,
		MarketSkills: &fakeMarketSkills{rows: []skilldom.Skill{
			marketSkill("delivery-worktree", nil, true),
		}},
	})
	source := expertWorkerSpecSnapshot(0, 0).Spec
	source.Workspace.Instructions = "deliver"
	source.Workspace.SkillIDs = nil
	sourceJSON, err := json.Marshal(source)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = svc.InstallMarketplaceExpert(ctx, MarketplaceInstallationRequest{
		InstallationID:       "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		TargetOrganizationID: 9,
		ActorUserID:          14,
		ModelResourceID:      301,
		RuntimeSnapshot: append([]byte(`{"version":1,"expert":{
			"version":1,"slug":"delivery-expert","name":"交付专家",
			"description":"交付","agent_slug":"codex-cli","prompt":"deliver",
			"interaction_mode":"acp","automation_level":"autonomous",
			"perpetual":false,"used_env_bundles":[],
			"skill_slugs":["delivery-worktree"],"knowledge_mounts":[],
			"config_overrides":{},"metadata":{}},"worker_spec":`),
			append(sourceJSON, '}')...),
	})

	require.ErrorContains(t, err, "create failed")
	require.Len(t, snapshots.deleteContexts, 1)
	require.NoError(t, snapshots.deleteErrors[0])
	require.Empty(t, snapshots.created)
}
