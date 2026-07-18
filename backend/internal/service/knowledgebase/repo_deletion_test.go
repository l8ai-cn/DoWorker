package knowledgebase

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	kbdomain "github.com/anthropics/agentsmesh/backend/internal/domain/knowledgebase"
	"github.com/anthropics/agentsmesh/backend/internal/infra/gitea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteRevokesDeployKeysAndRepoBeforeDatabase(t *testing.T) {
	var order []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/am-kb/org1-docs/keys":
			order = append(order, "list-keys")
			_ = json.NewEncoder(w).Encode([]gitea.DeployKey{{ID: 7}, {ID: 8}})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/repos/am-kb/org1-docs/keys/7":
			order = append(order, "delete-key-7")
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/repos/am-kb/org1-docs/keys/8":
			order = append(order, "delete-key-8")
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/repos/am-kb/org1-docs":
			order = append(order, "delete-repo")
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()
	repo := &deleteTestRepo{onDelete: func() { order = append(order, "delete-db") }}
	service := deleteTestService(server.URL, repo)

	require.NoError(t, service.Delete(context.Background(), 1, 2))
	assert.Equal(t, []string{
		"list-keys", "delete-key-7", "delete-key-8", "delete-repo", "delete-db",
	}, order)
}

func TestDeleteKeepsDatabaseRowWhenGiteaFails(t *testing.T) {
	for _, failurePath := range []string{"list", "key", "repo"} {
		t.Run(failurePath, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method == http.MethodGet:
					if failurePath == "list" {
						http.Error(w, "failed", http.StatusInternalServerError)
						return
					}
					_ = json.NewEncoder(w).Encode([]gitea.DeployKey{{ID: 7}})
				case r.URL.Path == "/api/v1/repos/am-kb/org1-docs/keys/7":
					if failurePath == "key" {
						http.Error(w, "failed", http.StatusInternalServerError)
					}
				case r.URL.Path == "/api/v1/repos/am-kb/org1-docs":
					if failurePath == "repo" {
						http.Error(w, "failed", http.StatusInternalServerError)
					}
				}
			}))
			defer server.Close()
			repo := &deleteTestRepo{}

			err := deleteTestService(server.URL, repo).Delete(context.Background(), 1, 2)

			require.Error(t, err)
			assert.Zero(t, repo.deleteCalls)
		})
	}
}

func TestDeleteTreatsMissingGiteaRepoAsAlreadyRevoked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()
	repo := &deleteTestRepo{}

	require.NoError(t, deleteTestService(server.URL, repo).Delete(context.Background(), 1, 2))
	assert.Equal(t, 1, repo.deleteCalls)
}

func TestCreateCleanupReportsRepositoryDeletionFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "delete failed", http.StatusInternalServerError)
	}))
	defer server.Close()
	cause := errors.New("database insert failed")

	err := deleteTestService(server.URL, &deleteTestRepo{}).
		failCreateAndCleanupRepo(context.Background(), "org1-docs", cause)

	require.Error(t, err)
	assert.ErrorIs(t, err, cause)
	assert.Contains(t, err.Error(), "cleanup repository")
}

func TestDeleteDisablesRecordWhenDatabaseDeleteFailsAndRetryCompletes(t *testing.T) {
	var remoteCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteCalls++
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode([]gitea.DeployKey{})
		}
	}))
	defer server.Close()
	repo := &deleteTestRepo{deleteErr: errors.New("database unavailable")}
	service := deleteTestService(server.URL, repo)

	err := service.Delete(context.Background(), 1, 2)

	require.Error(t, err)
	require.NotNil(t, repo.kb)
	assert.Empty(t, repo.kb.GitRepoPath)
	assert.Empty(t, repo.kb.HTTPCloneURL)
	assert.Equal(t, kbdomain.SyncStatusFailed, repo.kb.SyncStatus)
	assert.Equal(t, 1, repo.updateCalls)
	repo.kb.Slug = "docs"
	_, mountErr := service.ResolveMountsForPod(
		context.Background(),
		1,
		"",
		[]MountRequest{{KBSlug: "docs"}},
	)
	assert.ErrorIs(t, mountErr, ErrNotConfigured)
	firstRemoteCalls := remoteCalls

	repo.deleteErr = nil
	require.NoError(t, service.Delete(context.Background(), 1, 2))
	assert.Equal(t, firstRemoteCalls, remoteCalls)
	assert.Equal(t, 2, repo.deleteCalls)
}

func deleteTestService(baseURL string, repo *deleteTestRepo) *Service {
	return NewService(repo, gitea.NewClient(gitea.Config{
		BaseURL: baseURL, AdminToken: "token", Namespace: "am-kb",
	}), slog.Default())
}

type deleteTestRepo struct {
	deleteCalls int
	updateCalls int
	onDelete    func()
	deleteErr   error
	kb          *kbdomain.KnowledgeBase
}

func (repo *deleteTestRepo) Get(context.Context, int64, int64) (*kbdomain.KnowledgeBase, error) {
	if repo.kb == nil {
		repo.kb = &kbdomain.KnowledgeBase{
			GitRepoPath:  "am-kb/org1-docs",
			HTTPCloneURL: "http://gitea/am-kb/org1-docs.git",
		}
	}
	return repo.kb, nil
}
func (repo *deleteTestRepo) Delete(context.Context, int64, int64) error {
	repo.deleteCalls++
	if repo.onDelete != nil {
		repo.onDelete()
	}
	return repo.deleteErr
}
func (*deleteTestRepo) Create(context.Context, *kbdomain.KnowledgeBase) error {
	return errors.New("unused")
}
func (*deleteTestRepo) GetBySlug(context.Context, int64, string) (*kbdomain.KnowledgeBase, error) {
	return nil, errors.New("unused")
}
func (*deleteTestRepo) List(context.Context, *kbdomain.ListFilter) ([]*kbdomain.KnowledgeBase, error) {
	return nil, errors.New("unused")
}
func (*deleteTestRepo) ListExternal(context.Context) ([]*kbdomain.KnowledgeBase, error) {
	return nil, errors.New("unused")
}
func (repo *deleteTestRepo) ListBySlugs(context.Context, int64, []string) ([]*kbdomain.KnowledgeBase, error) {
	return []*kbdomain.KnowledgeBase{repo.kb}, nil
}
func (repo *deleteTestRepo) Update(_ context.Context, _, _ int64, updates map[string]any) error {
	repo.updateCalls++
	repo.kb.GitRepoPath, _ = updates["git_repo_path"].(string)
	repo.kb.HTTPCloneURL, _ = updates["http_clone_url"].(string)
	repo.kb.SyncStatus, _ = updates["sync_status"].(string)
	return nil
}
func (*deleteTestRepo) SlugExists(context.Context, int64, string) (bool, error) {
	return false, errors.New("unused")
}
func (*deleteTestRepo) ReplaceAgentMounts(context.Context, int64, int64, []*kbdomain.AgentMount) error {
	return errors.New("unused")
}
func (*deleteTestRepo) ListAgentMounts(context.Context, int64, int64) ([]*kbdomain.AgentMount, error) {
	return nil, errors.New("unused")
}
func (*deleteTestRepo) ListMountsForAgent(context.Context, int64, string) ([]*kbdomain.AgentMount, error) {
	return nil, errors.New("unused")
}
