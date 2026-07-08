package expert

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	agentpoddom "github.com/anthropics/agentsmesh/backend/internal/domain/agentpod"
	expertdom "github.com/anthropics/agentsmesh/backend/internal/domain/expert"
	agentpodSvc "github.com/anthropics/agentsmesh/backend/internal/service/agentpod"
	"github.com/anthropics/agentsmesh/backend/internal/service/gitops"
)

// --- in-memory expert repository ---

type fakeStore struct {
	rows      map[int64]*expertdom.Expert
	nextID    int64
	createErr error
	updateErr error
}

func newFakeStore() *fakeStore { return &fakeStore{rows: map[int64]*expertdom.Expert{}} }

func (f *fakeStore) Create(_ context.Context, e *expertdom.Expert) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.nextID++
	e.ID = f.nextID
	cp := *e
	f.rows[e.ID] = &cp
	return nil
}

func (f *fakeStore) Update(_ context.Context, e *expertdom.Expert) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if _, ok := f.rows[e.ID]; !ok {
		return expertdom.ErrNotFound
	}
	cp := *e
	f.rows[e.ID] = &cp
	return nil
}

func (f *fakeStore) Delete(_ context.Context, orgID, id int64) error {
	e, ok := f.rows[id]
	if !ok || e.OrganizationID != orgID {
		return expertdom.ErrNotFound
	}
	delete(f.rows, id)
	return nil
}

