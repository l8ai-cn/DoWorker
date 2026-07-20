package operatorcatalog

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	expertsvc "github.com/anthropics/agentsmesh/backend/internal/service/expert"
	skillsvc "github.com/anthropics/agentsmesh/backend/internal/service/skill"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestBootstrapVideoExpertsIsIdempotent(t *testing.T) {
	skills := &bootstrapSkillStore{}
	experts := newBootstrapExpertStore()
	workers := &bootstrapWorkerPreparer{}
	snapshots := &bootstrapSnapshotStore{}
	artifacts := &bootstrapDependencyArtifactStore{}
	bootstrapper := NewBootstrapper(skills, experts, workers, snapshots, artifacts)
	request := BootstrapRequest{
		OrganizationID:   7,
		OrganizationSlug: slugkit.MustNewForTest("dev-org"),
		PublisherUserID:  11,
		ReviewerUserID:   13,
		ModelResourceID:  17,
		RuntimeImageID:   19,
	}

	first, err := bootstrapper.Run(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, BootstrapResult{
		CreatedSkills:  7,
		CreatedExperts: 3,
		Published:      3,
	}, first)
	require.Len(t, skills.rows, 7)
	require.Len(t, experts.experts, 3)
	require.Len(t, experts.published, 3)
	require.Equal(t, 3, workers.calls)
	require.Equal(t, 3, snapshots.createCalls)
	require.Equal(t, 3, artifacts.createCalls)
	expert := experts.experts["video-production-expert"]
	require.JSONEq(
		t,
		`{"approval_mode":"never"}`,
		string(expert.ConfigOverrides),
	)
	require.Equal(
		t,
		videoExpertConfigOverrides(),
		snapshots.rows[*expert.WorkerSpecSnapshotID].Spec.TypeConfig.Values,
	)

	second, err := bootstrapper.Run(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, BootstrapResult{}, second)
	require.Equal(t, 3, workers.calls)
	require.Equal(t, 3, snapshots.createCalls)
	require.Equal(t, 3, artifacts.createCalls)
}

func TestBootstrapVideoExpertsRebuildsLegacyPromptArtifact(t *testing.T) {
	skills := &bootstrapSkillStore{}
	experts := newBootstrapExpertStore()
	workers := &bootstrapWorkerPreparer{}
	snapshots := newBootstrapSnapshotStore()
	artifacts := &bootstrapDependencyArtifactStore{}
	bootstrapper := NewBootstrapper(skills, experts, workers, snapshots, artifacts)
	request := BootstrapRequest{
		OrganizationID:   7,
		OrganizationSlug: slugkit.MustNewForTest("dev-org"),
		PublisherUserID:  11,
		ReviewerUserID:   13,
		ModelResourceID:  17,
		RuntimeImageID:   19,
	}
	_, err := bootstrapper.Run(context.Background(), request)
	require.NoError(t, err)
	expert := experts.experts["video-production-expert"]
	require.NotNil(t, expert.WorkerSpecSnapshotID)
	legacySnapshotID := *expert.WorkerSpecSnapshotID
	artifacts.rows[legacySnapshotID] = bootstrapArtifactDocument(
		legacyArtifactAgentfile(),
	)

	result, err := bootstrapper.Run(context.Background(), request)
	require.NoError(t, err)

	require.Equal(t, BootstrapResult{}, result)
	require.NotEqual(t, legacySnapshotID, *expert.WorkerSpecSnapshotID)
	require.Equal(t, int64(4), *expert.WorkerSpecSnapshotID)
	require.Equal(t, 4, workers.calls)
	require.Equal(t, 4, snapshots.createCalls)
	require.Equal(t, 4, artifacts.createCalls)
	require.True(
		t,
		artifactMatchesInstructionContract(
			artifacts.rows[*expert.WorkerSpecSnapshotID],
		),
	)
}

func TestBootstrapVideoExpertsRebuildsLegacyApprovalSnapshot(t *testing.T) {
	skills := &bootstrapSkillStore{}
	experts := newBootstrapExpertStore()
	workers := &bootstrapWorkerPreparer{}
	snapshots := newBootstrapSnapshotStore()
	artifacts := &bootstrapDependencyArtifactStore{}
	bootstrapper := NewBootstrapper(skills, experts, workers, snapshots, artifacts)
	request := BootstrapRequest{
		OrganizationID:   7,
		OrganizationSlug: slugkit.MustNewForTest("dev-org"),
		PublisherUserID:  11,
		ReviewerUserID:   13,
		ModelResourceID:  17,
		RuntimeImageID:   19,
	}
	_, err := bootstrapper.Run(context.Background(), request)
	require.NoError(t, err)
	expert := experts.experts["video-production-expert"]
	legacySnapshotID := *expert.WorkerSpecSnapshotID
	snapshot := snapshots.rows[legacySnapshotID]
	snapshot.Spec.TypeConfig.Values = map[string]any{}
	snapshots.rows[legacySnapshotID] = snapshot

	result, err := bootstrapper.Run(context.Background(), request)
	require.NoError(t, err)

	require.Equal(t, BootstrapResult{}, result)
	require.NotEqual(t, legacySnapshotID, *expert.WorkerSpecSnapshotID)
	require.Equal(
		t,
		videoExpertConfigOverrides(),
		snapshots.rows[*expert.WorkerSpecSnapshotID].Spec.TypeConfig.Values,
	)
}

