package expert

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	"github.com/anthropics/agentsmesh/backend/internal/domain/expertmarket"
	skilldom "github.com/anthropics/agentsmesh/backend/internal/domain/skill"
	specdomain "github.com/anthropics/agentsmesh/backend/internal/domain/workerspec"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

func TestMarketSubmissionRejectsUnavailableDependencies(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	orgID := int64(8)
	fixture.skills.rows = []skilldom.Skill{
		marketSkill("valid", nil, true),
		marketSkill("inactive", nil, false),
		marketSkill("org-only", &orgID, true),
		marketSkill("unpackaged", nil, true),
	}
	fixture.skills.rows[3].StorageKey = ""
	fixture.source.SkillSlugs = pq.StringArray{
		"valid",
		"missing",
		"inactive",
		"org-only",
		"unpackaged",
		"missing",
	}
	fixture.snapshots.source.Spec.Workspace.SkillIDs = []int64{
		21, 22, 23, 24, 999,
	}
	require.NoError(t, fixture.store.Update(context.Background(), fixture.source))

	_, err := fixture.service.SubmitMarketApplication(
		context.Background(),
		fixture.submissionRequest(),
	)
	var dependencyErr *MarketDependencyError
	require.ErrorAs(t, err, &dependencyErr)
	require.Equal(
		t,
		[]string{"inactive", "missing", "org-only", "unpackaged"},
		dependencyErr.Missing,
	)
}

func TestMarketSubmissionUsesWorkerSpecDependenciesNotExpertCache(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	fixture.skills.rows = append(
		fixture.skills.rows,
		marketSkill("inactive-runtime-skill", nil, false),
	)
	fixture.snapshots.source.Spec.Workspace.SkillIDs = append(
		fixture.snapshots.source.Spec.Workspace.SkillIDs,
		31,
	)

	_, err := fixture.service.SubmitMarketApplication(
		context.Background(),
		fixture.submissionRequest(),
	)
	var dependencyErr *MarketDependencyError
	require.ErrorAs(t, err, &dependencyErr)
	require.Equal(t, []string{"inactive-runtime-skill"}, dependencyErr.Missing)
}

func TestMarketSubmissionRejectsExpertWorkerSpecDrift(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	fixture.source.AgentSlug = "aider"
	fixture.source.InteractionMode = expertdom.InteractionModePTY
	fixture.source.AutomationLevel = expertdom.AutomationLevelInteractive
	fixture.source.AgentfileLayer = stringPointer(`REPO "private-source"`)
	require.NoError(t, fixture.store.Update(context.Background(), fixture.source))

	_, err := fixture.service.SubmitMarketApplication(
		context.Background(),
		fixture.submissionRequest(),
	)
	require.ErrorIs(t, err, ErrMarketSnapshotInvalid)
	require.ErrorContains(t, err, "agent_slug")
}

func TestMarketSubmissionRejectsUnsupportedIcon(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	request := fixture.submissionRequest()
	request.Icon = "sparkles"

	_, err := fixture.service.SubmitMarketApplication(context.Background(), request)

	require.ErrorContains(t, err, `market icon "sparkles" is unsupported`)
}

func TestMarketSubmissionSnapshotsAndVersionsAreImmutable(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	first, err := fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)
	require.NoError(t, err)
	require.Equal(t, 1, first.Release.Version)

	_, err = fixture.service.RejectMarketRelease(ctx, ReviewMarketReleaseRequest{
		ReviewerUserID:  99,
		ReleaseID:       first.Release.ID,
		RejectionReason: "needs a clearer prompt",
	})
	require.NoError(t, err)
	originalSnapshot := append([]byte(nil), first.Release.ExpertSnapshot...)

	updatedName := "Edited Source"
	fixture.source, err = fixture.service.Update(ctx, &UpdateExpertRequest{
		OrganizationID: fixture.source.OrganizationID,
		ExpertID:       fixture.source.ID,
		Name:           &updatedName,
		Prompt:         stringPointer("new prompt"),
	})
	require.NoError(t, err)
	second, err := fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)
	require.NoError(t, err)
	require.Equal(t, 2, second.Release.Version)

	storedFirst, err := fixture.market.GetReleaseByID(ctx, first.Release.ID)
	require.NoError(t, err)
	require.JSONEq(t, string(originalSnapshot), string(storedFirst.ExpertSnapshot))
	var secondSnapshot marketExpertSnapshot
	require.NoError(t, json.Unmarshal(second.Release.ExpertSnapshot, &secondSnapshot))
	require.Equal(t, updatedName, secondSnapshot.Name)
	require.Equal(t, "new prompt", *secondSnapshot.Prompt)
}

