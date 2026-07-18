package expert

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	poddomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	sessiondomain "github.com/anthropics/agentsmesh/backend/internal/domain/agentsession"
	itemdomain "github.com/anthropics/agentsmesh/backend/internal/domain/conversationitem"
	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	skilldomain "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func TestPublishFromPodBindsWorkerSpecSnapshot(t *testing.T) {
	snapshotID := int64(42)
	repositoryID := int64(23)
	branch := "legacy-branch"
	prompt := "legacy prompt"
	store := newFakeStore()
	snapshots := &expertSnapshotLoader{snapshot: expertWorkerSpecSnapshot(snapshotID, 7)}
	service := NewService(Deps{
		Store: store,
		Pods: &expertPodLoader{pod: &poddomain.Pod{
			OrganizationID:  7,
			PodKey:          "pod-source",
			AgentSlug:       "legacy-agent",
			RunnerID:        99,
			RepositoryID:    &repositoryID,
			BranchName:      &branch,
			Prompt:          prompt,
			InteractionMode: expertdom.InteractionModePTY,
			PodResourceBindings: poddomain.PodResourceBindings{
				WorkerSpecSnapshotID: &snapshotID,
			},
		}},
		WorkerSpecs: snapshots,
	})

	row, err := service.PublishFromPod(context.Background(), &PublishFromPodRequest{
		OrganizationID: 7,
		UserID:         5,
		PodKey:         "pod-source",
		Name:           "Review Expert",
		Slug:           "review-expert",
	})

	require.NoError(t, err)
	require.NotNil(t, row.WorkerSpecSnapshotID)
	assert.Equal(t, snapshotID, *row.WorkerSpecSnapshotID)
	require.NotNil(t, row.SourcePodKey)
	assert.Equal(t, "pod-source", *row.SourcePodKey)
	assert.Equal(t, "codex-cli", row.AgentSlug)
	assert.Equal(t, expertdom.InteractionModeACP, row.InteractionMode)
	assert.Equal(t, expertdom.AutomationLevelAutonomous, row.AutomationLevel)
	assert.Nil(t, row.RunnerID)
	assert.Nil(t, row.RepositoryID)
	assert.Nil(t, row.BranchName)
	assert.Nil(t, row.Prompt)
	assert.Nil(t, row.AgentfileLayer)
	assert.Empty(t, row.UsedEnvBundles)
	assert.Empty(t, row.SkillSlugs)
	assert.Equal(t, int64(7), snapshots.organizationID)
	assert.Equal(t, snapshotID, snapshots.snapshotID)
}

func TestPublishFromPodRejectsMissingWorkerSpecSnapshot(t *testing.T) {
	store := newFakeStore()
	service := NewService(Deps{
		Store: store,
		Pods: &expertPodLoader{pod: &poddomain.Pod{
			OrganizationID: 7,
			PodKey:         "pod-legacy",
		}},
		WorkerSpecs: &expertSnapshotLoader{},
	})

	_, err := service.PublishFromPod(context.Background(), &PublishFromPodRequest{
		OrganizationID: 7,
		UserID:         5,
		PodKey:         "pod-legacy",
		Name:           "Legacy",
	})

	require.ErrorIs(t, err, ErrPodWorkerSpecSnapshotRequired)
	assert.Empty(t, store.rows)
}

func TestPublishFromPodCachesSnapshotSkillSlugs(t *testing.T) {
	snapshotID := int64(42)
	snapshot := expertWorkerSpecSnapshot(snapshotID, 7)
	snapshot.Spec.Workspace.SkillIDs = []int64{11}
	service := NewService(Deps{
		Store: newFakeStore(),
		Pods: &expertPodLoader{pod: &poddomain.Pod{
			OrganizationID: 7,
			PodKey:         "pod-source",
			PodResourceBindings: poddomain.PodResourceBindings{
				WorkerSpecSnapshotID: &snapshotID,
			},
		}},
		WorkerSpecs: &expertSnapshotLoader{snapshot: snapshot},
		Skills: &expertSkillLoader{skills: map[int64]*skilldomain.Skill{
			11: {
				ID:             11,
				OrganizationID: workerSpecTestInt64Pointer(7),
				Slug:           "seedance-expert",
			},
		}},
	})

	row, err := service.PublishFromPod(context.Background(), &PublishFromPodRequest{
		OrganizationID: 7,
		UserID:         5,
		PodKey:         "pod-source",
		Name:           "Seedance Expert",
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"seedance-expert"}, []string(row.SkillSlugs))
}

func TestPublishFromPodRejectsInvalidWorkerSpecSnapshotID(t *testing.T) {
	snapshotID := int64(0)
	store := newFakeStore()
	snapshots := &expertSnapshotLoader{
		snapshot: expertWorkerSpecSnapshot(snapshotID, 7),
	}
	service := NewService(Deps{
		Store: store,
		Pods: &expertPodLoader{pod: &poddomain.Pod{
			OrganizationID: 7,
			PodKey:         "pod-invalid",
			PodResourceBindings: poddomain.PodResourceBindings{
				WorkerSpecSnapshotID: &snapshotID,
			},
		}},
		WorkerSpecs: snapshots,
	})

	_, err := service.PublishFromPod(context.Background(), &PublishFromPodRequest{
		OrganizationID: 7,
		UserID:         5,
		PodKey:         "pod-invalid",
		Name:           "Invalid",
	})

	require.ErrorIs(t, err, ErrWorkerSpecSnapshotMismatch)
	assert.Zero(t, snapshots.snapshotID)
	assert.Empty(t, store.rows)
}

