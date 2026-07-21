package skill

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldom "github.com/l8ai-cn/agentcloud/backend/internal/domain/skill"
	"github.com/l8ai-cn/agentcloud/backend/internal/service/gitops"
)

// --- in-memory skill catalog repository ---

type fakeStore struct {
	rows      map[int64]*skilldom.Skill
	nextID    int64
	createErr error
	updateErr error
}

func newFakeStore() *fakeStore { return &fakeStore{rows: map[int64]*skilldom.Skill{}} }

func orgMatches(rowOrg *int64, orgID int64) bool {
	return rowOrg != nil && *rowOrg == orgID
}

func (f *fakeStore) Create(_ context.Context, s *skilldom.Skill) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.nextID++
	s.ID = f.nextID
	cp := *s
	f.rows[s.ID] = &cp
	return nil
}

func (f *fakeStore) Update(_ context.Context, s *skilldom.Skill) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if _, ok := f.rows[s.ID]; !ok {
		return skilldom.ErrNotFound
	}
	cp := *s
	f.rows[s.ID] = &cp
	return nil
}

func (f *fakeStore) UpdateIfVersion(_ context.Context, s *skilldom.Skill, expectedVersion int) (bool, error) {
	if f.updateErr != nil {
		return false, f.updateErr
	}
	current, ok := f.rows[s.ID]
	if !ok {
		return false, skilldom.ErrNotFound
	}
	if current.Version != expectedVersion {
		return false, nil
	}
	cp := *s
	f.rows[s.ID] = &cp
	return true, nil
}

func (f *fakeStore) Delete(_ context.Context, orgID, id int64) error {
	s, ok := f.rows[id]
	if !ok || !orgMatches(s.OrganizationID, orgID) {
		return skilldom.ErrNotFound
	}
	delete(f.rows, id)
	return nil
}

