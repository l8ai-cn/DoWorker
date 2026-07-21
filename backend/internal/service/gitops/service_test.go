package gitops

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/l8ai-cn/agentcloud/backend/internal/infra/gitea"
)

func newTestService(t *testing.T, handler http.HandlerFunc) (Service, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	svc := NewService(
		gitea.NewClient(gitea.Config{BaseURL: srv.URL, AdminToken: "t", Namespace: "am-experts"}),
		slog.Default(),
	)
	require.NotNil(t, svc)
	return svc, srv.Close
}

func TestNewService_NilWhenGitNil(t *testing.T) {
	assert.Nil(t, NewService(nil, nil))
}

type commitBody struct {
	Branch  string `json:"branch"`
	Message string `json:"message"`
	Files   []struct {
		Operation string `json:"operation"`
		Path      string `json:"path"`
		Content   string `json:"content"`
		SHA       string `json:"sha"`
	} `json:"files"`
}

func TestService_ProvisionSeedsAndReturnsRepo(t *testing.T) {
	var gotCommit commitBody
	svc, done := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orgs/am-experts":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/orgs/am-experts/repos"):
			_ = json.NewEncoder(w).Encode(gitea.Repo{Name: "org7-x", DefaultBranch: "main"})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/contents"):
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotCommit))
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{}"))
		default:
			t.Errorf("unexpected call: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	defer done()

	repo, err := svc.Provision(context.Background(), ProvisionParams{
		OrgID: 7, Slug: "x",
		CommitMessage: "init",
		Seed: []FileChange{
			{Path: "agent.md", Content: []byte("hello")},
			{Path: "assets/avatar.png", Content: []byte{0x00, 0x01, 0x02}},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "org7-x", repo.Name)
	assert.Equal(t, "am-experts/org7-x", repo.Path)
	assert.Equal(t, "main", repo.DefaultBranch)
	assert.Contains(t, repo.HTTPCloneURL, "/am-experts/org7-x.git")

	require.Len(t, gotCommit.Files, 2)
	byPath := map[string]string{}
	for _, f := range gotCommit.Files {
		assert.Equal(t, "create", f.Operation)
		decoded, err := base64.StdEncoding.DecodeString(f.Content)
		require.NoError(t, err)
		byPath[f.Path] = string(decoded)
	}
	assert.Equal(t, "hello", byPath["agent.md"])
	assert.Equal(t, "\x00\x01\x02", byPath["assets/avatar.png"], "binary content is lossless")
}

func TestService_ProvisionSeedFailureDeletesRepo(t *testing.T) {
	var deleted bool
	svc, done := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/orgs/am-experts":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/repos"):
			_ = json.NewEncoder(w).Encode(gitea.Repo{Name: "org7-x", DefaultBranch: "main"})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/contents"):
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/repos/am-experts/org7-x":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected call: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	defer done()

	_, err := svc.Provision(context.Background(), ProvisionParams{
		OrgID: 7, Slug: "x",
		Seed: []FileChange{{Path: "agent.md", Content: []byte("hi")}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "seed commit")
	assert.True(t, deleted, "repo must be deleted on seed failure")
}

func TestService_CommitCreateVsUpdate(t *testing.T) {
	var gotCommit commitBody
	svc, done := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/contents/existing.md"):
			_ = json.NewEncoder(w).Encode(gitea.ContentEntry{Path: "existing.md", SHA: "abc123", Type: "file"})
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/contents/new.md"):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/contents"):
			require.NoError(t, json.NewDecoder(r.Body).Decode(&gotCommit))
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{}"))
		default:
			t.Errorf("unexpected call: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	defer done()

	err := svc.Commit(context.Background(), "org7-x", "main", "edit", Author{}, []FileChange{
		{Path: "existing.md", Content: []byte("v2")},
		{Path: "new.md", Content: []byte("brand new")},
	})
	require.NoError(t, err)

	ops := map[string]struct{ op, sha string }{}
	for _, f := range gotCommit.Files {
		ops[f.Path] = struct{ op, sha string }{f.Operation, f.SHA}
	}
	assert.Equal(t, "update", ops["existing.md"].op)
	assert.Equal(t, "abc123", ops["existing.md"].sha)
	assert.Equal(t, "create", ops["new.md"].op)
	assert.Empty(t, ops["new.md"].sha)
}

func TestService_ReadFileMaps404ToErrNotFound(t *testing.T) {
	svc, done := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/contents/present.md"):
			_ = json.NewEncoder(w).Encode(gitea.ContentEntry{
				Name: "present.md", Path: "present.md", Type: "file", Size: 5, SHA: "s1",
				Content: base64.StdEncoding.EncodeToString([]byte("hello")),
			})
		case strings.Contains(r.URL.Path, "/contents/missing.md"):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		default:
			t.Errorf("unexpected call: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	defer done()

	content, entry, err := svc.ReadFile(context.Background(), "org7-x", "main", "present.md")
	require.NoError(t, err)
	assert.Equal(t, "hello", string(content))
	assert.Equal(t, "s1", entry.SHA)

	_, _, err = svc.ReadFile(context.Background(), "org7-x", "main", "missing.md")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestService_ListDirAndTree(t *testing.T) {
	svc, done := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/git/trees/"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tree": []gitea.TreeEntry{
					{Path: "agent.md", Type: "blob", Size: 3, SHA: "b1"},
					{Path: "assets", Type: "tree", SHA: "t1"},
					{Path: "assets/avatar.png", Type: "blob", Size: 9, SHA: "b2"},
				},
			})
		case strings.Contains(r.URL.Path, "/contents"):
			_ = json.NewEncoder(w).Encode([]gitea.ContentEntry{
				{Name: "agent.md", Path: "agent.md", Type: "file", Size: 3, SHA: "b1"},
				{Name: "assets", Path: "assets", Type: "dir", SHA: "t1"},
			})
		default:
			t.Errorf("unexpected call: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	defer done()

	dir, err := svc.ListDir(context.Background(), "org7-x", "main", "")
	require.NoError(t, err)
	require.Len(t, dir, 2)
	assert.Equal(t, "file", dir[0].Type)
	assert.Equal(t, "dir", dir[1].Type)

	tree, err := svc.ListTree(context.Background(), "org7-x", "main")
	require.NoError(t, err)
	byPath := map[string]string{}
	for _, e := range tree {
		byPath[e.Path] = e.Type
	}
	assert.Equal(t, "file", byPath["agent.md"], "blob -> file")
	assert.Equal(t, "dir", byPath["assets"], "tree -> dir")
	assert.Equal(t, "file", byPath["assets/avatar.png"])
}

func TestService_ListTreeMaps404(t *testing.T) {
	svc, done := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	})
	defer done()

	_, err := svc.ListTree(context.Background(), "org7-x", "main")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestService_NamingAndCloneHelpers(t *testing.T) {
	svc, done := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer done()

	assert.Equal(t, "am-experts", svc.Namespace())
	assert.Equal(t, "org5-analyst", svc.RepoName(5, "analyst"))
	assert.Equal(t, "am-experts/org5-analyst", svc.RepoPath(5, "analyst"))
	assert.Equal(t, "org5-analyst", svc.RepoNameFromPath("am-experts/org5-analyst"))
	assert.Contains(t, svc.CloneURL("org5-analyst"), "/am-experts/org5-analyst.git")
}

func TestGiteaHTTPError_FormatCompatibleAndTypeInspectable(t *testing.T) {
	var probed *gitea.HTTPError
	client := gitea.NewClient(gitea.Config{Namespace: "am-experts"})
	_ = client
	// Exercise the typed error via a live 404 from the test server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("nope"))
	}))
	defer srv.Close()
	c := gitea.NewClient(gitea.Config{BaseURL: srv.URL, AdminToken: "t", Namespace: "am-experts"})
	_, err := c.GetFile(context.Background(), "r", "main", "f")
	require.Error(t, err)
	require.ErrorAs(t, err, &probed)
	assert.Equal(t, http.StatusNotFound, probed.StatusCode)
	assert.Contains(t, probed.Error(), "→ 404")
}
