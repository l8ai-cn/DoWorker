package gitops

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anthropics/agentsmesh/backend/internal/infra/gitea"
)

func TestRestoreTreeUpdatesOriginalFilesAndDeletesNewFiles(t *testing.T) {
	var gotCommit commitBody
	svc, done := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/git/trees/"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tree": []gitea.TreeEntry{
					{Path: "SKILL.md", Type: "blob", SHA: "skill-sha"},
					{Path: "new.txt", Type: "blob", SHA: "new-sha"},
				},
			})
		case strings.Contains(r.URL.Path, "/contents/SKILL.md"):
			_ = json.NewEncoder(w).Encode(gitea.ContentEntry{
				Path: "SKILL.md",
				Type: "file",
				SHA:  "skill-sha",
				Content: base64.StdEncoding.EncodeToString(
					[]byte("changed"),
				),
			})
		case strings.Contains(r.URL.Path, "/contents/new.txt"):
			_ = json.NewEncoder(w).Encode(gitea.ContentEntry{
				Path:    "new.txt",
				Type:    "file",
				SHA:     "new-sha",
				Content: base64.StdEncoding.EncodeToString([]byte("new")),
			})
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

	err := RestoreTree(context.Background(), svc, "org7-video-editing", "main", &TreeSnapshot{
		Files: map[string][]byte{"SKILL.md": []byte("original")},
	})
	require.NoError(t, err)
	require.Len(t, gotCommit.Files, 2)
	operations := make(map[string]struct{ operation, sha string }, 2)
	for _, file := range gotCommit.Files {
		operations[file.Path] = struct{ operation, sha string }{file.Operation, file.SHA}
	}
	assert.Equal(t, "update", operations["SKILL.md"].operation)
	assert.Equal(t, "skill-sha", operations["SKILL.md"].sha)
	assert.Equal(t, "delete", operations["new.txt"].operation)
	assert.Equal(t, "new-sha", operations["new.txt"].sha)
}