func (f *fakeStore) GetByID(_ context.Context, orgID, id int64) (*skilldom.Skill, error) {
	s, ok := f.rows[id]
	if !ok || !orgMatches(s.OrganizationID, orgID) {
		return nil, skilldom.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (f *fakeStore) GetAnyByID(_ context.Context, id int64) (*skilldom.Skill, error) {
	s, ok := f.rows[id]
	if !ok {
		return nil, skilldom.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (f *fakeStore) GetBySlug(_ context.Context, orgID int64, slug string) (*skilldom.Skill, error) {
	for _, s := range f.rows {
		if orgMatches(s.OrganizationID, orgID) && s.Slug == slug {
			cp := *s
			return &cp, nil
		}
	}
	return nil, skilldom.ErrNotFound
}

func (f *fakeStore) GetPlatformBySlug(
	_ context.Context,
	slug string,
) (*skilldom.Skill, error) {
	for _, skill := range f.rows {
		if skill.OrganizationID == nil && skill.Slug == slug {
			copy := *skill
			return &copy, nil
		}
	}
	return nil, skilldom.ErrNotFound
}

func (f *fakeStore) FindByUpstream(_ context.Context, orgID int64, upstreamURL, upstreamSubdir string) (*skilldom.Skill, error) {
	for _, s := range f.rows {
		if orgMatches(s.OrganizationID, orgID) && s.UpstreamURL == upstreamURL && s.UpstreamSubdir == upstreamSubdir {
			cp := *s
			return &cp, nil
		}
	}
	return nil, skilldom.ErrNotFound
}

func (f *fakeStore) SlugExists(_ context.Context, orgID int64, slug string, excludeID int64) (bool, error) {
	for _, s := range f.rows {
		if orgMatches(s.OrganizationID, orgID) && s.Slug == slug && s.ID != excludeID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeStore) List(_ context.Context, orgID int64, limit, offset int) ([]skilldom.Skill, int64, error) {
	var out []skilldom.Skill
	for _, s := range f.rows {
		if orgMatches(s.OrganizationID, orgID) {
			out = append(out, *s)
		}
	}
	return out, int64(len(out)), nil
}

func (f *fakeStore) ListAll(_ context.Context, orgID int64) ([]skilldom.Skill, error) {
	var out []skilldom.Skill
	for _, s := range f.rows {
		if orgMatches(s.OrganizationID, orgID) {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (f *fakeStore) ListActivePlatformBySlugs(
	_ context.Context,
	slugs []string,
) ([]skilldom.Skill, error) {
	required := make(map[string]struct{}, len(slugs))
	for _, slug := range slugs {
		required[slug] = struct{}{}
	}
	var out []skilldom.Skill
	for _, skill := range f.rows {
		if _, ok := required[skill.Slug]; ok &&
			skill.OrganizationID == nil && skill.IsActive {
			out = append(out, *skill)
		}
	}
	return out, nil
}

func (f *fakeStore) ListCatalog(_ context.Context, orgID int64, query, category string) ([]skilldom.Skill, error) {
	var out []skilldom.Skill
	for _, s := range f.rows {
		if s.IsActive && (s.OrganizationID == nil || orgMatches(s.OrganizationID, orgID)) {
			out = append(out, *s)
		}
	}
	return out, nil
}

func newTestService(store skilldom.Repository, g gitops.Service, pkg SkillPackagerBridge) *Service {
	return NewService(Deps{Store: store, Gitops: g, Packager: pkg})
}

// --- tests ---

func TestNewService_NilDepsDisabled(t *testing.T) {
	assert.Nil(t, NewService(Deps{Gitops: nil, Packager: &fakePackager{}, Store: newFakeStore()}))
	assert.Nil(t, NewService(Deps{Gitops: gitops.NewFake("am-skills"), Packager: nil, Store: newFakeStore()}))
	assert.Nil(t, NewService(Deps{Gitops: gitops.NewFake("am-skills"), Packager: &fakePackager{}, Store: nil}))
	assert.NotNil(t, NewService(Deps{Gitops: gitops.NewFake("am-skills"), Packager: &fakePackager{}, Store: newFakeStore()}))
}

func TestCreate_ProvisionsSeedsAndPackages(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	pkg := &fakePackager{}
	svc := newTestService(store, fake, pkg)

	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7, UserID: 3,
		Name:         "Web Search",
		Description:  "Search the web",
		License:      "MIT",
		Instructions: "# Web Search\n\nUse this to search.",
	})
	require.NoError(t, err)

	assert.Equal(t, "web-search", row.Slug)
	assert.Equal(t, "am-skills/org7-web-search", row.GitRepoPath)
	assert.Equal(t, "main", row.DefaultBranch)
	assert.Equal(t, skilldom.SourceGitops, row.InstallSource)
	assert.NotEmpty(t, row.ContentSha)
	assert.Contains(t, row.StorageKey, "skills/catalog/")
	assert.Positive(t, row.PackageSize)
	assert.Equal(t, 1, row.Version)
	require.NotNil(t, row.HTTPCloneURL)

	// Repo seeded with SKILL.md + skill.json.
	repo, ok := fake.Repos["org7-web-search"]
	require.True(t, ok)
	assert.Contains(t, string(repo.Files["SKILL.md"]), "name: web-search")
	assert.Contains(t, string(repo.Files["SKILL.md"]), "# Web Search")
	assert.Contains(t, string(repo.Files["skill.json"]), `"slug": "web-search"`)

	// Packager saw the materialized files and derived the matching slug.
	assert.Equal(t, 1, pkg.calls)
	assert.Equal(t, []string{"am-skills/org7-web-search"}, pkg.catalogIdentities)
	assert.Equal(t, "web-search", frontmatterName(pkg.lastSkillMd))
	assert.Contains(t, pkg.lastSkillCfg, `"name": "Web Search"`)
}

func TestCreate_DBFailureDeletesRepo(t *testing.T) {
	store := newFakeStore()
	store.createErr = errors.New("db down")
	fake := gitops.NewFake("am-skills")
	svc := newTestService(store, fake, &fakePackager{})

	_, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7, Name: "Web Search", Instructions: "body",
	})
	require.Error(t, err)
	_, ok := fake.Repos["org7-web-search"]
	assert.False(t, ok, "repo must be compensating-deleted on DB failure")
}

func TestCreate_PackageFailureDeletesRepo(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	pkg := &fakePackager{failErr: errors.New("package boom")}
	svc := newTestService(store, fake, pkg)

	_, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7, Name: "Web Search", Instructions: "body",
	})
	require.Error(t, err)
	_, ok := fake.Repos["org7-web-search"]
	assert.False(t, ok, "repo must be compensating-deleted on packaging failure")
	assert.Empty(t, store.rows)
}

func TestCreate_ValidationErrors(t *testing.T) {
	svc := newTestService(newFakeStore(), gitops.NewFake("am-skills"), &fakePackager{})
	_, err := svc.Create(context.Background(), &CreateSkillRequest{OrganizationID: 7, Instructions: "b"})
	assert.ErrorIs(t, err, ErrNameRequired)
	_, err = svc.Create(context.Background(), &CreateSkillRequest{OrganizationID: 7, Name: "X"})
	assert.ErrorIs(t, err, ErrInstructionsRequired)
}

func TestUpdate_CommitsAndRepackages(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	pkg := &fakePackager{}
	svc := newTestService(store, fake, pkg)

	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7, Name: "Web Search", Instructions: "# v1 body",
	})
	require.NoError(t, err)
	shaV1 := row.ContentSha

	newBody := "# v2 body\n\nMore detail."
	updated, err := svc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7, SkillID: row.ID,
		Instructions: &newBody,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)
	assert.NotEqual(t, shaV1, updated.ContentSha)
	assert.Contains(t, updated.StorageKey, "skills/catalog/")
	assert.Equal(t, []string{
		"am-skills/org7-web-search",
		"am-skills/org7-web-search",
	}, pkg.catalogIdentities)

	repo := fake.Repos["org7-web-search"]
	assert.Contains(t, string(repo.Files["SKILL.md"]), "# v2 body")
}