func TestBootstrapVideoExpertsRejectsExistingExpertDrift(t *testing.T) {
	skills := &bootstrapSkillStore{}
	experts := newBootstrapExpertStore()
	experts.experts["video-production-expert"] = &expertdom.Expert{
		ID: 1, OrganizationID: 7, Slug: "video-production-expert",
		Name: "Different expert",
	}
	bootstrapper := NewBootstrapper(
		skills,
		experts,
		&bootstrapWorkerPreparer{},
		&bootstrapSnapshotStore{},
		&bootstrapDependencyArtifactStore{},
	)

	_, err := bootstrapper.Run(context.Background(), BootstrapRequest{
		OrganizationID:   7,
		OrganizationSlug: slugkit.MustNewForTest("dev-org"),
		PublisherUserID:  11,
		ReviewerUserID:   13,
		ModelResourceID:  17,
		RuntimeImageID:   19,
	})

	require.ErrorIs(t, err, ErrCatalogConflict)
}

func TestBootstrapVideoExpertsRejectsChangedRuntimeBindings(t *testing.T) {
	tests := map[string]func(*BootstrapRequest){
		"model resource": func(request *BootstrapRequest) {
			request.ModelResourceID = 23
		},
		"runtime image": func(request *BootstrapRequest) {
			request.RuntimeImageID = 29
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			bootstrapper := NewBootstrapper(
				&bootstrapSkillStore{},
				newBootstrapExpertStore(),
				&bootstrapWorkerPreparer{},
				newBootstrapSnapshotStore(),
				&bootstrapDependencyArtifactStore{},
			)
			request := BootstrapRequest{
				OrganizationID: 7, OrganizationSlug: slugkit.MustNewForTest("dev-org"),
				PublisherUserID: 11, ReviewerUserID: 13,
				ModelResourceID: 17, RuntimeImageID: 19,
			}
			_, err := bootstrapper.Run(context.Background(), request)
			require.NoError(t, err)

			mutate(&request)
			_, err = bootstrapper.Run(context.Background(), request)

			require.ErrorIs(t, err, ErrCatalogConflict)
		})
	}
}

func TestBootstrapVideoExpertsRejectsTypedNilDependency(t *testing.T) {
	bootstrapper := NewBootstrapper(
		(*skillsvc.PlatformCatalogService)(nil),
		newBootstrapExpertStore(),
		&bootstrapWorkerPreparer{},
		&bootstrapSnapshotStore{},
		&bootstrapDependencyArtifactStore{},
	)

	require.NotPanics(t, func() {
		_, err := bootstrapper.Run(context.Background(), BootstrapRequest{
			OrganizationID:   7,
			OrganizationSlug: slugkit.MustNewForTest("dev-org"),
			PublisherUserID:  11,
			ReviewerUserID:   13,
			ModelResourceID:  17,
			RuntimeImageID:   19,
		})
		require.EqualError(t, err, "operator catalog dependencies are incomplete")
	})
}

type bootstrapSkillStore struct {
	rows map[string]*skilldom.Skill
}

func (store *bootstrapSkillStore) EnsurePlatformSkill(
	_ context.Context,
	request *skillsvc.EnsurePlatformSkillRequest,
) (*skilldom.Skill, bool, error) {
	if store.rows == nil {
		store.rows = map[string]*skilldom.Skill{}
	}
	if row := store.rows[request.Slug]; row != nil {
		return row, false, nil
	}
	row := &skilldom.Skill{
		ID:          int64(len(store.rows) + 1),
		Slug:        request.Slug,
		DisplayName: request.Name,
		Description: request.Description,
		License:     request.License,
		Tags:        pq.StringArray(request.Tags),
		IsActive:    true,
		ContentSha:  "content",
		StorageKey:  "storage/" + request.Slug,
		PackageSize: 1,
		Version:     1,
	}
	store.rows[request.Slug] = row
	return row, true, nil
}