func TestMarketSubmissionRejectsASecondApplicationForTheSameExpert(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	request := fixture.submissionRequest()
	request.Slug = "custom-video-production"
	first, err := fixture.service.SubmitMarketApplication(ctx, request)
	require.NoError(t, err)
	_, err = fixture.service.RejectMarketRelease(ctx, ReviewMarketReleaseRequest{
		ReviewerUserID: 99, ReleaseID: first.Release.ID,
		RejectionReason: "revise",
	})
	require.NoError(t, err)

	request.Slug = fixture.source.Slug
	_, err = fixture.service.SubmitMarketApplication(ctx, request)

	require.ErrorIs(t, err, ErrMarketApplicationSlugMismatch)
	require.Len(t, fixture.market.applications, 1)
}

func TestListPublisherMarketReleasesIncludesApplicationSlug(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	request := fixture.submissionRequest()
	request.Slug = "custom-video-production"
	submission, err := fixture.service.SubmitMarketApplication(ctx, request)
	require.NoError(t, err)

	releases, total, err := fixture.service.ListPublisherMarketReleases(
		ctx,
		fixture.source.OrganizationID,
		20,
		0,
	)

	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, releases, 1)
	require.Equal(t, submission.Application.Slug.String(), releases[0].ApplicationSlug)
}

func TestSourceMarketApplicationScansPastFirstReleasePage(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	for index := range 101 {
		application := expertmarket.Application{
			ID: int64(index + 10), Slug: slugkit.Slug("other-app"),
			PublisherOrganizationID: fixture.source.OrganizationID,
		}
		release := expertmarket.Release{
			ID: int64(index + 10), ApplicationID: application.ID,
			PublisherOrganizationID: fixture.source.OrganizationID,
			SourceExpertID:          int64(index + 1000),
			Version:                 1,
		}
		fixture.market.applications[application.ID] = application
		fixture.market.releases[release.ID] = release
	}
	targetApplication := expertmarket.Application{
		ID: 999, Slug: slugkit.Slug("custom-video-production"),
		PublisherOrganizationID: fixture.source.OrganizationID,
	}
	fixture.market.applications[targetApplication.ID] = targetApplication
	fixture.market.releases[999] = expertmarket.Release{
		ID: 999, ApplicationID: targetApplication.ID,
		PublisherOrganizationID: fixture.source.OrganizationID,
		SourceExpertID:          fixture.source.ID,
		Version:                 2,
	}

	found, err := fixture.service.sourceMarketApplication(
		ctx,
		fixture.source.OrganizationID,
		fixture.source.ID,
	)

	require.NoError(t, err)
	require.Equal(t, targetApplication.ID, found.ID)
}

func TestMarketSubmissionRejectsMalformedSourceSnapshotFields(t *testing.T) {
	tests := map[string]func(*expertdom.Expert){
		"knowledge mounts must be an array": func(source *expertdom.Expert) {
			source.KnowledgeMounts = json.RawMessage(`{"slug":"video-assets"}`)
		},
		"config overrides must be an object": func(source *expertdom.Expert) {
			source.ConfigOverrides = json.RawMessage(`null`)
		},
		"metadata must be an object": func(source *expertdom.Expert) {
			source.Metadata = json.RawMessage(`[]`)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newMarketServiceFixture(t)
			mutate(fixture.source)
			require.NoError(t, fixture.store.Update(
				context.Background(),
				fixture.source,
			))

			_, err := fixture.service.SubmitMarketApplication(
				context.Background(),
				fixture.submissionRequest(),
			)
			require.ErrorIs(t, err, ErrMarketSnapshotInvalid)
		})
	}
}

func TestMarketSubmissionRejectsOrganizationScopedWorkerSpecReferences(t *testing.T) {
	tests := map[string]func(*expertdom.Expert, *fakeMarketSnapshots){
		"repository": func(_ *expertdom.Expert, snapshots *fakeMarketSnapshots) {
			snapshots.source.Spec.Workspace.RepositoryID = int64Pointer(88)
		},
		"knowledge": func(_ *expertdom.Expert, snapshots *fakeMarketSnapshots) {
			snapshots.source.Spec.Workspace.KnowledgeMounts = []specdomain.KnowledgeMount{
				{
					KnowledgeBaseID: 91,
					Mode:            specdomain.KnowledgeMountReadOnly,
				},
			}
		},
		"environment": func(_ *expertdom.Expert, snapshots *fakeMarketSnapshots) {
			snapshots.source.Spec.Workspace.EnvBundleIDs = []specdomain.RuntimeEnvBundleID{92}
		},
		"secret": func(_ *expertdom.Expert, snapshots *fakeMarketSnapshots) {
			snapshots.source.Spec.TypeConfig.SecretRefs = map[string]specdomain.SecretReference{
				"TOKEN": {
					Kind: slugkit.MustNewForTest("env-bundle"),
					ID:   93,
				},
			}
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newMarketServiceFixture(t)
			mutate(fixture.source, fixture.snapshots)

			_, err := fixture.service.SubmitMarketApplication(
				context.Background(),
				fixture.submissionRequest(),
			)
			require.ErrorIs(t, err, ErrMarketSnapshotInvalid)
		})
	}
}

