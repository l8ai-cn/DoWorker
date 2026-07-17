package expert

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
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
	mutex                sync.Mutex
	rows                 map[int64]*expertdom.Expert
	nextID               int64
	createErr            error
	updateErr            error
	marketLookupMisses   int
	beforeMarketUpdate   func()
	marketUpdateStarted  chan struct{}
	marketUpdateContinue chan struct{}
}

func newFakeStore() *fakeStore { return &fakeStore{rows: map[int64]*expertdom.Expert{}} }

func (f *fakeStore) Create(_ context.Context, e *expertdom.Expert) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.createErr != nil {
		return f.createErr
	}
	for _, existing := range f.rows {
		if e.SourceMarketApplicationID != nil &&
			existing.OrganizationID == e.OrganizationID &&
			existing.SourceMarketApplicationID != nil &&
			*existing.SourceMarketApplicationID == *e.SourceMarketApplicationID {
			return errors.New("duplicate market installation")
		}
	}
	f.nextID++
	e.ID = f.nextID
	if e.Revision == 0 {
		e.Revision = 1
	}
	cp := *e
	f.rows[e.ID] = &cp
	return nil
}

func (f *fakeStore) Update(_ context.Context, e *expertdom.Expert) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.updateErr != nil {
		return f.updateErr
	}
	current, ok := f.rows[e.ID]
	if !ok {
		return expertdom.ErrNotFound
	}
	if current.Revision != e.Revision {
		return expertdom.ErrConflict
	}
	e.Revision++
	cp := *e
	f.rows[e.ID] = &cp
	return nil
}

func (f *fakeStore) Delete(_ context.Context, orgID, id int64) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	e, ok := f.rows[id]
	if !ok || e.OrganizationID != orgID {
		return expertdom.ErrNotFound
	}
	delete(f.rows, id)
	return nil
}

func (f *fakeStore) GetByID(_ context.Context, orgID, id int64) (*expertdom.Expert, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	e, ok := f.rows[id]
	if !ok || e.OrganizationID != orgID {
		return nil, expertdom.ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (f *fakeStore) GetBySlug(_ context.Context, orgID int64, slug string) (*expertdom.Expert, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	for _, e := range f.rows {
		if e.OrganizationID == orgID && e.Slug == slug {
			cp := *e
			return &cp, nil
		}
	}
	return nil, expertdom.ErrNotFound
}

func (f *fakeStore) GetByMarketApplication(
	_ context.Context,
	orgID, applicationID int64,
) (*expertdom.Expert, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.marketLookupMisses > 0 {
		f.marketLookupMisses--
		return nil, expertdom.ErrNotFound
	}
	for _, expert := range f.rows {
		if expert.OrganizationID == orgID &&
			expert.SourceMarketApplicationID != nil &&
			*expert.SourceMarketApplicationID == applicationID {
			copy := *expert
			return &copy, nil
		}
	}
	return nil, expertdom.ErrNotFound
}

func (f *fakeStore) UpdateMarketRelease(
	_ context.Context,
	orgID, expertID, applicationID int64,
	update expertdom.MarketReleaseUpdate,
) error {
	if f.marketUpdateStarted != nil {
		close(f.marketUpdateStarted)
		<-f.marketUpdateContinue
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.beforeMarketUpdate != nil {
		f.beforeMarketUpdate()
	}
	if f.updateErr != nil {
		return f.updateErr
	}
	expert, ok := f.rows[expertID]
	if !ok || expert.OrganizationID != orgID ||
		expert.SourceMarketApplicationID == nil ||
		*expert.SourceMarketApplicationID != applicationID {
		return expertdom.ErrNotFound
	}
	if expert.Revision != update.ExpectedRevision {
		return expertdom.ErrConflict
	}
	applyMarketReleaseUpdate(expert, update)
	expert.Revision++
	return nil
}

func (f *fakeStore) SlugExists(_ context.Context, orgID int64, slug string, excludeID int64) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	for _, e := range f.rows {
		if e.OrganizationID == orgID && e.Slug == slug && e.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeStore) List(_ context.Context, orgID int64, limit, offset int) ([]expertdom.Expert, int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	var out []expertdom.Expert
	for _, e := range f.rows {
		if e.OrganizationID == orgID {
			out = append(out, *e)
		}
	}
	return out, int64(len(out)), nil
}

func (f *fakeStore) RecordRun(_ context.Context, orgID, id int64, at time.Time) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
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