func (f *fakeStore) GetByID(_ context.Context, orgID, id int64) (*expertdom.Expert, error) {
	e, ok := f.rows[id]
	if !ok || e.OrganizationID != orgID {
		return nil, expertdom.ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (f *fakeStore) GetBySlug(_ context.Context, orgID int64, slug string) (*expertdom.Expert, error) {
	for _, e := range f.rows {
		if e.OrganizationID == orgID && e.Slug == slug {
			cp := *e
			return &cp, nil
		}
	}
	return nil, expertdom.ErrNotFound
}

func (f *fakeStore) SlugExists(_ context.Context, orgID int64, slug string, excludeID int64) (bool, error) {
	for _, e := range f.rows {
		if e.OrganizationID == orgID && e.Slug == slug && e.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeStore) List(_ context.Context, orgID int64, limit, offset int) ([]expertdom.Expert, int64, error) {
	var out []expertdom.Expert
	for _, e := range f.rows {
		if e.OrganizationID == orgID {
			out = append(out, *e)
		}
	}
	return out, int64(len(out)), nil
}

func (f *fakeStore) RecordRun(_ context.Context, orgID, id int64, at time.Time) error {
	e, ok := f.rows[id]
	if !ok || e.OrganizationID != orgID {
		return expertdom.ErrNotFound
	}
	e.RunCount++
	e.LastRunAt = &at
	return nil
}

// --- in-memory dispatcher ---

type fakeDispatcher struct {
	lastReq *agentpodSvc.OrchestrateCreatePodRequest
	err     error
}

func (d *fakeDispatcher) CreatePod(
	_ context.Context, req *agentpodSvc.OrchestrateCreatePodRequest,
) (*agentpodSvc.OrchestrateCreatePodResult, error) {
	d.lastReq = req
	if d.err != nil {
		return nil, d.err
	}
	return &agentpodSvc.OrchestrateCreatePodResult{Pod: &agentpoddom.Pod{}}, nil
}

func newTestService(store expertdom.Repository, g gitops.Service, d PodDispatcher) *Service {
	return NewService(Deps{Store: store, Dispatch: d, Gitops: g})
}

func strptr(s string) *string { return &s }

// --- tests ---

func TestCreate_ProvisionsAndSeedsRepo(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-experts")
	svc := newTestService(store, fake, &fakeDispatcher{})

	pngHeader := []byte("\x89PNG\r\n\x1a\n" + "restpayload")
	row, err := svc.Create(context.Background(), &CreateExpertRequest{
		OrganizationID: 7, UserID: 1, Name: "Data Analyst", AgentSlug: "claude-code",
		Description: strptr("crunches numbers"), SkillSlugs: []string{"web-search"},
		Avatar: &AvatarInput{Data: pngHeader, Ext: "png"}, ExpertType: strptr("analysis"),
	})
	require.NoError(t, err)

	require.NotNil(t, row.GitRepoPath)
	assert.Equal(t, "am-experts/org7-data-analyst", *row.GitRepoPath)
	assert.Equal(t, "main", row.DefaultBranch)
	require.NotNil(t, row.HTTPCloneURL)

	repo, ok := fake.Repos["org7-data-analyst"]
	require.True(t, ok, "repo must be provisioned")
	assert.Contains(t, repo.Files, "agent.md")
	assert.Contains(t, repo.Files, "expert.json")
	assert.Contains(t, repo.Files, "README.md")
	assert.Contains(t, repo.Files, "assets/avatar.png")
	assert.Equal(t, pngHeader, repo.Files["assets/avatar.png"])

	// metadata cache persisted with avatar + type.
	meta := parseExpertMetadata(row.Metadata)
	assert.Equal(t, "assets/avatar.png", meta.Avatar)
	assert.Equal(t, "analysis", meta.ExpertType)

	// expert.json reflects avatar + type.
	var cfg map[string]any
	require.NoError(t, json.Unmarshal(repo.Files["expert.json"], &cfg))
	assert.Equal(t, "assets/avatar.png", cfg["avatar"])
	assert.Equal(t, "analysis", cfg["expertType"])
}

func TestCreate_DBFailureDeletesRepo(t *testing.T) {
	store := newFakeStore()
	store.createErr = errors.New("db down")
	fake := gitops.NewFake("am-experts")
	svc := newTestService(store, fake, &fakeDispatcher{})

	_, err := svc.Create(context.Background(), &CreateExpertRequest{
		OrganizationID: 7, UserID: 1, Name: "Doomed", AgentSlug: "claude-code",
	})
	require.Error(t, err)
	assert.Empty(t, fake.Repos, "compensating cleanup must delete the repo")
}

func TestCreate_NilGitopsDBOnly(t *testing.T) {
	store := newFakeStore()
	svc := newTestService(store, nil, &fakeDispatcher{})

	row, err := svc.Create(context.Background(), &CreateExpertRequest{
		OrganizationID: 7, UserID: 1, Name: "Plain", AgentSlug: "claude-code",
	})
	require.NoError(t, err)
	assert.Nil(t, row.GitRepoPath)
	assert.False(t, svc.GitEnabled())
}

func TestUpdate_CommitsAndRefreshesCache(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-experts")
	svc := newTestService(store, fake, &fakeDispatcher{})

	row, err := svc.Create(context.Background(), &CreateExpertRequest{
		OrganizationID: 7, UserID: 1, Name: "V1", AgentSlug: "claude-code",
	})
	require.NoError(t, err)

	updated, err := svc.Update(context.Background(), &UpdateExpertRequest{
		OrganizationID: 7, ExpertID: row.ID, Name: strptr("V2"),
		AgentfileLayer: strptr("PROMPT \"hello\""), ExpertType: strptr("writer"),
	})
	require.NoError(t, err)
	assert.Equal(t, "V2", updated.Name)

	repo := fake.Repos["org7-v1"]
	require.NotNil(t, repo)
	assert.Equal(t, "PROMPT \"hello\"", string(repo.Files["agent.md"]))
	var cfg map[string]any
	require.NoError(t, json.Unmarshal(repo.Files["expert.json"], &cfg))
	assert.Equal(t, "V2", cfg["name"])
	assert.Equal(t, "writer", cfg["expertType"])
}

func TestUpdate_LazyBackfillLegacyRow(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-experts")
	svc := newTestService(store, fake, &fakeDispatcher{})

	// Seed a legacy row directly (no repo).
	legacy := &expertdom.Expert{
		OrganizationID: 7, Slug: "legacy", Name: "Legacy", AgentSlug: "claude-code",
		InteractionMode: "pty", DefaultBranch: "main", Metadata: json.RawMessage("{}"),
	}
	require.NoError(t, store.Create(context.Background(), legacy))
	require.Nil(t, legacy.GitRepoPath)

	updated, err := svc.Update(context.Background(), &UpdateExpertRequest{
		OrganizationID: 7, ExpertID: legacy.ID, Description: strptr("now backed"),
	})
	require.NoError(t, err)
	require.NotNil(t, updated.GitRepoPath)
	assert.Equal(t, "am-experts/org7-legacy", *updated.GitRepoPath)
	_, ok := fake.Repos["org7-legacy"]
	assert.True(t, ok, "legacy row must be provisioned on update")
}

func TestRun_SourcesAgentMdFromGit(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-experts")
	disp := &fakeDispatcher{}
	svc := newTestService(store, fake, disp)

	row, err := svc.Create(context.Background(), &CreateExpertRequest{
		OrganizationID: 7, UserID: 1, Name: "Runner", AgentSlug: "claude-code",
		AgentfileLayer: strptr("PROMPT \"from-git\""),
	})
	require.NoError(t, err)

	// Simulate the DB cache lagging behind Git.
	require.NoError(t, store.Update(context.Background(), func() *expertdom.Expert {
		e, _ := store.GetByID(context.Background(), 7, row.ID)
		e.AgentfileLayer = strptr("PROMPT \"stale-db\"")
		return e
	}()))

	_, err = svc.Run(context.Background(), &RunExpertRequest{
		OrganizationID: 7, UserID: 1, ExpertSlug: "runner",
	})
	require.NoError(t, err)
	require.NotNil(t, disp.lastReq)
	require.NotNil(t, disp.lastReq.AgentfileLayer)
	assert.Equal(t, "PROMPT \"from-git\"", *disp.lastReq.AgentfileLayer)

	// Cache reconciled from Git.
	refreshed, _ := store.GetByID(context.Background(), 7, row.ID)
	require.NotNil(t, refreshed.AgentfileLayer)
	assert.Equal(t, "PROMPT \"from-git\"", *refreshed.AgentfileLayer)
}

func TestRun_FallsBackToDBWhenGitMisses(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-experts")
	disp := &fakeDispatcher{}
	svc := newTestService(store, fake, disp)

	// Row references a repo that does not exist in the fake -> git read misses.
	legacy := &expertdom.Expert{
		OrganizationID: 7, Slug: "ghost", Name: "Ghost", AgentSlug: "claude-code",
		InteractionMode: "pty", DefaultBranch: "main", Metadata: json.RawMessage("{}"),
		GitRepoPath:    strptr("am-experts/org7-ghost"),
		AgentfileLayer: strptr("PROMPT \"db-fallback\""),
	}
	require.NoError(t, store.Create(context.Background(), legacy))

	_, err := svc.Run(context.Background(), &RunExpertRequest{
		OrganizationID: 7, UserID: 1, ExpertSlug: "ghost",
	})
	require.NoError(t, err)
	require.NotNil(t, disp.lastReq.AgentfileLayer)
	assert.Equal(t, "PROMPT \"db-fallback\"", *disp.lastReq.AgentfileLayer)
}

func TestDelete_RemovesRowAndRepo(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-experts")
	svc := newTestService(store, fake, &fakeDispatcher{})

	row, err := svc.Create(context.Background(), &CreateExpertRequest{
		OrganizationID: 7, UserID: 1, Name: "Temp", AgentSlug: "claude-code",
	})
	require.NoError(t, err)
	require.Contains(t, fake.Repos, "org7-temp")

	require.NoError(t, svc.Delete(context.Background(), 7, row.ID))
	_, err = store.GetByID(context.Background(), 7, row.ID)
	assert.ErrorIs(t, err, expertdom.ErrNotFound)
	assert.NotContains(t, fake.Repos, "org7-temp")
}

func TestMergeMetadataPreservesUnknownKeys(t *testing.T) {
	base := json.RawMessage(`{"custom":"keep","avatar":"old.png"}`)
	out := mergeMetadata(base, strptr("assets/avatar.gif"), strptr("analysis"))
	var m map[string]any
	require.NoError(t, json.Unmarshal(out, &m))
	assert.Equal(t, "keep", m["custom"])
	assert.Equal(t, "assets/avatar.gif", m["avatar"])
	assert.Equal(t, "analysis", m["expertType"])
}
