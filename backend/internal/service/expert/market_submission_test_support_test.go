package expert

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	expertdom "github.com/l8ai-cn/agentcloud/backend/internal/domain/expert"
	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
)

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

func marketSkill(slug string, organizationID *int64, active bool) skilldom.Skill {
	id := map[string]int64{
		"remotion-best-practices": 11,
		"video-delivery-qa":       12,
		"worktree":                13,
		"e2e":                     14,
		"gh-merge":                15,
		"merge":                   16,
		"delivery-worktree":       17,
		"delivery-e2e":            18,
		"delivery-github-merge":   19,
		"delivery-gitlab-merge":   20,
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