func TestPublishFromPodRejectsCrossOrganizationSnapshot(t *testing.T) {
	snapshotID := int64(42)
	store := newFakeStore()
	service := NewService(Deps{
		Store: store,
		Pods: &expertPodLoader{pod: &poddomain.Pod{
			OrganizationID: 7,
			PodKey:         "pod-source",
			PodResourceBindings: poddomain.PodResourceBindings{
				WorkerSpecSnapshotID: &snapshotID,
			},
		}},
		WorkerSpecs: &expertSnapshotLoader{
			snapshot: expertWorkerSpecSnapshot(snapshotID, 8),
		},
	})

	_, err := service.PublishFromPod(context.Background(), &PublishFromPodRequest{
		OrganizationID: 7,
		UserID:         5,
		PodKey:         "pod-source",
		Name:           "Cross Org",
	})

	require.ErrorIs(t, err, ErrWorkerSpecSnapshotMismatch)
	assert.Empty(t, store.rows)
}

func TestRunLegacyExpertRequiresRepublish(t *testing.T) {
	store := newFakeStore()
	require.NoError(t, store.Create(context.Background(), &expertdom.Expert{
		OrganizationID: 7,
		Slug:           "legacy",
		Name:           "Legacy",
		AgentSlug:      "codex-cli",
	}))
	dispatcher := &fakeDispatcher{}
	service := NewService(Deps{Store: store, Dispatch: dispatcher})

	_, err := service.Run(context.Background(), &RunExpertRequest{
		OrganizationID: 7,
		UserID:         5,
		ExpertSlug:     "legacy",
	})

	require.ErrorIs(t, err, ErrExpertRepublishRequired)
	assert.Nil(t, dispatcher.lastReq)
}