func TestMarketReviewPublishesRejectsAndWithdraws(t *testing.T) {
	fixture := newMarketServiceFixture(t)
	ctx := context.Background()
	submission, err := fixture.service.SubmitMarketApplication(
		ctx,
		fixture.submissionRequest(),
	)
	require.NoError(t, err)

	_, err = fixture.service.RejectMarketRelease(ctx, ReviewMarketReleaseRequest{
		ReviewerUserID: 99,
		ReleaseID:      submission.Release.ID,
	})
	require.ErrorIs(t, err, ErrMarketRejectionReasonRequired)

	published, err := fixture.service.ApproveMarketRelease(
		ctx,
		ReviewMarketReleaseRequest{
			ReviewerUserID: 99,
			ReleaseID:      submission.Release.ID,
		},
	)
	require.NoError(t, err)
	require.Equal(t, expertmarket.ReleaseStatusPublished, published.Status)

	items, err := fixture.service.ListMarketApplications(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, submission.Application.ID, items[0].ID)

	_, err = fixture.service.WithdrawMarketRelease(
		ctx,
		WithdrawMarketReleaseRequest{
			PublisherOrganizationID: fixture.source.OrganizationID,
			ReleaseID:               submission.Release.ID,
		},
	)
	require.NoError(t, err)
	items, err = fixture.service.ListMarketApplications(ctx)
	require.NoError(t, err)
	require.Empty(t, items)
}

type marketServiceFixture struct {
	service   *Service
	store     *fakeStore
	market    *fakeExpertMarket
	skills    *fakeMarketSkills
	snapshots *fakeMarketSnapshots
	locker    *fakeMarketInstallationLocker
	source    *expertdom.Expert
}

func newMarketServiceFixture(t *testing.T) *marketServiceFixture {
	t.Helper()
	store := newFakeStore()
	market := newFakeExpertMarket()
	skills := &fakeMarketSkills{
		rows: []skilldom.Skill{
			marketSkill("remotion-best-practices", nil, true),
			marketSkill("video-delivery-qa", nil, true),
		},
	}
	snapshotID := int64(77)
	snapshots := &fakeMarketSnapshots{
		source: expertWorkerSpecSnapshot(snapshotID, 7),
	}
	locker := &fakeMarketInstallationLocker{}
	snapshots.source.Spec.Workspace.SkillIDs = []int64{11, 12}
	snapshots.source.Spec.Workspace.Instructions = "produce a short video"
	source := &expertdom.Expert{
		OrganizationID:       7,
		Slug:                 "video-production-source",
		Name:                 "Video Production Expert",
		Description:          stringPointer("production"),
		AgentSlug:            "codex-cli",
		Prompt:               stringPointer("produce a short video"),
		InteractionMode:      expertdom.InteractionModeACP,
		AutomationLevel:      expertdom.AutomationLevelAutonomous,
		SkillSlugs:           pq.StringArray{"remotion-best-practices", "video-delivery-qa"},
		KnowledgeMounts:      json.RawMessage("[]"),
		ConfigOverrides:      json.RawMessage("{}"),
		Metadata:             json.RawMessage(`{"expert_type":"video"}`),
		WorkerSpecSnapshotID: &snapshotID,
		CreatedByID:          3,
	}
	require.NoError(t, store.Create(context.Background(), source))
	service := NewService(Deps{
		Store:             store,
		WorkerSpecs:       snapshots,
		WorkerSpecWriter:  snapshots,
		MarketWorkerSpecs: snapshots,
		MarketInstallLock: locker,
		Market:            market,
		MarketSkills:      skills,
	})
	return &marketServiceFixture{
		service:   service,
		store:     store,
		market:    market,
		skills:    skills,
		snapshots: snapshots,
		locker:    locker,
		source:    source,
	}
}

func (fixture *marketServiceFixture) submissionRequest() SubmitMarketApplicationRequest {
	return SubmitMarketApplicationRequest{
		OrganizationID: fixture.source.OrganizationID,
		UserID:         fixture.source.CreatedByID,
		SourceExpertID: fixture.source.ID,
		Slug:           "video-production-expert",
		Summary:        "Creates vertical short videos",
		Description:    "Plans, renders, and validates short videos.",
		Category:       "video",
		Icon:           "clapperboard",
		Tags:           []string{"short-video", "production"},
		Outcomes:       []string{"playable mp4"},
	}
}

func marketSkill(
	slug string,
	organizationID *int64,
	active bool,
) skilldom.Skill {
	id := map[string]int64{
		"remotion-best-practices": 11,
		"video-delivery-qa":       12,
		"valid":                   21,
		"inactive":                22,
		"org-only":                23,
		"unpackaged":              24,
		"inactive-runtime-skill":  31,
	}[slug]
	return skilldom.Skill{
		ID:             id,
		Slug:           slug,
		OrganizationID: organizationID,
		IsActive:       active,
		Version:        2,
		ContentSha:     "sha-" + slug,
		StorageKey:     "skills/" + slug,
	}
}

func stringPointer(value string) *string {
	return &value
}
