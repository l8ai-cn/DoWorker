package knowledgebase

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

	"github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/infra/gitea"
	"github.com/anthropics/agentsmesh/backend/internal/service/knowledgebase/connector"
)

type fakeConnector struct {
	docs []connector.Doc
}

func (f *fakeConnector) SourceType() string { return "feishu" }

func (f *fakeConnector) ListDocs(_ context.Context, _ json.RawMessage) ([]connector.DocRef, error) {
	refs := make([]connector.DocRef, len(f.docs))
	for i, d := range f.docs {
		refs[i] = d.Ref
	}
	return refs, nil
}

func (f *fakeConnector) FetchDoc(_ context.Context, _ json.RawMessage, ref connector.DocRef) (*connector.Doc, error) {
	for _, d := range f.docs {
		if d.Ref.ID == ref.ID {
			doc := d
			return &doc, nil
		}
	}
	return nil, connector.ErrBadConfig
}

type fakeSyncRepo struct {
	knowledgebase.Repository
	updates []map[string]any
}

func (r *fakeSyncRepo) Update(_ context.Context, _, _ int64, updates map[string]any) error {
	r.updates = append(r.updates, updates)
	return nil
}

type commitRequest struct {
	Branch  string `json:"branch"`
	Message string `json:"message"`
	Files   []struct {
		Operation string `json:"operation"`
		Path      string `json:"path"`
		Content   string `json:"content"`
		SHA       string `json:"sha"`
	} `json:"files"`
}

func TestSyncFromConnector_CommitsChangedDocsOnly(t *testing.T) {
	unchanged := connector.Doc{
		Ref:      connector.DocRef{ID: "d1", Title: "Stable", Path: "stable-d1.md"},
		Markdown: "same content",
	}
	updated := connector.Doc{
		Ref:      connector.DocRef{ID: "d2", Title: "Changed", Path: "changed-d2.md"},
		Markdown: "new content",
	}
	fresh := connector.Doc{
		Ref:      connector.DocRef{ID: "d3", Title: "New", Path: "new-d3.md"},
		Markdown: "brand new",
	}

	var commits []commitRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/git/trees/"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tree": []map[string]any{
					{"path": "raw/feishu/stable-d1.md", "type": "blob", "sha": gitBlobSHA("same content")},
					{"path": "raw/feishu/changed-d2.md", "type": "blob", "sha": gitBlobSHA("old content")},
				},
			})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/contents"):
			var req commitRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			commits = append(commits, req)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{}"))
		default:
			t.Errorf("unexpected gitea call: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	repo := &fakeSyncRepo{}
	svc := NewService(repo, gitea.NewClient(gitea.Config{BaseURL: srv.URL, AdminToken: "t"}), slog.Default())
	kb := &knowledgebase.KnowledgeBase{
		ID: 1, OrganizationID: 7, Slug: "team-docs",
		GitRepoPath: "am-kb/org-7-team-docs", DefaultBranch: "main",
		SourceType: "feishu", SourceConfig: json.RawMessage(`{}`),
	}

	conn := &fakeConnector{docs: []connector.Doc{unchanged, updated, fresh}}
	require.NoError(t, svc.SyncFromConnector(context.Background(), kb, conn))

	require.Len(t, commits, 1)
	files := commits[0].Files
	require.Len(t, files, 2, "unchanged doc must be skipped")

	byPath := map[string]struct {
		Operation string
		Content   string
		SHA       string
	}{}
	for _, f := range files {
		decoded, err := base64.StdEncoding.DecodeString(f.Content)
		require.NoError(t, err)
		byPath[f.Path] = struct {
			Operation string
			Content   string
			SHA       string
		}{f.Operation, string(decoded), f.SHA}
	}
	assert.Equal(t, "update", byPath["raw/feishu/changed-d2.md"].Operation)
	assert.Equal(t, gitBlobSHA("old content"), byPath["raw/feishu/changed-d2.md"].SHA)
	assert.Equal(t, "new content", byPath["raw/feishu/changed-d2.md"].Content)
	assert.Equal(t, "create", byPath["raw/feishu/new-d3.md"].Operation)

	require.Len(t, repo.updates, 2)
	assert.Equal(t, knowledgebase.SyncStatusSyncing, repo.updates[0]["sync_status"])
	assert.Equal(t, knowledgebase.SyncStatusSynced, repo.updates[1]["sync_status"])
	assert.NotNil(t, repo.updates[1]["last_synced_at"])
}

func TestSyncFromConnector_NoChangesSkipsCommit(t *testing.T) {
	doc := connector.Doc{
		Ref:      connector.DocRef{ID: "d1", Title: "Stable", Path: "stable-d1.md"},
		Markdown: "same content",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			t.Errorf("no commit expected, got %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tree": []map[string]any{
				{"path": "raw/feishu/stable-d1.md", "type": "blob", "sha": gitBlobSHA("same content")},
			},
		})
	}))
	defer srv.Close()

	repo := &fakeSyncRepo{}
	svc := NewService(repo, gitea.NewClient(gitea.Config{BaseURL: srv.URL, AdminToken: "t"}), slog.Default())
	kb := &knowledgebase.KnowledgeBase{
		ID: 1, OrganizationID: 7, Slug: "team-docs",
		GitRepoPath: "am-kb/org-7-team-docs", DefaultBranch: "main",
		SourceType: "feishu", SourceConfig: json.RawMessage(`{}`),
	}
	require.NoError(t, svc.SyncFromConnector(context.Background(), kb, &fakeConnector{docs: []connector.Doc{doc}}))
	assert.Equal(t, knowledgebase.SyncStatusSynced, repo.updates[len(repo.updates)-1]["sync_status"])
}

func TestGitBlobSHA_MatchesGitObjectID(t *testing.T) {
	// $ printf 'hello\n' | git hash-object --stdin
	assert.Equal(t, "ce013625030ba8dba906f756967f9e9ca394464a", gitBlobSHA("hello\n"))
}
