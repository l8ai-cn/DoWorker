package gitea

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePublicBranchCommit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/repos/dev-org/demo":
			_, _ = w.Write([]byte(`{"private":false}`))
		case "/api/v1/repos/dev-org/demo/branches/main":
			_, _ = w.Write([]byte(`{"commit":{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, AdminToken: "token"})
	commit, err := client.ResolvePublicBranchCommit(context.Background(), "dev-org/demo", "main")

	require.NoError(t, err)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", commit)
}

func TestResolvePublicBranchCommitRejectsPrivateRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/repos/dev-org/private", r.URL.Path)
		_, _ = w.Write([]byte(`{"private":true}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, AdminToken: "token"})
	_, err := client.ResolvePublicBranchCommit(context.Background(), "dev-org/private", "main")

	require.ErrorContains(t, err, "is not public")
}

func TestResolveBranchCommitEscapesSlashInBranchName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/repos/dev-org/demo/branches/release%2F2026", r.URL.EscapedPath())
		_, _ = w.Write([]byte(`{"commit":{"id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}`))
	}))
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL, AdminToken: "token"})
	commit, err := client.ResolveBranchCommit(context.Background(), "dev-org/demo", "release/2026")

	require.NoError(t, err)
	assert.Equal(t, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", commit)
}
