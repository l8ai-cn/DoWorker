package gitea

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDeployKeyUsesRepositoryScopeAndRequestedMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/repos/am-kb/org1-docs/keys", r.URL.Path)
		assert.Equal(t, "token admin-token", r.Header.Get("Authorization"))
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "agentcloud-read-only", body["title"])
		assert.Equal(t, "ssh-ed25519 public-key", body["key"])
		assert.Equal(t, true, body["read_only"])
		_ = json.NewEncoder(w).Encode(DeployKey{
			ID: 7, Title: "agentcloud-read-only", ReadOnly: true,
		})
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL: server.URL, AdminToken: "admin-token", Namespace: "am-kb",
	})
	key, err := client.CreateDeployKey(
		context.Background(),
		"org1-docs",
		"agentcloud-read-only",
		"ssh-ed25519 public-key\n",
		true,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(7), key.ID)
	assert.True(t, key.ReadOnly)
}

func TestSSHCloneURLRequiresConfiguredBase(t *testing.T) {
	client := NewClient(Config{Namespace: "am-kb"})
	assert.Empty(t, client.SSHCloneURL("org1-docs"))

	client = NewClient(Config{
		Namespace: "am-kb", SSHCloneBaseURL: "ssh://git@gitea.internal:22/",
	})
	assert.Equal(
		t,
		"ssh://git@gitea.internal:22/am-kb/org1-docs.git",
		client.SSHCloneURL("org1-docs"),
	)
}

func TestSSHKnownHostsReturnsPinnedValue(t *testing.T) {
	client := NewClient(Config{SSHKnownHosts: "gitea ssh-ed25519 host-key"})
	assert.Equal(t, "gitea ssh-ed25519 host-key", client.SSHKnownHosts())
}

func TestDeleteDeployKeyTreatsMissingKeyAsRevoked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/api/v1/repos/am-kb/org1-docs/keys/7", r.URL.Path)
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()
	client := NewClient(Config{
		BaseURL: server.URL, AdminToken: "admin-token", Namespace: "am-kb",
	})

	require.NoError(t, client.DeleteDeployKey(context.Background(), "org1-docs", 7))
}