func TestUpdate_PreservesBodyWhenInstructionsNil(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	svc := newTestService(store, fake, &fakePackager{})

	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7, Name: "Web Search", Description: "old", Instructions: "# preserved body",
	})
	require.NoError(t, err)

	newDesc := "new description"
	updated, err := svc.Update(context.Background(), &UpdateSkillRequest{
		OrganizationID: 7, SkillID: row.ID, Description: &newDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, "new description", updated.Description)

	repo := fake.Repos["org7-web-search"]
	assert.Contains(t, string(repo.Files["SKILL.md"]), "# preserved body")
	assert.Contains(t, string(repo.Files["SKILL.md"]), "description: new description")
}

func TestDelete_RemovesRowAndRepo(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	svc := newTestService(store, fake, &fakePackager{})

	row, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7, Name: "Web Search", Instructions: "body",
	})
	require.NoError(t, err)

	require.NoError(t, svc.Delete(context.Background(), 7, row.ID))
	_, err = svc.Get(context.Background(), 7, "web-search")
	assert.ErrorIs(t, err, skilldom.ErrNotFound)
	_, ok := fake.Repos["org7-web-search"]
	assert.False(t, ok)
}

func TestListAndGet_ServedFromDBCache(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	svc := newTestService(store, fake, &fakePackager{})

	_, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7, Name: "Alpha", Instructions: "a",
	})
	require.NoError(t, err)

	items, total, err := svc.List(context.Background(), 7, 50, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, items, 1)

	got, err := svc.Get(context.Background(), 7, "alpha")
	require.NoError(t, err)
	assert.Equal(t, "Alpha", got.DisplayName)
}

func TestReadSkillFile_And_Tree(t *testing.T) {
	store := newFakeStore()
	fake := gitops.NewFake("am-skills")
	svc := newTestService(store, fake, &fakePackager{})

	_, err := svc.Create(context.Background(), &CreateSkillRequest{
		OrganizationID: 7, Name: "Web Search", Instructions: "# body",
	})
	require.NoError(t, err)

	data, entry, err := svc.ReadSkillFile(context.Background(), 7, "web-search", "SKILL.md")
	require.NoError(t, err)
	assert.Equal(t, "SKILL.md", entry.Path)
	assert.Contains(t, string(data), "name: web-search")

	entries, err := svc.ListSkillTree(context.Background(), 7, "web-search")
	require.NoError(t, err)
	paths := map[string]bool{}
	for _, e := range entries {
		paths[e.Path] = true
	}
	assert.True(t, paths["SKILL.md"])
	assert.True(t, paths["skill.json"])
}