func TestRunDispatchesSnapshotWithOnlyAliasAndPromptOverrides(t *testing.T) {
	snapshotID := int64(42)
	store := newFakeStore()
	require.NoError(t, store.Create(context.Background(), &expertdom.Expert{
		OrganizationID:       7,
		Slug:                 "review",
		Name:                 "Review",
		AgentSlug:            "legacy-agent",
		RunnerID:             workerSpecTestInt64Pointer(88),
		WorkerSpecSnapshotID: &snapshotID,
	}))
	dispatcher := &fakeDispatcher{}
	items := &expertConversationItemWriter{position: 1}
	service := NewService(Deps{
		Store:       store,
		Dispatch:    dispatcher,
		WorkerSpecs: &expertSnapshotLoader{snapshot: expertWorkerSpecSnapshot(snapshotID, 7)},
		Items:       items,
	})
	alias := "one-off"
	prompt := "Review only the security boundary."

	_, err := service.Run(context.Background(), &RunExpertRequest{
		OrganizationID: 7,
		UserID:         5,
		ExpertSlug:     "review",
		Alias:          &alias,
		PromptOverride: &prompt,
	})

	require.NoError(t, err)
	require.NotNil(t, dispatcher.lastReq)
	require.NotNil(t, dispatcher.lastReq.WorkerSpecSnapshotID)
	assert.Equal(t, snapshotID, *dispatcher.lastReq.WorkerSpecSnapshotID)
	require.NotNil(t, dispatcher.lastReq.WorkerSpecPromptOverride)
	assert.Equal(t, prompt, *dispatcher.lastReq.WorkerSpecPromptOverride)
	assert.Equal(t, &alias, dispatcher.lastReq.Alias)
	assert.Zero(t, dispatcher.lastReq.RunnerID)
	assert.Empty(t, dispatcher.lastReq.AgentSlug)
	assert.Nil(t, dispatcher.lastReq.RepositoryID)
	assert.Nil(t, dispatcher.lastReq.AgentfileLayer)
	require.NotNil(t, dispatcher.lastReq.SessionProvision)
	assert.False(t, dispatcher.lastReq.SessionProvision.UpdateExisting)
	require.NotNil(t, dispatcher.lastReq.PrepareSession)
	require.NoError(t, dispatcher.lastReq.PrepareSession(
		context.Background(),
		&sessiondomain.Session{ID: "conv-review"},
	))
	require.NotNil(t, items.item)
	assert.Equal(t, "conv-review", items.item.SessionID)
	var payload struct {
		Role    string `json:"role"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	require.NoError(t, json.Unmarshal(items.item.Payload, &payload))
	assert.Equal(t, "user", payload.Role)
	require.Len(t, payload.Content, 1)
	assert.Equal(t, prompt, payload.Content[0].Text)
}

func TestRunPersistsSnapshotInitialTaskWithoutPromptOverride(t *testing.T) {
	snapshotID := int64(42)
	store := newFakeStore()
	require.NoError(t, store.Create(context.Background(), &expertdom.Expert{
		OrganizationID:       7,
		Slug:                 "review",
		Name:                 "Review",
		WorkerSpecSnapshotID: &snapshotID,
	}))
	dispatcher := &fakeDispatcher{}
	items := &expertConversationItemWriter{position: 1}
	service := NewService(Deps{
		Store:       store,
		Dispatch:    dispatcher,
		WorkerSpecs: &expertSnapshotLoader{snapshot: expertWorkerSpecSnapshot(snapshotID, 7)},
		Items:       items,
	})

	_, err := service.Run(context.Background(), &RunExpertRequest{
		OrganizationID: 7,
		UserID:         5,
		ExpertSlug:     "review",
	})

	require.NoError(t, err)
	require.NotNil(t, dispatcher.lastReq)
	require.NotNil(t, dispatcher.lastReq.PrepareSession)
	require.NoError(t, dispatcher.lastReq.PrepareSession(
		context.Background(),
		&sessiondomain.Session{ID: "conv-review"},
	))
	require.NotNil(t, items.item)
	var payload struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	require.NoError(t, json.Unmarshal(items.item.Payload, &payload))
	require.Len(t, payload.Content, 1)
	assert.Equal(t, "Run checks.", payload.Content[0].Text)
}

type expertConversationItemWriter struct {
	position int64
	item     *itemdomain.Item
}

func (writer *expertConversationItemWriter) NextPosition(
	context.Context,
	string,
) (int64, error) {
	return writer.position, nil
}

func (writer *expertConversationItemWriter) Append(
	_ context.Context,
	item *itemdomain.Item,
) error {
	writer.item = item
	return nil
}

type expertPodLoader struct {
	pod *poddomain.Pod
	err error
}

func (loader *expertPodLoader) GetPod(context.Context, string) (*poddomain.Pod, error) {
	return loader.pod, loader.err
}

type expertSnapshotLoader struct {
	snapshot       specdomain.Snapshot
	err            error
	organizationID int64
	snapshotID     int64
}

type expertSkillLoader struct {
	skills map[int64]*skilldomain.Skill
}

func (loader *expertSkillLoader) GetAnyByID(
	_ context.Context,
	id int64,
) (*skilldomain.Skill, error) {
	row, ok := loader.skills[id]
	if !ok {
		return nil, skilldomain.ErrNotFound
	}
	return row, nil
}

func (loader *expertSnapshotLoader) GetByID(
	_ context.Context,
	organizationID, snapshotID int64,
) (specdomain.Snapshot, error) {
	loader.organizationID = organizationID
	loader.snapshotID = snapshotID
	return loader.snapshot, loader.err
}

func expertWorkerSpecSnapshot(id, organizationID int64) specdomain.Snapshot {
	spec := specdomain.NewV1(
		specdomain.Runtime{
			ModelBinding: specdomain.ModelBinding{
				ResourceID:         101,
				ResourceRevision:   7,
				ConnectionID:       201,
				ConnectionRevision: 9,
				ProviderKey:        slugkit.MustNewForTest("openai"),
				ProtocolAdapter:    slugkit.MustNewForTest("openai-compatible"),
				ModelID:            "gpt-5",
			},
			WorkerType: specdomain.WorkerType{
				Slug:           slugkit.MustNewForTest("codex-cli"),
				DefinitionHash: strings.Repeat("a", 64),
			},
			Image: specdomain.RuntimeImage{
				ID:     1,
				Digest: "sha256:" + strings.Repeat("b", 64),
			},
		},
		specdomain.Placement{
			Policy: specdomain.PlacementPolicyExplicit,
			ComputeTarget: specdomain.ComputeTarget{
				ID:   1,
				Kind: specdomain.ComputeTargetKindRunnerPool,
			},
			DeploymentMode: specdomain.DeploymentModePooled,
			ResourceProfile: specdomain.ResourceProfile{
				ID: 1,
				Resources: specdomain.ResourceRequestsLimits{
					CPURequestMilliCPU: 200,
					CPULimitMilliCPU:   1000,
					MemoryRequestBytes: 256 << 20,
					MemoryLimitBytes:   1 << 30,
				},
			},
		},
		specdomain.TypeConfig{
			SchemaVersion:   1,
			Values:          map[string]any{},
			SecretRefs:      map[string]specdomain.SecretReference{},
			InteractionMode: specdomain.InteractionModeACP,
			AutomationLevel: specdomain.AutomationLevelAutonomous,
		},
		specdomain.Workspace{
			SkillIDs:        []int64{},
			KnowledgeMounts: []specdomain.KnowledgeMount{},
			EnvBundleIDs:    []specdomain.RuntimeEnvBundleID{},
			InitialTask:     "Run checks.",
		},
		specdomain.Lifecycle{TerminationPolicy: specdomain.TerminationPolicyManual},
		specdomain.Metadata{Alias: "review-worker"},
	)
	return specdomain.Snapshot{ID: id, OrganizationID: organizationID, Spec: spec}
}

func workerSpecTestInt64Pointer(value int64) *int64 {
	return &value
}