type bootstrapExpertStore struct {
	experts     map[string]*expertdom.Expert
	published   map[string]*expertsvc.PublishedMarketApplication
	submitted   map[int64]expertsvc.SubmitMarketApplicationRequest
	releases    map[int64]expertmarket.Release
	nextExpert  int64
	nextRelease int64
}

func newBootstrapExpertStore() *bootstrapExpertStore {
	return &bootstrapExpertStore{
		experts:   map[string]*expertdom.Expert{},
		published: map[string]*expertsvc.PublishedMarketApplication{},
		submitted: map[int64]expertsvc.SubmitMarketApplicationRequest{},
		releases:  map[int64]expertmarket.Release{},
	}
}

func (store *bootstrapExpertStore) GetBySlug(
	_ context.Context,
	_ int64,
	slug string,
) (*expertdom.Expert, error) {
	row := store.experts[slug]
	if row == nil {
		return nil, expertdom.ErrNotFound
	}
	return row, nil
}

func (store *bootstrapExpertStore) Create(
	_ context.Context,
	request *expertsvc.CreateExpertRequest,
) (*expertdom.Expert, error) {
	store.nextExpert++
	row := &expertdom.Expert{
		ID:                   store.nextExpert,
		OrganizationID:       request.OrganizationID,
		Slug:                 request.Slug,
		Name:                 request.Name,
		Description:          request.Description,
		AgentSlug:            request.AgentSlug,
		Prompt:               request.Prompt,
		InteractionMode:      request.InteractionMode,
		AutomationLevel:      request.AutomationLevel,
		SkillSlugs:           pq.StringArray(request.SkillSlugs),
		WorkerSpecSnapshotID: request.WorkerSpecSnapshotID,
	}
	row.ConfigOverrides, _ = json.Marshal(request.ConfigOverrides)
	store.experts[row.Slug] = row
	return row, nil
}

func (store *bootstrapExpertStore) RebindWorkerSpecSnapshot(
	_ context.Context,
	_ int64,
	expertID int64,
	snapshotID int64,
) (*expertdom.Expert, error) {
	for _, row := range store.experts {
		if row.ID != expertID {
			continue
		}
		row.WorkerSpecSnapshotID = &snapshotID
		return row, nil
	}
	return nil, expertdom.ErrNotFound
}

func (store *bootstrapExpertStore) SubmitMarketApplication(
	_ context.Context,
	request expertsvc.SubmitMarketApplicationRequest,
) (*expertsvc.MarketSubmission, error) {
	if _, exists := store.published[request.Slug]; exists {
		return nil, expertmarket.ErrConflict
	}
	store.nextRelease++
	release := expertmarket.Release{
		ID:             store.nextRelease,
		SourceExpertID: request.SourceExpertID,
		Status:         expertmarket.ReleaseStatusPendingReview,
	}
	store.submitted[release.ID] = request
	store.releases[release.ID] = release
	return &expertsvc.MarketSubmission{Release: release}, nil
}

func (store *bootstrapExpertStore) GetPublishedMarketApplication(
	_ context.Context,
	slug string,
) (*expertsvc.PublishedMarketApplication, error) {
	row := store.published[slug]
	if row == nil {
		return nil, expertsvc.ErrMarketApplicationNotFound
	}
	return row, nil
}

func (store *bootstrapExpertStore) ListPublisherMarketReleases(
	context.Context,
	int64,
	int,
	int,
) ([]expertmarket.Release, int64, error) {
	rows := make([]expertmarket.Release, 0, len(store.releases))
	for _, release := range store.releases {
		rows = append(rows, release)
	}
	return rows, int64(len(rows)), nil
}

func (store *bootstrapExpertStore) ApproveMarketRelease(
	_ context.Context,
	request expertsvc.ReviewMarketReleaseRequest,
) (*expertmarket.Release, error) {
	submission, exists := store.submitted[request.ReleaseID]
	if !exists {
		return nil, errors.New("submission missing")
	}
	release := store.releases[request.ReleaseID]
	release.Status = expertmarket.ReleaseStatusPublished
	release.Name = store.experts[submission.Slug].Name
	release.Summary = submission.Summary
	release.Description = submission.Description
	release.Category = submission.Category
	release.Icon = submission.Icon
	release.Tags = submission.Tags
	release.Outcomes = submission.Outcomes
	store.releases[release.ID] = release
	store.published[submission.Slug] = &expertsvc.PublishedMarketApplication{
		Application: expertmarket.Application{
			Slug:            slugkit.Slug(submission.Slug),
			IsOperatorOwned: submission.IsOperatorOwned,
		},
		Release: release,
	}
	return &release, nil
}
