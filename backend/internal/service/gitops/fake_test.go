package gitops

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFake_ProvisionSeedRoundTrip(t *testing.T) {
	f := NewFake("am-experts")
	ctx := context.Background()

	repo, err := f.Provision(ctx, ProvisionParams{
		OrgID: 7, Slug: "data-analyst",
		Seed: []FileChange{
			{Path: "agent.md", Content: []byte("# Data Analyst")},
			{Path: "assets/avatar.png", Content: []byte{0x89, 0x50, 0x4e, 0x47}},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "org7-data-analyst", repo.Name)
	assert.Equal(t, "am-experts/org7-data-analyst", repo.Path)
	assert.Equal(t, "main", repo.DefaultBranch)
	assert.Equal(t, "https://gitea.local/am-experts/org7-data-analyst.git", repo.HTTPCloneURL)

	content, entry, err := f.ReadFile(ctx, repo.Name, "main", "agent.md")
	require.NoError(t, err)
	assert.Equal(t, "# Data Analyst", string(content))
	assert.Equal(t, "file", entry.Type)
	assert.NotEmpty(t, entry.SHA)

	// Binary content survives round trip.
	raw, _, err := f.ReadFile(ctx, repo.Name, "main", "assets/avatar.png")
	require.NoError(t, err)
	assert.Equal(t, []byte{0x89, 0x50, 0x4e, 0x47}, raw)
}

func TestFake_CommitCreatesAndUpdates(t *testing.T) {
	f := NewFake("am-experts")
	ctx := context.Background()
	repo, err := f.Provision(ctx, ProvisionParams{OrgID: 1, Slug: "x",
		Seed: []FileChange{{Path: "expert.json", Content: []byte(`{"schema":1}`)}}})
	require.NoError(t, err)

	_, firstEntry, err := f.ReadFile(ctx, repo.Name, "main", "expert.json")
	require.NoError(t, err)

	require.NoError(t, f.Commit(ctx, repo.Name, "main", "update", Author{}, []FileChange{
		{Path: "expert.json", Content: []byte(`{"schema":2}`)}, // update
		{Path: "README.md", Content: []byte("hi")},             // create
	}))

	content, updatedEntry, err := f.ReadFile(ctx, repo.Name, "main", "expert.json")
	require.NoError(t, err)
	assert.Equal(t, `{"schema":2}`, string(content))
	assert.NotEqual(t, firstEntry.SHA, updatedEntry.SHA, "pseudo-SHA changes on content change")

	readme, _, err := f.ReadFile(ctx, repo.Name, "main", "README.md")
	require.NoError(t, err)
	assert.Equal(t, "hi", string(readme))
}

func TestFake_ReadMissingFileIsNotFound(t *testing.T) {
	f := NewFake("am-skills")
	ctx := context.Background()
	_, err := f.Provision(ctx, ProvisionParams{OrgID: 1, Slug: "x"})
	require.NoError(t, err)

	_, _, err = f.ReadFile(ctx, "org1-x", "main", "nope.md")
	assert.ErrorIs(t, err, ErrNotFound)

	_, _, err = f.ReadFile(ctx, "missing-repo", "main", "any")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFake_ListDirAndTree(t *testing.T) {
	f := NewFake("am-experts")
	ctx := context.Background()
	_, err := f.Provision(ctx, ProvisionParams{OrgID: 2, Slug: "y", Seed: []FileChange{
		{Path: "agent.md", Content: []byte("a")},
		{Path: "expert.json", Content: []byte("b")},
		{Path: "assets/avatar.png", Content: []byte("c")},
	}})
	require.NoError(t, err)

	root, err := f.ListDir(ctx, "org2-y", "main", "")
	require.NoError(t, err)
	byName := map[string]Entry{}
	for _, e := range root {
		byName[e.Name] = e
	}
	assert.Equal(t, "file", byName["agent.md"].Type)
	assert.Equal(t, "file", byName["expert.json"].Type)
	assert.Equal(t, "dir", byName["assets"].Type)

	assets, err := f.ListDir(ctx, "org2-y", "main", "assets")
	require.NoError(t, err)
	require.Len(t, assets, 1)
	assert.Equal(t, "assets/avatar.png", assets[0].Path)

	tree, err := f.ListTree(ctx, "org2-y", "main")
	require.NoError(t, err)
	paths := map[string]string{}
	for _, e := range tree {
		paths[e.Path] = e.Type
	}
	assert.Equal(t, "file", paths["agent.md"])
	assert.Equal(t, "file", paths["assets/avatar.png"])
	assert.Equal(t, "dir", paths["assets"])
}

func TestFake_DeleteRepo(t *testing.T) {
	f := NewFake("am-experts")
	ctx := context.Background()
	_, err := f.Provision(ctx, ProvisionParams{OrgID: 3, Slug: "z"})
	require.NoError(t, err)
	require.NoError(t, f.DeleteRepo(ctx, "org3-z"))
	_, _, err = f.ReadFile(ctx, "org3-z", "main", "agent.md")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestFake_FailureInjection(t *testing.T) {
	ctx := context.Background()

	f := NewFake("am-experts")
	f.FailProvision = true
	_, err := f.Provision(ctx, ProvisionParams{OrgID: 1, Slug: "x"})
	require.Error(t, err)

	f2 := NewFake("am-experts")
	_, err = f2.Provision(ctx, ProvisionParams{OrgID: 1, Slug: "x"})
	require.NoError(t, err)
	f2.FailCommit = true
	err = f2.Commit(ctx, "org1-x", "main", "m", Author{}, []FileChange{{Path: "a", Content: []byte("b")}})
	require.Error(t, err)
}

func TestFake_NamingAndCloneHelpers(t *testing.T) {
	f := NewFake("am-skills")
	assert.Equal(t, "org9-web-search", f.RepoName(9, "web-search"))
	assert.Equal(t, "am-skills/org9-web-search", f.RepoPath(9, "web-search"))
	assert.Equal(t, "org9-web-search", f.RepoNameFromPath("am-skills/org9-web-search"))
	assert.Equal(t, "bare", f.RepoNameFromPath("bare"))
	assert.Equal(t, "https://gitea.local/am-skills/org9-web-search.git", f.CloneURL("org9-web-search"))

	f.CloneBaseURL = "https://git.example.com/"
	assert.Equal(t, "https://git.example.com/am-skills/org9-web-search.git", f.CloneURL("org9-web-search"))
}

func TestFake_ImplementsServiceAndNilConvention(t *testing.T) {
	// NewFake is a non-nil concrete type; NewService(nil, nil) is the nil path.
	var svc Service = NewFake("am-experts")
	require.NotNil(t, svc)
	assert.Equal(t, "am-experts", svc.Namespace())

	assert.True(t, errors.Is(ErrNotConfigured, ErrNotConfigured))
}
